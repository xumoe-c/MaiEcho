package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/tidwall/gjson"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type BilibiliCollector struct {
	storage  storage.Storage
	c        *colly.Collector
	isBanned int32
	cookie   string
}

func NewBilibiliCollector(s storage.Storage, cookie string, proxy string) *BilibiliCollector {
	c := colly.NewCollector(
		colly.Async(true), // 启用异步
	)
	// 允许重复访问相同的 URL (例如多次采集同一首歌的搜索结果)
	c.AllowURLRevisit = true

	// 使用随机 User-Agent
	extensions.RandomUserAgent(c)

	// 如果配置了代理，则设置代理
	if proxy != "" {
		if err := c.SetProxy(proxy); err != nil {
			logger.Error("设置代理失败", "module", "collector.bilibili", "proxy", proxy, "error", err)
		} else {
			logger.Info("已启用代理", "module", "collector.bilibili", "proxy", proxy)
		}
	}

	// 设置并发和延迟以避免被封禁
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*bilibili.com*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second, // Increase delay to be safer
	}); err != nil {
		logger.Error("设置采集限制失败", "module", "collector.bilibili", "error", err)
	}

	bc := &BilibiliCollector{
		storage: s,
		c:       c,
		cookie:  cookie,
	}

	bc.setupCallbacks()
	return bc
}

func (b *BilibiliCollector) Name() string {
	return "bilibili"
}

func (b *BilibiliCollector) Collect(ctx context.Context, keyword string) error {
	// 检查是否处于封禁状态
	if atomic.LoadInt32(&b.isBanned) == 1 {
		return fmt.Errorf("collector is currently banned/rate-limited")
	}

	logger.Info("开始采集", "module", "collector.bilibili", "collector", b.Name(), "keyword", keyword)

	// 采集前 3 页
	for page := 1; page <= 3; page++ {
		// 再次检查封禁状态
		if atomic.LoadInt32(&b.isBanned) == 1 {
			logger.Warn("检测到封禁/错误，停止当前任务", "module", "collector.bilibili", "keyword", keyword)
			break
		}

		// API: https://api.bilibili.com/x/web-interface/search/all/v2?keyword=...&page=...
		apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/all/v2?keyword=%s&page=%d", url.QueryEscape(keyword), page)

		// Create colly context and pass SongID if available
		collyCtx := colly.NewContext()
		collyCtx.Put("keyword", keyword) // Store keyword in context
		if songID, ok := ctx.Value("song_id").(uint); ok {
			collyCtx.Put("song_id", songID)
		}

		// 随机延迟，模拟人类行为
		time.Sleep(time.Duration(1000+time.Now().UnixNano()%2000) * time.Millisecond)

		if err := b.c.Request("GET", apiURL, nil, collyCtx, nil); err != nil {
			logger.Error("访问页面失败", "module", "collector.bilibili", "page", page, "error", err)
		}
	}

	b.c.Wait()
	return nil
}

func (b *BilibiliCollector) setupCallbacks() {
	b.c.OnRequest(func(r *colly.Request) {
		// 如果已封禁，取消请求
		if atomic.LoadInt32(&b.isBanned) == 1 {
			r.Abort()
			return
		}

		// 设置请求头以模拟真实浏览器，避免 412 WAF 错误
		if bvid := r.Ctx.Get("bvid"); bvid != "" {
			r.Headers.Set("Referer", fmt.Sprintf("https://www.bilibili.com/video/%s", bvid))
		} else {
			r.Headers.Set("Referer", "https://search.bilibili.com/all?keyword="+r.URL.Query().Get("keyword"))
		}

		// User-Agent is now handled by extensions.RandomUserAgent
		r.Headers.Set("Accept", "application/json, text/plain, */*")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		// 更新 Cookie，优先使用配置中的 Cookie
		if b.cookie != "" {
			r.Headers.Set("Cookie", b.cookie)
		} else {
			r.Headers.Set("Cookie", "buvid3=infoc; b_nut=1693881600;")
		}

		logger.Info("Visiting API", "module", "collector.bilibili", "url", r.URL.String())
	})

	b.c.OnError(func(r *colly.Response, err error) {
		logger.Error("请求失败", "module", "collector.bilibili", "url", r.Request.URL, "status", r.StatusCode, "error", err)

		// 检查是否为反爬虫错误 (412 Precondition Failed, 403 Forbidden)
		if r.StatusCode == http.StatusPreconditionFailed || r.StatusCode == http.StatusForbidden || r.StatusCode == 412 {
			atomic.StoreInt32(&b.isBanned, 1)
			logger.Error("触发反爬虫机制，已标记为封禁状态，停止后续请求", "module", "collector.bilibili")
		}
	})

	b.c.OnResponse(func(r *colly.Response) {
		// 检查是否为搜索 API
		switch r.Request.URL.Path {
		case "/x/web-interface/search/all/v2":
			b.handleSearchResponse(r.Body, r.Ctx)
		case "/x/v2/reply":
			b.handleReplyResponse(r.Body, r.Request.URL.Query().Get("oid"), r.Ctx)
		}
	})
}

func (b *BilibiliCollector) handleSearchResponse(body []byte, ctx *colly.Context) {
	json := string(body)

	// Check for API error code
	code := gjson.Get(json, "code").Int()
	if code != 0 {
		logger.Error("Bilibili API returned error", "module", "collector.bilibili", "code", code, "message", gjson.Get(json, "message").String())
		return
	}

	// 获取关联的歌曲信息用于相关性检查
	var song *model.Song
	if songIDVal := ctx.GetAny("song_id"); songIDVal != nil {
		if id, ok := songIDVal.(uint); ok {
			s, err := b.storage.GetSong(id)
			if err == nil {
				song = s
			}
		}
	}

	// 使用 gjson 解析视频列表
	// 路径: data.result[result_type="video"].data
	results := gjson.Get(json, "data.result")
	if !results.Exists() {
		logger.Warn("No results found in response", "module", "collector.bilibili")
		return
	}

	foundVideoSection := false
	results.ForEach(func(k, value gjson.Result) bool {
		if value.Get("result_type").String() == "video" {
			foundVideoSection = true
			videos := value.Get("data")
			logger.Info("Processing video section", "module", "collector.bilibili", "count", len(videos.Array()))
			videos.ForEach(func(k, v gjson.Result) bool {
				bvid := v.Get("bvid").String()
				aid := v.Get("id").Int()
				title := v.Get("title").String()
				desc := v.Get("description").String()
				author := v.Get("author").String()

				// 相关性检查
				if song != nil {
					cleanTitle := b.cleanHTML(title)
					if !b.isRelevant(cleanTitle, song) {
						logger.Info("跳过不相关视频", "module", "collector.bilibili", "title", cleanTitle, "song", song.Title)
						return true // continue
					}
				}

				logger.Info("Found Video", "module", "collector.bilibili", "bvid", bvid, "title", title, "author", author)

				// 保存视频信息为评论
				comment := &model.Comment{
					Source:      "Bilibili",
					SourceTitle: b.cleanHTML(title), // Clean HTML tags from title
					ExternalID:  bvid,
					Content:     desc, // 视频描述
					Author:      author,
					PostDate:    time.Now(), // 占位符
					SearchTag:   ctx.Get("keyword"),
				}

				// Link to SongID if available in context
				if songIDVal := ctx.GetAny("song_id"); songIDVal != nil {
					if songID, ok := songIDVal.(uint); ok {
						comment.SongID = &songID
					}
				}

				if err := b.storage.CreateComment(comment); err != nil {
					logger.Error("Failed to save video info", "module", "collector.bilibili", "error", err)
				}

				// 同时保存到独立的视频表
				video := &model.Video{
					Source:      "Bilibili",
					ExternalID:  bvid,
					Title:       b.cleanHTML(title), // Clean HTML tags from title
					Description: desc,
					Author:      author,
					URL:         fmt.Sprintf("https://www.bilibili.com/video/%s", bvid),
					PublishTime: time.Now(), // 占位符
				}
				if err := b.storage.CreateVideo(video); err != nil {
					logger.Error("Failed to save video record", "module", "collector.bilibili", "error", err)
				}

				// 获取该视频的评论
				// API: https://api.bilibili.com/x/v2/reply?type=1&oid={aid}&sort=1&ps=20
				replyURL := fmt.Sprintf("https://api.bilibili.com/x/v2/reply?type=1&oid=%d&sort=1&ps=20", aid)

				newCtx := colly.NewContext()
				newCtx.Put("title", b.cleanHTML(title))
				newCtx.Put("bvid", bvid)
				newCtx.Put("keyword", ctx.Get("keyword")) // Pass keyword to reply context
				// Pass song_id to reply context as well
				if songIDVal := ctx.GetAny("song_id"); songIDVal != nil {
					newCtx.Put("song_id", songIDVal)
				}

				if err := b.c.Request("GET", replyURL, nil, newCtx, nil); err != nil {
					logger.Error("请求评论失败", "module", "collector.bilibili", "url", replyURL, "error", err)
				}

				return true
			})
			return false // 找到视频部分后停止外层循环
		}
		return true
	})

	if !foundVideoSection {
		logger.Warn("No video section found in search results", "module", "collector.bilibili")
	}
}

func (b *BilibiliCollector) cleanHTML(src string) string {
	// 移除 <em> 和 </em> 标签
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(src, "")
}

func (b *BilibiliCollector) handleReplyResponse(body []byte, oid string, ctx *colly.Context) {
	json := string(body)
	title := ctx.Get("title")

	// 解析评论列表
	// 路径: data.replies
	replies := gjson.Get(json, "data.replies")
	replies.ForEach(func(key, value gjson.Result) bool {
		rpid := value.Get("rpid").String()
		content := value.Get("content.message").String()
		author := value.Get("member.uname").String()
		ctime := value.Get("ctime").Int()

		comment := &model.Comment{
			Source:      "Bilibili",
			SourceTitle: title,
			ExternalID:  rpid,
			Content:     content,
			Author:      author,
			PostDate:    time.Unix(ctime, 0),
			SearchTag:   ctx.Get("keyword"),
		}

		// Link to SongID if available in context
		if songIDVal := ctx.GetAny("song_id"); songIDVal != nil {
			if songID, ok := songIDVal.(uint); ok {
				comment.SongID = &songID
			}
		}

		if err := b.storage.CreateComment(comment); err != nil {
			logger.Error("保存评论失败", "module", "collector.bilibili", "rpid", rpid, "error", err)
		}
		return true
	})
}

func (b *BilibiliCollector) isRelevant(videoTitle string, song *model.Song) bool {
	videoTitle = strings.ToLower(videoTitle)
	songTitle := strings.ToLower(song.Title)

	// 1. 检查标题是否包含歌曲名
	if strings.Contains(videoTitle, songTitle) {
		return true
	}

	// 2. 检查标题是否包含任意一个有效的别名
	for _, alias := range song.Aliases {
		if len(alias.Alias) >= 2 && strings.Contains(videoTitle, strings.ToLower(alias.Alias)) {
			return true
		}
	}

	return false
}

// SetRelevanceAnalyzer 注入相关性分析器
func (b *BilibiliCollector) SetRelevanceAnalyzer(analyzer interface{}) {
	// 这里使用 interface{} 是为了避免循环依赖，实际应该定义一个接口
	// 但由于时间关系，我们暂时不在这里实现 LLM 标题检查，
	// 因为在 collector 中同步调用 LLM 会严重拖慢爬虫速度。
	// 建议将 LLM 标题检查放在后续的数据清洗/分析阶段，或者作为可选的增强功能。
}
