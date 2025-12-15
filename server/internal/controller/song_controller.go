package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/service"
)

type SongController struct {
	Service service.SongService
}

func NewSongController(s service.SongService) *SongController {
	return &SongController{Service: s}
}

// GetSong 获取歌曲详情
// @Summary 获取歌曲详情
// @Description 获取歌曲及其谱面详情 (通过 GameID)
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "Game ID"
// @Success 200 {object} model.Song
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /songs/{id} [get]
func (c *SongController) GetSong(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("获取歌曲失败:GameID参数无效", "module", "controller.song", "idStr", idStr, "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的GameID"})
		return
	}

	song, err := c.Service.GetSongByGameID(id)
	if err != nil {
		logger.Error("获取歌曲失败:通过GameID查询失败", "module", "controller.song", "gameID", id, "error", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的歌曲"})
		return
	}

	logger.Info("成功获取歌曲信息", "module", "controller.song", "gameID", id, "title", song.Title)
	ctx.JSON(http.StatusOK, song)
}

// ListSongs 获取歌曲列表
// @Summary 获取歌曲列表
// @Description 支持分页和多种筛选条件
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   version  query     string   false  "版本"
// @Param   type     query     string   false  "类型 (DX/Standard)"
// @Param   genre    query     string   false  "流派"
// @Param   min_ds   query     number   false  "最小定数"
// @Param   max_ds   query     number   false  "最大定数"
// @Param   is_new   query     boolean  false  "是否新歌"
// @Param   keyword  query     string   false  "搜索关键词"
// @Param   page     query     int      false  "页码"
// @Param   page_size query    int      false  "每页数量"
// @Success 200 {object} model.SongListResponse
// @Failure 400 {object} map[string]string
// @Router /songs [get]
func (c *SongController) ListSongs(ctx *gin.Context) {
	var filter model.SongFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		logger.Warn("获取歌曲列表失败:查询参数绑定错误", "module", "controller.song", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认分页
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	result, err := c.Service.GetSongs(filter)
	if err != nil {
		logger.Error("获取歌曲列表失败:数据库查询错误", "module", "controller.song", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("成功获取歌曲列表", "module", "controller.song", "page", filter.Page, "pageSize", filter.PageSize, "total", result.Total)
	ctx.JSON(http.StatusOK, result)
}

// CreateSong 创建歌曲
// @Summary 创建歌曲
// @Description 向数据库添加新歌曲
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   song  body      model.Song  true  "Song JSON"
// @Success 200 {object} model.Song
// @Failure 400 {object} map[string]string
// @Router /songs [post]
func (c *SongController) CreateSong(ctx *gin.Context) {
	var song model.Song
	if err := ctx.ShouldBindJSON(&song); err != nil {
		logger.Warn("创建歌曲失败:请求体绑定错误", "module", "controller.song", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Service.CreateSong(&song); err != nil {
		logger.Error("创建歌曲失败:服务执行错误", "module", "controller.song", "title", song.Title, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("成功创建歌曲", "module", "controller.song", "gameID", song.ID, "title", song.Title)
	ctx.JSON(http.StatusOK, song)
}

// SyncSongs 同步歌曲数据
// @Summary 同步歌曲数据
// @Description 从Diving-Fish API获取并更新歌曲数据
// @Tags songs
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /songs/sync [post]
func (c *SongController) SyncSongs(ctx *gin.Context) {
	logger.Info("开始同步歌曲数据", "module", "controller.song")
	if err := c.Service.SyncFromDivingFish(); err != nil {
		logger.Error("歌曲数据同步失败", "module", "controller.song", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("歌曲数据同步完成", "module", "controller.song")
	ctx.JSON(http.StatusOK, gin.H{"message": "同步完成"})
}

// RefreshAliases 刷新歌曲别名
// @Summary 刷新歌曲别名
// @Description 从YuzuChan API获取最新的歌曲别名
// @Tags songs
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /songs/aliases/refresh [post]
func (c *SongController) RefreshAliases(ctx *gin.Context) {
	logger.Info("开始刷新歌曲别名", "module", "controller.song")
	if err := c.Service.RefreshAliases(); err != nil {
		logger.Error("别名刷新失败", "module", "controller.song", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("歌曲别名刷新完成", "module", "controller.song")
	ctx.JSON(http.StatusOK, gin.H{"message": "别名刷新成功"})
}
