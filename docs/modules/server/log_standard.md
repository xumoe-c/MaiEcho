# 日志模块介绍

## 概述

本项目内部模块的日志，使用 `server/internal/logger/logger.go` 提供的统一日志接口，并在所有日志调用中添加 `"module"` 标签以标明日志产生的源模块。

## 模块总览

### 1. Agent 模块（`server/internal/agent/`）

**文件数**：5 个**日志调用**：18 个

- `agent.go`: 模板解析失败错误日志
- `analyzer.go`: 评论搜索、分析操作的日志
- `cleaner.go`: LLM 结果解析日志
- `mapper.go`: 映射操作日志
- `relevance.go`: 相关性分析日志

**模块标签**：

- `"module", "agent"` - 主模块
- `"module", "agent.analyzer"` - 分析器子模块
- `"module", "agent.cleaner"` - 清理器子模块
- `"module", "agent.mapper"` - 映射器子模块
- `"module", "agent.relevance"` - 相关性子模块

### 2. Collector 模块（`server/internal/collector/`）

**文件数**：2 个**日志调用**：20 个

- `bilibili_discovery.go`: 标签扫描、视频发现、评论保存日志
- `bilibili.go`: 代理设置、API 请求、响应处理日志

**模块标签**：

- `"module", "collector.bilibili"` - Bilibili 采集器
- `"module", "collector.bilibili_discovery"` - Bilibili 发现采集器

### 3. Config 模块（`server/internal/config/`）

**文件数**：2 个**日志调用**：8 个

- `config.go`: 文件加载、解析、验证日志
- `prompt.go`: 提示词文件加载和解析日志

**模块标签**：

- `"module", "config"` - 配置模块
- `"module", "config.prompt"` - 提示词子模块

### 4. Controller 模块（`server/internal/controller/`）

**文件数**：4 个**日志调用**：20+ 个

- `analysis_controller.go`: 分析请求的错误和成功日志
- `collector_controller.go`: 采集请求的日志
- `song_controller.go`: 歌曲相关操作的日志
- `status_controller.go`: 系统状态查询日志

**模块标签**：

- `"module", "controller.analysis"` - 分析控制器
- `"module", "controller.collector"` - 采集控制器
- `"module", "controller.song"` - 歌曲控制器
- `"module", "controller.status"` - 状态控制器

### 5. LLM 模块（`server/internal/llm/`）

**文件数**：1 个**日志调用**：5 个

- `client.go`: API 请求失败和响应处理日志

**模块标签**：

- `"module", "llm"` - LLM 模块

**特殊功能**：实现了 LLM 对话专用日志记录功能，可将所有对话内容（包括系统提示、用户提示和响应）记录到单独文件。

### 6. Provider 模块（`server/internal/provider/`）

**文件数**：2 个**日志调用**：15 个

- `divingfish/client.go`: 谱面数据获取日志
- `yuzuchan/client.go`: 别名数据获取日志

**模块标签**：

- `"module", "provider.divingfish"` - Diving-Fish 数据提供商
- `"module", "provider.yuzuchan"` - YuzuChan 数据提供商

### 7. Scheduler 模块（`server/internal/scheduler/`）

**文件数**：1 个**日志调用**：8 个

- `scheduler.go`: 任务调度、工作线程日志

**模块标签**：

- `"module", "scheduler"` - 调度器模块

### 8. Service 模块（`server/internal/service/`）

**文件数**：2 个**日志调用**：12 个

- `collector_service.go`: 发现任务、回填任务日志
- `song_service.go`: 歌曲导入、别名刷新日志

**模块标签**：

- `"module", "service.collector"` - 采集服务
- `"module", "service.song"` - 歌曲服务

### 9. Status 模块（`server/internal/status/`）

**文件数**：1 个**日志调用**：1 个

- `status.go`: 系统状态日志

**模块标签**：

- `"module", "status"` - 状态模块

### 10. Main 模块（`server/cmd/maiecho/`）

**文件数**：1 个**日志调用**：2 个

- `main.go`: 服务器启动和初始化日志

**模块标签**：

- `"module", "main"` - 主程序模块

## 日志格式标准

### 基础格式

所有日志都遵循以下格式：

```go
logger.Info(message, "module", "module.submodule", "key1", value1, "key2", value2, ...)
```

### 示例

```go
// 信息日志
logger.Info("歌曲数据同步完成", "module", "provider.divingfish", "count", 100)

// 错误日志
logger.Error("采集失败", "module", "scheduler", "keyword", "舞萌", "error", err)

// 警告日志
logger.Warn("任务队列已满，丢弃任务", "module", "scheduler", "keyword", keyword)

// 调试日志
logger.Debug("任务启动", "module", "scheduler", "worker_id", 0)
```

### 模块标签命名约定

- **顶级模块**：`"module", "moduleName"`

  - 例如：`"module", "llm"`, `"module", "config"`
- **子模块**：`"module", "parentModule.childModule"`

  - 例如：`"module", "agent.analyzer"`, `"module", "provider.divingfish"`
- **服务子模块**：`"module", "service.serviceName"`

  - 例如：`"module", "service.collector"`, `"module", "service.song"`
- **控制器子模块**：`"module", "controller.controllerName"`

  - 例如：`"module", "controller.analysis"`, `"module", "controller.song"`

## 特殊功能

### LLM 对话日志

LLM 模块实现了对话记录专用日志功能，允许将所有 LLM API 调用记录到单独文件。

**配置**：

```yaml
log:
  llm_log_path: logs/llm_conversations.log
```

**记录内容**：

- 模型名称
- 系统提示词
- 用户提示词
- 模型响应
- 错误信息（如果有）

### 日志级别

项目使用以下日志级别：


| 级别  | 用途                   | 函数             |
| ----- | ---------------------- | ---------------- |
| INFO  | 记录关键操作、成功事件 | `logger.Info()`  |
| ERROR | 记录错误和异常         | `logger.Error()` |
| WARN  | 记录警告信息           | `logger.Warn()`  |
| DEBUG | 记录调试信息           | `logger.Debug()` |
| FATAL | 记录致命错误并退出     | `logger.Fatal()` |

## 关键特性

1. **可观测性**

   - 所有日志都带有模块标识，便于追踪问题源头
   - 结构化的键值对日志便于自动化分析
2. **调试体验**

   - 统一的日志格式和标准
   - 清晰的模块层级结构
   - 错误日志包含相关上下文信息
3. **监控和告警**

   - 可基于模块标签设置告警规则
   - 可按模块统计日志量和错误率
   - 支持日志聚合和分析
4. **性能和可维护性**

   - 使用 zap 日志库提供高性能日志记录
   - 支持多输出配置（控制台、文件）
   - 易于扩展和定制

## 配置文件

### 主配置文件位置

`server/config/config.yaml`

```yaml
log:
  level: info
  output_path: logs/maiecho.log
  encoding: console
  llm_log_path: logs/llm_conversations.log
```

### 环境变量支持

所有配置项都支持环境变量覆盖：

```bash
LOG_LEVEL=debug
LOG_OUTPUT_PATH=logs/debug.log
LOG_ENCODING=json
LOG_LLM_LOG_PATH=logs/llm.log
```
