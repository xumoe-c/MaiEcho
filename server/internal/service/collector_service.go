package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/agent"
	"github.com/xumoe-c/maiecho/server/internal/collector"
	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/scheduler"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type CollectorService interface {
	// TriggerCollection 根据关键词触发一次采集任务
	TriggerCollection(keyword string, songID *uint) error
	// BackfillCollection 为数据库中的所有歌曲排队采集任务
	BackfillCollection() error
	// StartDiscovery 启动定期发现任务
	StartDiscovery()
	// StartScheduler 启动后台工作线程
	StartScheduler()
	// StopScheduler 停止后台工作线程
	StopScheduler()
	// GetSongByGameID 通过 GameID 获取歌曲
	GetSongByGameID(gameID int) (*model.Song, error)
	// CheckAliasSuitability 检查别名是否适合作为搜索关键词
	CheckAliasSuitability(ctx context.Context, song *model.Song, alias *model.SongAlias) (bool, error)
}

type collectorServiceImpl struct {
	scheduler          *scheduler.Scheduler
	songService        SongService
	storage            storage.Storage
	discoveryCollector collector.Collector
	discoveryTicker    *time.Ticker
	discoveryDone      chan bool
	relevanceAnalyzer  *agent.RelevanceAnalyzer
}

func NewCollectorService(s storage.Storage, songService SongService, cfg *config.Config, llmClient *llm.Client, prompts *config.PromptConfig) CollectorService {
	// 初始化采集器
	discovery := collector.NewBilibiliDiscoveryCollector(s)
	// 初始化主采集器 (用于深度采集评论)
	mainCollector := collector.NewBilibiliCollector(s, cfg.Bilibili.Cookie, cfg.Bilibili.Proxy)

	// 调度器同时管理发现采集器和主采集器
	collectors := []collector.Collector{discovery, mainCollector}

	// 初始化调度器
	sched := scheduler.NewScheduler(collectors, 1, 1000)

	return &collectorServiceImpl{
		scheduler:          sched,
		songService:        songService,
		storage:            s,
		discoveryCollector: discovery,
		discoveryDone:      make(chan bool),
		relevanceAnalyzer:  agent.NewRelevanceAnalyzer(llmClient, prompts),
	}
}

func (s *collectorServiceImpl) StartDiscovery() {
	// 每小时运行一次发现任务
	s.discoveryTicker = time.NewTicker(1 * time.Hour)

	go func() {
		// 立即运行一次
		s.runDiscovery()

		for {
			select {
			case <-s.discoveryDone:
				return
			case <-s.discoveryTicker.C:
				s.runDiscovery()
			}
		}
	}()
	logger.Info("发现服务已启动，每小时运行一次", "module", "service.collector")
}

func (s *collectorServiceImpl) runDiscovery() {
	logger.Info("运行发现任务", "module", "service.collector")
	spinner := logger.StartSpinner("运行发现任务...")
	ctx := context.Background()
	// 扫描常用标签
	tags := []string{"maimai", "舞萌DX", "maimai DX"}
	for _, tag := range tags {
		spinner.UpdateText(fmt.Sprintf("发现: 扫描标签 '%s'", tag))
		if err := s.discoveryCollector.Collect(ctx, tag); err != nil {
			logger.Error("发现失败", "module", "service.collector", "tag", tag, "error", err)
		}
	}
	spinner.Success("发现任务完成")
}

func (s *collectorServiceImpl) StartScheduler() {
	s.scheduler.Start()
	s.StartDiscovery() // 也在调度器启动时启动发现
}

func (s *collectorServiceImpl) StopScheduler() {
	s.scheduler.Stop()
	if s.discoveryTicker != nil {
		s.discoveryTicker.Stop()
		s.discoveryDone <- true
	}
}

func (s *collectorServiceImpl) TriggerCollection(keyword string, songID *uint) error {
	task := scheduler.Task{Keyword: keyword}
	if songID != nil {
		task.SongID = *songID
	}
	s.scheduler.AddTask(task)
	return nil
}

func (s *collectorServiceImpl) BackfillCollection() error {
	songs, err := s.songService.GetAllSongs()
	if err != nil {
		return fmt.Errorf("获取歌曲失败: %w", err)
	}

	logger.Info("开始回填", "module", "service.collector", "songCount", len(songs))

	// 如果歌曲数量较多，使用进度条显示回填任务排队进度
	tracker := logger.NewProgressTracker(len(songs), "排队回填任务中...")
	defer tracker.Stop()

	for _, song := range songs {
		// 构造关键词: "歌曲标题 舞萌 maimai 手元 谱面确认"
		keyword := fmt.Sprintf("%s 舞萌 maimai 手元 谱面确认", song.Title)
		s.scheduler.AddTask(scheduler.Task{Keyword: keyword, SongID: song.ID})
		tracker.Increment()
	}

	return nil
}

func (s *collectorServiceImpl) GetSongByGameID(gameID int) (*model.Song, error) {
	return s.songService.GetSongByGameID(gameID)
}

func (s *collectorServiceImpl) CheckAliasSuitability(ctx context.Context, song *model.Song, alias *model.SongAlias) (bool, error) {
	// 1. 检查缓存
	if alias.IsSuitable != nil {
		return *alias.IsSuitable, nil
	}

	// 2. 调用 LLM 分析
	isSuitable, err := s.relevanceAnalyzer.CheckAliasSuitability(ctx, song.Title, song.Artist, alias.Alias)
	if err != nil {
		return false, err
	}

	// 3. 更新缓存
	if err := s.storage.UpdateSongAliasSuitability(alias.ID, isSuitable); err != nil {
		logger.Error("更新别名适合性失败", "module", "service.collector", "alias_id", alias.ID, "error", err)
	} else {
		// 更新内存中的对象，避免重复检查
		alias.IsSuitable = &isSuitable
	}

	return isSuitable, nil
}
