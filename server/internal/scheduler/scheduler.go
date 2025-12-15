package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/collector"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type Task struct {
	Keyword string
	Source  string // 可选：指定采集器名称以仅使用特定采集器
	SongID  uint   // 可选：关联的歌曲ID
}

type Scheduler struct {
	taskQueue   chan Task
	collectors  []collector.Collector
	workerCount int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewScheduler(collectors []collector.Collector, workerCount int, bufferSize int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		taskQueue:   make(chan Task, bufferSize),
		collectors:  collectors,
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Scheduler) Start() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	logger.Info("调度器启动", "module", "scheduler", "workers", s.workerCount)
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	// TODO:清空队列？还是直接关闭。
	// 目前，我们只是停止处理。
	logger.Info("调度器停止", "module", "scheduler")
}

func (s *Scheduler) AddTask(task Task) {
	select {
	case s.taskQueue <- task:
		logger.Info("任务已添加到队列", "module", "scheduler", "keyword", task.Keyword)
	default:
		logger.Warn("任务队列已满，丢弃任务", "module", "scheduler", "keyword", task.Keyword)
	}
}

func (s *Scheduler) worker(id int) {
	defer s.wg.Done()
	logger.Debug("任务启动", "module", "scheduler", "worker_id", id)

	for {
		select {
		case <-s.ctx.Done():
			logger.Debug("任务停止", "module", "scheduler", "worker_id", id)
			return
		case task := <-s.taskQueue:
			// 任务处理的视觉反馈
			spinnerText := fmt.Sprintf("任务 %d 处理中: %s", id, task.Keyword)
			spinner := logger.StartSpinner(spinnerText)

			// 在所有适用的采集器上执行任务
			for _, c := range s.collectors {
				if task.Source != "" && c.Name() != task.Source {
					continue
				}

				// 全局节奏控制：在任务之间稍作休眠以确保安全
				// 这是对内部采集器速率限制的补充
				time.Sleep(5 * time.Second)

				// Create a context with SongID if present
				ctx := s.ctx
				if task.SongID != 0 {
					ctx = context.WithValue(ctx, "song_id", task.SongID)
				}

				if err := c.Collect(ctx, task.Keyword); err != nil {
					spinner.Fail(fmt.Sprintf("任务 %d: 采集器 %s 处理 %s 失败: %v", id, c.Name(), task.Keyword, err))
					logger.Error("采集失败",
						"module", "scheduler",
						"worker_id", id,
						"collector", c.Name(),
						"keyword", task.Keyword,
						"error", err,
					)
				} else {
					logger.Info("采集成功",
						"module", "scheduler",
						"worker_id", id,
						"collector", c.Name(),
						"keyword", task.Keyword,
					)
				}
			}
			spinner.Success(fmt.Sprintf("任务 %d 完成: %s", id, task.Keyword))
		}
	}
}
