# LLM 对话记录功能

## 概述

LLM 对话记录功能可将所有的 LLM API 调用（包括请求参数和响应内容）记录到单独的日志文件中，便于调试、审计和分析。

## 配置方法

### 方式一：在 config.yaml 中配置

```yaml
log:
  level: info
  output_path: logs/maiecho.log          # 主日志文件
  encoding: console                       # 日志格式
  llm_log_path: logs/llm_conversations.log  # LLM 对话日志文件路径
```

### 方式二：通过环境变量配置

```bash
export LOG_LLM_LOG_PATH=logs/llm_conversations.log
```

### 默认配置

如果不指定 `llm_log_path`，系统默认值为：

```
logs/llm_conversations.log
```

## 日志内容

LLM 对话日志以 JSON 格式记录，包含以下信息：

### 成功的对话记录

```json
{
  "level": "INFO",
  "timestamp": "2025-12-16T10:30:45.123Z",
  "caller": "llm/client.go:XX",
  "msg": "LLM对话成功",
  "model": "qwen-plus",
  "systemPrompt": "你是一个音游数据分析助手...",
  "userPrompt": "分析这个歌曲的难度数据...",
  "response": "根据提供的数据分析..."
}
```

### 失败的对话记录

```json
{
  "level": "ERROR",
  "timestamp": "2025-12-16T10:30:45.123Z",
  "caller": "llm/client.go:XX",
  "msg": "LLM对话失败",
  "model": "qwen-plus",
  "systemPrompt": "你是一个音游数据分析助手...",
  "userPrompt": "分析这个歌曲的难度数据...",
  "error": "connection timeout"
}
```

## 代码使用

### 在 LLM 客户端中自动记录

当使用 `llm.Client` 的 `Chat()` 或 `ChatWithReasoning()` 方法时，对话内容会自动记录到 LLM 专用日志文件中。

```go
// 对话会自动被记录
response, err := llmClient.Chat(ctx, systemPrompt, userPrompt)
```

### 手动记录对话（可选）

如果需要在其他地方手动记录 LLM 对话，可以使用 `logger` 包提供的函数：

```go
import "github.com/xumoe-c/maiecho/server/internal/logger"

// 记录对话
logger.LogLLMConversation(
    "qwen-plus",                    // 模型名称
    "你是一个音游数据分析助手",      // 系统提示
    "分析这个歌曲...",              // 用户提示
    "响应内容...",                   // 模型响应
    nil,                             // 错误（如果有的话）
)

// 记录失败的对话
logger.LogLLMConversation(
    "qwen-plus",
    systemPrompt,
    userPrompt,
    "",
    fmt.Errorf("API 请求失败"),
)
```

## 日志文件位置

- **主日志**：`logs/maiecho.log`
- **LLM 对话日志**：`logs/llm_conversations.log`

## 日志管理建议

1. **定期轮转**：使用日志轮转工具（如 logrotate）定期归档和压缩日志文件
2. **敏感信息**：注意日志可能包含用户输入和模型响应，确保日志文件的访问权限受限
3. **存储空间**：监控日志文件大小，特别是在高频使用的情况下
4. **分析工具**：可以使用 `jq` 等 JSON 处理工具分析日志文件

### 示例：查看最近 10 条 LLM 对话

```bash
tail -n 10 logs/llm_conversations.log | jq '.'
```

### 示例：统计特定模型的调用次数

```bash
grep '"model":"qwen-plus"' logs/llm_conversations.log | wc -l
```
