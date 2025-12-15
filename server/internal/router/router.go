package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/xumoe-c/maiecho/server/docs" // swagger doc
	"github.com/xumoe-c/maiecho/server/internal/controller"
	"github.com/xumoe-c/maiecho/server/internal/service"
)

func NewRouter(songService service.SongService, collectorService service.CollectorService, analysisService *service.AnalysisService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Controllers
	songController := controller.NewSongController(songService)
	collectorController := controller.NewCollectorController(collectorService)
	analysisController := controller.NewAnalysisController(analysisService)
	statusController := controller.NewStatusController()

	v1 := r.Group("/api/v1")
	{
		// System
		v1.GET("/system/status", statusController.GetStatus)

		v1.GET("/songs", songController.ListSongs)
		v1.GET("/songs/:id", songController.GetSong)
		v1.POST("/songs", songController.CreateSong)
		v1.POST("/songs/sync", songController.SyncSongs)
		v1.POST("/songs/aliases/refresh", songController.RefreshAliases)

		v1.POST("/collect", collectorController.TriggerCollection)
		v1.POST("/collect/backfill", collectorController.BackfillCollection)

		v1.POST("/analysis/songs/:id", analysisController.AnalyzeSong)
		v1.POST("/analysis/batch", analysisController.BatchAnalyzeSongs)
		v1.GET("/analysis/songs/:id", analysisController.GetAnalysisResult)
	}

	return r
}
