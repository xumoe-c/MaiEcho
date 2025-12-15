package main

import (
	"log"
	"strings"

	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/provider/divingfish"
	"github.com/xumoe-c/maiecho/server/internal/provider/yuzuchan"
	"github.com/xumoe-c/maiecho/server/internal/router"
	"github.com/xumoe-c/maiecho/server/internal/service"
	"github.com/xumoe-c/maiecho/server/internal/status"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

// @title MaiEcho API
// @version 1.0
// @description MaiEcho 是一个用于音乐评论和分析的服务器应用程序。
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 初始化状态模块
	status.Init()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 加载提示词
	prompts, err := config.LoadPrompts()
	if err != nil {
		log.Fatalf("加载提示词失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("启动 MaiEcho 服务器",
		"module", "main",
		"version", "1.0",
		"port", cfg.ServerPort,
	)

	// 初始化存储
	// 目前使用 sqlite，文件将创建在当前目录
	db, err := storage.NewDatabase(".\\server\\sqlite_db\\maiecho.db")
	if err != nil {
		logger.Fatal("连接数据库失败", "module", "main", "error", err)
	}

	// 初始化 LLM 客户端
	llmClient := llm.NewClient(cfg.LLM)

	// 初始化服务
	dfClient := divingfish.NewClient()
	yzClient := yuzuchan.NewClient()
	songService := service.NewSongService(db, dfClient, yzClient)
	collectorService := service.NewCollectorService(db, songService, cfg, llmClient, prompts)

	analysisService := service.NewAnalysisService(db, llmClient, prompts)

	// 启动调度器
	collectorService.StartScheduler()
	defer collectorService.StopScheduler()

	// 初始化路由
	r := router.NewRouter(songService, collectorService, analysisService)

	// 启动 API 服务器
	addr := cfg.ServerPort
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	logger.Info("启动服务器成功", "module", "main", "address", addr)
	if err := r.Run(addr); err != nil {
		logger.Fatal("启动服务器失败", "module", "main", "error", err)
	}
}
