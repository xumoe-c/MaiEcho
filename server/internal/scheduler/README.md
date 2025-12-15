# Scheduler 模块 (Task Scheduler)

## 1. 结构 (Structure)
*   `scheduler.go`: 任务调度器实现。

## 2. 功能 (Functionality)
*   **后台任务**: 管理和执行后台任务（如定时发现新歌、定期更新数据）。
*   **Worker Pool**: 简单的 Worker 池模型，并发处理任务。

## 3. 依赖关系 (Dependencies)
*   `internal/service`: 调用业务服务执行任务。

## 4. 开发进度 (Status)
*   [x] 基础 Worker Pool 实现。
*   [x] 支持启动/停止调度器。

## 5. 计划 (Plan)
*   [ ] 集成 Cron 表达式支持定时任务。
*   [ ] 支持分布式任务调度 (Redis/Etcd)。
*   [ ] 增加任务状态监控与持久化。
