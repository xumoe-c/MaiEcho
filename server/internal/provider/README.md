# Provider 模块 (External Data Providers)

## 1. 结构 (Structure)
*   `divingfish/`: Diving-Fish (水鱼查分器) API 客户端。
*   `yuzuchan/`: YuzuChan API 客户端 (用于获取歌曲别名)。

## 2. 功能 (Functionality)
*   **外部数据获取**: 封装与第三方 API 的交互逻辑。
*   **数据适配**: 将第三方数据格式转换为内部模型 (`model`).

## 3. 依赖关系 (Dependencies)
*   `internal/model`: 数据模型。

## 4. 开发进度 (Status)
*   [x] Diving-Fish 客户端实现（含 ETag 缓存、统计数据聚合）。
*   [x] YuzuChan 客户端实现（获取歌曲别名）。
    *   *注：批量获取接口返回 403，目前采用逐个歌曲查询的策略。*

## 5. 计划 (Plan)
*   [ ] 增加更多数据源支持。
*   [ ] 统一的 Provider 接口定义。
