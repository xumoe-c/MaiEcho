package collector

import (
	"context"
)

// Collector 定义了采集器接口
type Collector interface {
	// Name 返回采集器的名称（例如 "bilibili", "tieba"）
	Name() string

	// Collect 执行针对特定关键词的采集任务
	Collect(ctx context.Context, keyword string) error
}
