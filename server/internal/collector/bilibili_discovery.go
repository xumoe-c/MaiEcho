package collector

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/tidwall/gjson"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

// BilibiliDiscoveryCollector 是一个用于发现 Bilibili 上新视频的采集器
type BilibiliDiscoveryCollector struct {
	storage storage.Storage
	c       *colly.Collector
}

func NewBilibiliDiscoveryCollector(s storage.Storage) *BilibiliDiscoveryCollector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.Async(true),
	)
	// 允许重复访问相同的 URL (因为我们需要定期扫描同一个搜索页面)
	c.AllowURLRevisit = true

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*bilibili.com*",
		Parallelism: 1,
		RandomDelay: 5 * time.Second,
	})

	bdc := &BilibiliDiscoveryCollector{
		storage: s,
		c:       c,
	}
	bdc.setupCallbacks()
	return bdc
}

func (b *BilibiliDiscoveryCollector) Name() string {
	return "bilibili_discovery"
}

func (b *BilibiliDiscoveryCollector) Collect(ctx context.Context, tag string) error {
	logger.Info("扫描标签", "module", "collector.bilibili_discovery", "collector", b.Name(), "tag", tag)

	// b站搜索API
	apiURL := fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/all/v2?keyword=%s&order=pubdate", url.QueryEscape(tag))

	return b.c.Visit(apiURL)
}

func (b *BilibiliDiscoveryCollector) setupCallbacks() {
	b.c.OnResponse(func(r *colly.Response) {
		if r.Request.URL.Path == "/x/web-interface/search/all/v2" {
			b.handleDiscoveryResponse(r.Body)
		}
	})
}

func (b *BilibiliDiscoveryCollector) handleDiscoveryResponse(body []byte) {
	json := string(body)

	// 解析列表
	results := gjson.Get(json, "data.result")
	results.ForEach(func(key, value gjson.Result) bool {
		if value.Get("result_type").String() == "video" {
			videos := value.Get("data")
			videos.ForEach(func(k, v gjson.Result) bool {
				bvid := v.Get("bvid").String()
				title := v.Get("title").String()
				// pubDate := v.Get("pubdate").Int()

				logger.Info("发现新视频", "module", "collector.bilibili_discovery", "bvid", bvid, "title", title)

				comment := &model.Comment{
					Source:     "Bilibili_Discovery",
					ExternalID: bvid,
					Content:    v.Get("description").String(),
					Author:     v.Get("author").String(),
					PostDate:   time.Now(), // 使用当前时间作为发现时间
				}

				if err := b.storage.CreateComment(comment); err != nil {
					// 忽略重复错误
					logger.Error("保存视频评论失败", "module", "collector.bilibili_discovery", "bvid", bvid, "error", err)
				}

				return true
			})
			return false
		}
		return true
	})
}
