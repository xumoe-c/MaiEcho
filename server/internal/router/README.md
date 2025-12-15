# Router 模块 (Routing)

## 1. 结构 (Structure)
*   `router.go`: 路由注册与中间件配置。

## 2. 功能 (Functionality)
*   **路由分发**: 将 HTTP 请求映射到对应的 Controller 方法。
*   **中间件配置**: 配置 CORS、Logger、Recovery 等中间件。
*   **Swagger 集成**: 注册 Swagger API 文档路由。

## 3. 依赖关系 (Dependencies)
*   `github.com/gin-gonic/gin`: Web 框架。
*   `internal/controller`: 控制器。

## 4. 开发进度 (Status)
*   [x] API v1 路由组设置。
*   [x] Swagger UI 路由。

## 5. 计划 (Plan)
*   [ ] 添加 API 版本控制 (v2)。
*   [ ] 实现更细粒度的权限控制路由。
