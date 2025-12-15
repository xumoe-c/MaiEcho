# Config 模块 (Configuration)

## 1. 结构 (Structure)

*   `config.go`: 核心配置结构体定义与加载逻辑。
*   `prompt.go`: 提示词 (Prompt) 模板的加载与管理逻辑。
*   `README.md`: 本文档。

## 2. 功能 (Functionality)

### 2.1 应用配置 (App Configuration)
*   **多源加载**: 使用 `spf13/viper` 支持从 `config.yaml` 文件、环境变量 (Environment Variables) 和命令行参数加载配置。
*   **结构化定义**:
    *   `Server`: 端口、模式（Debug/Release）。
    *   `Database`: SQLite/MySQL 连接串。
    *   `LLM`: API Key、Base URL、模型名称、超时设置。
    *   `Log`: 日志级别、输出路径。
    *   `Collector`: 代理设置、Cookie 配置。

### 2.2 提示词管理 (Prompt Management)
*   **模板化**: 支持从 `prompts.yaml` 加载 Go Template 格式的提示词。
*   **动态渲染**: 提供 `ExecuteTemplate` 方法，支持在运行时注入变量（如歌曲信息、评论列表、谱面数据）生成最终 Prompt。
*   **版本管理**: 将 Prompt 与代码分离，便于独立迭代和调优。

## 3. 依赖关系 (Dependencies)

*   `github.com/spf13/viper`: 强大的配置管理库。
*   `gopkg.in/yaml.v3`: YAML 解析库。

## 4. 开发进度 (Status)

*   [x] 基础配置结构定义 (`Server`, `Database`, `LLM`, `Collector`)。
*   [x] 配置文件与环境变量加载逻辑。
*   [x] **Prompt 配置加载与模板渲染引擎**。

## 5. 计划 (Plan)

*   [ ] **配置热重载 (Hot Reload)**: 监听 `config.yaml` 和 `prompts.yaml` 的变化，无需重启服务即可更新配置。
*   [ ] **配置校验 (Validation)**: 引入 `go-playground/validator` 对关键配置项（如 API Key 格式、端口范围）进行启动时校验。
*   [ ] **多环境支持**: 完善 `config.dev.yaml`, `config.prod.yaml` 的加载策略。
