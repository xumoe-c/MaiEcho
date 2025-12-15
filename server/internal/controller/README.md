# Controller 模块 (HTTP Handlers)

## 1. 结构 (Structure)

*   `song_controller.go`: 乐曲管理接口。负责歌曲列表查询、详情获取、别名刷新及与外部数据源（Diving-Fish）的同步。
*   `collector_controller.go`: 采集控制接口。负责触发针对特定歌曲或全量歌曲的评论采集任务。
*   `analysis_controller.go`: 智能分析接口。负责触发 LLM 分析流程及获取聚合后的分析报告。
*   `status_controller.go`: 系统状态接口。提供健康检查和版本信息。

## 2. 功能 (Functionality)

*   **请求解析与验证**: 使用 `gin` 的 `ShouldBindJSON` 等方法解析请求体，并进行基础参数校验。
*   **业务编排**: 调用 `Service` 层执行具体的业务逻辑（如采集、分析、存储）。
*   **响应标准化**: 统一封装 JSON 响应格式，处理 HTTP 状态码。
*   **Swagger 集成**: 包含 Swagger 注释，用于自动生成 API 文档。

## 3. 依赖关系 (Dependencies)

*   `github.com/gin-gonic/gin`: 高性能 Web 框架。
*   `internal/service`: 业务逻辑层接口。
*   `internal/model`: 数据模型定义。

## 4. 开发进度 (Status)

*   [x] **乐曲管理**: 列表/详情查询、同步 Diving-Fish 数据、别名刷新。
*   [x] **采集控制**: 单曲采集触发、后台批量采集。
*   [x] **智能分析**:
    *   单曲分析触发。
    *   批量分析触发。
    *   **聚合结果查询** (包含歌曲总览与各谱面详情)。
*   [x] **系统状态**: 健康检查接口。

## 5. 计划 (Plan)

*   [ ] **鉴权 (Auth)**: 添加 API Key 或 JWT 中间件，保护管理接口（如触发采集/分析）。
*   [ ] **限流 (Rate Limiting)**: 针对高频接口（如分析触发）添加限流中间件。
*   [ ] **WebSocket**: 实现实时进度推送（采集进度、分析进度）。

