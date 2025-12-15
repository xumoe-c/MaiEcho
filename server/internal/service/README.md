# Service 模块 (Business Logic)

## 1. 结构 (Structure)
*   `song_service.go`: 乐曲管理逻辑（同步、查询）。
*   `collector_service.go`: 采集任务管理逻辑。
*   `analysis_service.go`: 分析任务管理逻辑。
*   `service.go`: 服务接口定义。

## 2. 功能 (Functionality)
*   **业务编排**: 协调 Storage、LLM、Collector 等底层模块，实现具体的业务用例。
*   **数据同步**: 处理从 Diving-Fish API 同步数据的复杂逻辑（含 ETag 缓存）。
*   **别名刷新**: 从 YuzuChan API 获取并更新歌曲别名 (`RefreshAliases`)。
*   **分析聚合**: 实现 `GetAggregatedAnalysisResultByGameID`，将歌曲级分析与各谱面级分析结果聚合为统一视图。

## 3. 依赖关系 (Dependencies)
*   `internal/storage`: 数据存取。
*   `internal/llm`: AI 能力。
*   `internal/collector`: 数据采集能力。
*   `internal/agent`: 分析能力。
*   `internal/provider`: 外部数据提供商（如 Diving-Fish）。

## 4. 开发进度 (Status)
*   [x] 乐曲同步服务（含完整字段与缓存）。
*   [x] 采集与分析服务基础实现。
*   [x] **聚合分析逻辑**: 支持将 Song 和 Chart 的独立分析结果组合返回。

## 5. 计划 (Plan)
*   [ ] 增加事务管理 (Transaction) 支持。
*   [ ] 优化并发处理逻辑。
