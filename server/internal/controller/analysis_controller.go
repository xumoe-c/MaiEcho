package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/service"
)

type AnalysisController struct {
	service *service.AnalysisService
}

func NewAnalysisController(s *service.AnalysisService) *AnalysisController {
	return &AnalysisController{service: s}
}

// AnalyzeSong 分析歌曲
// @Summary 分析歌曲
// @Description 触发针对指定歌曲ID(GameID)的LLM分析流程
// @Tags analysis
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /analysis/songs/{id} [post]
func (c *AnalysisController) AnalyzeSong(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("无效的GameID", "module", "controller.analysis", "idStr", idStr, "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的GameID"})
		return
	}

	logger.Info("开始分析歌曲", "module", "controller.analysis", "gameID", id)

	// 运行在后台还是阻塞？
	// 目前为了立即看到结果，采用阻塞方式，但LLM响应较慢。
	// 为了MVP的简单性，暂时采用阻塞方式。
	if err := c.service.AnalyzeSongByGameID(ctx.Request.Context(), id); err != nil {
		logger.Error("歌曲分析失败", "module", "controller.analysis", "gameID", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("歌曲分析完成", "module", "controller.analysis", "gameID", id)
	ctx.JSON(http.StatusOK, gin.H{"message": "分析完成"})
}

type BatchAnalysisRequest struct {
	GameIDs []int `json:"game_ids" binding:"required"`
}

// BatchAnalyzeSongs 批量分析歌曲
// @Summary 批量分析歌曲
// @Description 触发针对多个歌曲ID(GameID)的LLM分析流程
// @Tags analysis
// @Accept json
// @Produce json
// @Param request body BatchAnalysisRequest true "Batch Analysis Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /analysis/batch [post]
func (c *AnalysisController) BatchAnalyzeSongs(ctx *gin.Context) {
	var req BatchAnalysisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("批量分析请求绑定失败", "module", "controller.analysis", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("启动批量分析任务", "module", "controller.analysis", "count", len(req.GameIDs))

	// 简单的循环调用，实际生产中应放入消息队列
	go func() {
		for _, id := range req.GameIDs {
			// 使用新的 context，因为请求 context 会在请求结束时取消
			if err := c.service.AnalyzeSongByGameID(context.Background(), id); err != nil {
				logger.Error("批量分析中单个歌曲分析失败", "module", "controller.analysis", "gameID", id, "error", err)
				continue
			}
		}
		logger.Info("批量分析任务完成", "module", "controller.analysis", "count", len(req.GameIDs))
	}()

	ctx.JSON(http.StatusOK, gin.H{"message": "批量分析任务已在后台启动"})
}

// AnalysisResponse 聚合了歌曲和谱面的分析结果
type AnalysisResponse struct {
	SongResult   *model.AnalysisResult   `json:"song_result"`
	ChartResults []*model.AnalysisResult `json:"chart_results"`
}

// GetAnalysisResult 获取分析结果
// @Summary 获取分析结果
// @Description 获取指定歌曲(GameID)的最新分析结果，包含歌曲总览和各谱面详情
// @Tags analysis
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} AnalysisResponse
// @Failure 404 {object} map[string]string
// @Router /analysis/songs/{id} [get]
func (c *AnalysisController) GetAnalysisResult(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("无效的GameID", "module", "controller.analysis", "idStr", idStr, "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的GameID"})
		return
	}

	result, err := c.service.GetAggregatedAnalysisResultByGameID(id)
	if err != nil {
		logger.Error("获取分析结果失败", "module", "controller.analysis", "gameID", id, "error", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "未找到分析结果"})
		return
	}

	logger.Info("获取分析结果成功", "module", "controller.analysis", "gameID", id)
	ctx.JSON(http.StatusOK, result)
}
