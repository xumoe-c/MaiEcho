# MaiEcho 文档目录

## 🌏 项目概览 (Overview)

- [项目主页 (Main README)](../README.md)
- [架构设计 (Architecture)](modules/server/architecture.md)
- [技术设计 (Technical Design)](modules/server/technical_design.md)

## 🧩 服务端模块 (Server Modules)

### 核心组件 (Core)

- [配置管理 (Config)](../server/internal/config/README.md)
- [数据模型 (Model)](../server/internal/model/README.md)
- [存储层 (Storage)](../server/internal/storage/README.md)
- [路由层 (Router)](../server/internal/router/README.md)

### 业务逻辑 (Business Logic)

- [控制器 (Controller)](../server/internal/controller/README.md)
- [服务层 (Service)](../server/internal/service/README.md)

### 功能模块 (Features)

- [智能体 (Agent)](../server/internal/agent/README.md)
- [数据采集器 (Collector)](../server/internal/collector/README.md)
- [任务调度 (Scheduler)](../server/internal/scheduler/README.md)
- [LLM 客户端 (LLM)](../server/internal/llm/README.md)

### 外部集成 (Integrations)

- [服务提供商 (Provider)](../server/internal/provider/README.md)
  - [Diving Fish Client](../server/internal/provider/divingfish/README.md)

## 🎨 客户端 (Client)

- [设计系统 (Design System)](modules/client/DESIGN_SYSTEM.md)

## 📏 规范与标准 (Standards)

- [API 参考 (API Reference)](api/api_reference.md)
- [日志规范 (Log Standard)](modules/server/log_standard.md)
- [LLM 日志规范 (LLM Log Standard)](modules/server/llm_log_standard.md)

## 🛠️ 工具 (Tools)

- [CLI 测试工具](../test/cli/README.md)
