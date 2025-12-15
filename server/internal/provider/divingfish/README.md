# Diving-Fish Provider

## 1. 结构 (Structure)

* `client.go`: 客户端实现，包含 API 请求、ETag 处理和数据映射逻辑。

## 2. 功能 (Functionality)

* **歌曲同步**: 从 `/music_data` 获取完整歌曲列表。
* **统计数据**: 从 `/chart_stats` 获取谱面拟合难度和达成率统计。
* **智能缓存**: 利用 ETag 和 `If-None-Match` 避免重复下载未变更的数据。
* **封面映射**: 自动处理 DX 谱面 ID 映射逻辑以获取正确的封面 URL。

## 3. 依赖关系 (Dependencies)

* `internal/model`: 使用 `Song` 和 `Chart` 模型。

## 4. 开发进度 (Status)

* [X]  完整数据获取与映射（含拟合数据）。
* [X]  ETag 缓存机制。
* [X]  统计数据聚合。

## 5. 计划 (Plan)

* [ ]  增加重试机制。
* [ ]  支持获取单曲详情（如果 API 支持）。
