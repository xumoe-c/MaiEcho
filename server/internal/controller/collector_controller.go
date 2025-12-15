package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/service"
)

type CollectorController struct {
	Service service.CollectorService
}

func NewCollectorController(s service.CollectorService) *CollectorController {
	return &CollectorController{Service: s}
}

type CollectRequest struct {
	Keyword string `json:"keyword"`
	GameID  int    `json:"game_id"`
}

// TriggerCollection 触发数据收集任务
// @Summary 触发数据收集任务
// @Description 针对特定关键词或GameID启动数据收集任务
// @Tags collector
// @Accept  json
// @Produce  json
// @Param   request  body      CollectRequest  true  "Collection Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /collect [post]
func (c *CollectorController) TriggerCollection(ctx *gin.Context) {
	var req CollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("收集请求绑定失败", "module", "controller.collector", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.GameID > 0 {
		// 通过 GameID 获取歌曲信息
		song, err := c.Service.GetSongByGameID(req.GameID)
		if err != nil {
			logger.Error("通过GameID获取歌曲失败", "module", "controller.collector", "gameID", req.GameID, "error", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的歌曲"})
			return
		}

		logger.Info("开始为歌曲触发收集", "module", "controller.collector", "gameID", req.GameID, "title", song.Title)

		// 收集主标题
		keywords := []string{song.Title + " maimai"}
		// 如果标题较短，增加带作者的组合以提高精确度
		if len(song.Title) < 5 && song.Artist != "" {
			keywords = append(keywords, fmt.Sprintf("%s %s maimai", song.Title, song.Artist))
		}

		// 收集别名 (限制数量以防过多)
		for i, alias := range song.Aliases {
			if i >= 5 { // 稍微放宽限制
				break
			}
			if alias.Alias != "" {
				// 跳过过短的别名，避免匹配到无关内容
				if len(alias.Alias) < 2 {
					continue
				}

				// 使用 LLM 检查别名是否适合作为搜索关键词
				// 注意：这里是同步调用，可能会增加请求延迟。如果别名很多，建议改为异步或后台处理。
				// 为了用户体验，我们这里只对“可疑”的别名（例如长度较短或包含通用词）进行检查，或者全部检查但接受延迟。鉴于这是触发任务的接口，稍微的延迟是可以接受的。
				isSuitable, err := c.Service.CheckAliasSuitability(ctx, song, &song.Aliases[i])
				if err != nil {
					// 如果 LLM 失败，降级为默认接受（或者记录日志）
					logger.Warn("别名适合性检查失败，降级处理", "module", "controller.collector", "alias", alias.Alias, "error", err)
				} else if !isSuitable {
					logger.Info("跳过不适合的别名", "module", "controller.collector", "alias", alias.Alias)
					continue
				}

				keywords = append(keywords, alias.Alias+" 舞萌 maimai 手元 谱面确认")
			}
		}

		// 触发所有关键词的收集
		go func() {
			for _, kw := range keywords {
				if err := c.Service.TriggerCollection(kw, &song.ID); err != nil {
					logger.Error("触发关键词收集失败", "module", "controller.collector", "keyword", kw, "gameID", req.GameID, "error", err)
				}
				// 避免并发过高
				time.Sleep(2 * time.Second)
			}
			logger.Info("基于GameID的收集任务已完成", "module", "controller.collector", "gameID", req.GameID, "keywordCount", len(keywords))
		}()

		ctx.JSON(http.StatusOK, gin.H{"message": "基于GameID的数据收集任务已启动", "keywords": keywords})
		return
	}

	if req.Keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "必须提供 keyword 或 game_id"})
		return
	}

	// 附加 "舞萌 maimai 手元 谱面确认" 以提高相关性，避免过多无关数据
	// TODO:目前只是一个简单的做法，未来可以改进为更智能的关键词处理
	searchKeyword := req.Keyword + " 舞萌 maimai 手元 谱面确认"

	if err := c.Service.TriggerCollection(searchKeyword, nil); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "数据收集任务已启动", "keyword": searchKeyword})
}

// BackfillCollection 触发回填数据收集
// @Summary 触发回填数据收集
// @Description 用于初始化数据源，为无数据的歌曲收集评论数据
// @Tags collector
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collect/backfill [post]
func (c *CollectorController) BackfillCollection(ctx *gin.Context) {
	if err := c.Service.BackfillCollection(); err != nil {
		logger.Error("回填数据收集任务失败", "module", "controller.collector", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("回填数据收集任务已启动", "module", "controller.collector")
	ctx.JSON(http.StatusOK, gin.H{"message": "回填数据收集任务已排队"})
}
