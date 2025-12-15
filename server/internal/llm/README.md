# LLM 模块 (Large Language Model Client)

## 1. 结构 (Structure)
*   `client.go`: LLM 客户端封装。

## 2. 功能 (Functionality)
*   **API 交互**: 封装 `openai-go` SDK，与阿里云 DashScope (Qwen) 或其他兼容 OpenAI 协议的模型服务交互。
*   **请求封装**: 简化 Chat Completion 请求的构建。

## 3. 依赖关系 (Dependencies)
*   `github.com/openai/openai-go`: 官方 Go SDK。
*   `internal/config`: 获取 API Key 和 Base URL。

## 4. 开发进度 (Status)
*   [x] 基础客户端封装。
*   [x] 支持自定义 Base URL 和 Model。

## 5. 计划 (Plan)
*   [ ] 支持流式响应 (Streaming)。
*   [ ] 增加 Token 计数与成本估算。
*   [ ] 支持更多 LLM 提供商（如 DeepSeek, Claude）。
