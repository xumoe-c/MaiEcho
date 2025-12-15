package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/status"
)

type StatusController struct{}

func NewStatusController() *StatusController {
	return &StatusController{}
}

// GetStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取服务器运行状态、资源使用情况等
// @Tags system
// @Produce  json
// @Success 200 {object} status.SystemStatus
// @Router /system/status [get]
func (c *StatusController) GetStatus(ctx *gin.Context) {
	logger.Info("获取系统状态", "module", "controller.status")
	s := status.GetSystemStatus()
	ctx.JSON(http.StatusOK, s)
}
