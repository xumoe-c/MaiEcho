# MaiEcho CLI Tester

这是一个用于测试 MaiEcho 服务器 API 的命令行工具。它提供了一个交互式界面，方便开发者调试和验证服务器功能。

## 功能特性

* **🔍 搜索歌曲 (Search Songs)**: 通过关键词搜索歌曲库。
* **📄 获取详情 (Get Song Details)**: 查看特定歌曲的详细信息，包括谱面数据。
* **📥 触发采集 (Trigger Collection)**: 手动触发针对特定歌曲的数据采集任务。
* **🤖 触发分析 (Trigger Analysis)**: 请求 LLM 对特定歌曲进行分析。
* **📊 获取分析结果 (Get Analysis Result)**: 查看歌曲的 AI 分析报告。
* **🔄 同步数据 (Sync Data)**: 触发与外部数据源（如 Diving Fish）的全量同步。
* **🏷️ 别名搜索 (Search Aliases)**: 测试别名搜索功能。

## 快速开始

### 前置要求

* Go 1.21+
* MaiEcho 服务器正在运行 (默认地址: `http://localhost:8080`)

### 编译

在项目根目录下运行以下命令编译 CLI 工具：

```bash
# 进入 CLI 目录
cd test/cli

# 编译 (Windows)
go build -o ../../maiecho-cli.exe .

# 编译 (Linux/macOS)
go build -o ../../maiecho-cli .
```

或者直接在根目录运行：

```bash
go build -o maiecho-cli.exe ./test/cli
```

### 运行

确保 MaiEcho 服务器已经启动，然后运行编译好的可执行文件：

```bash
./maiecho-cli.exe
```

## 使用说明

工具启动后会显示一个交互式菜单，使用键盘上下键选择功能，按回车确认。

### 1. 搜索歌曲

输入关键词（如 "兔子洞" 或 "Rabbit"），系统将返回匹配的歌曲列表，包含 ID、标题、艺术家和类型。

### 2. 获取详情

输入歌曲的 **Game ID**（不是数据库主键），查看歌曲的元数据和所有难度的谱面信息（定数、物量等）。

### 3. 触发采集

输入 Game ID，强制服务器重新从外部源采集该歌曲的信息。

### 4. 触发分析

输入 Game ID，将歌曲歌词和评论发送给 LLM 进行分析。这是一个异步任务。

### 5. 获取分析结果

输入 Game ID，查看 LLM 生成的分析报告。如果分析尚未完成，可能会返回空或提示进行中。
