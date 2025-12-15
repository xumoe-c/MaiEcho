# Model 模块 (Data Models)

## 1. 结构 (Structure)

*   `song.go`: 乐曲 (`Song`)、谱面 (`Chart`) 及别名 (`SongAlias`) 的定义。
*   `comment.go`: 评论 (`Comment`) 数据定义。
*   `video.go`: 视频 (`Video`) 元数据定义。
*   `analysis.go`: 分析结果 (`AnalysisResult`) 定义。
*   `filter.go`: 查询过滤器定义。
*   `model.go`: 通用基础模型。

## 2. 核心实体 (Core Entities)

### 2.1 Song (乐曲)
*   **核心字段**: `GameID` (Diving-Fish ID), `Title`, `Artist`, `Type` (DX/Std)。
*   **关联**: 一对多关联 `Chart` 和 `SongAlias`。

### 2.2 Chart (谱面)
*   **核心字段**: `Difficulty` (Basic...Re:Master), `Level` (13+), `DS` (官方定数)。
*   **高级字段**: `FitDiff` (拟合定数), `AvgDX` (平均DX分), `AvgAchievement` (平均达成率)。
*   **用途**: 存储谱面的客观数据，用于辅助 Agent 进行“诈称/逆诈称”判断。

### 2.3 Comment (评论)
*   **核心字段**: `Source` (Bilibili), `SourceTitle` (视频标题), `Content`, `ExternalID` (rpid)。
*   **关联**: 可选关联 `SongID` 或 `ChartID`。
*   **用途**: 存储原始舆情数据。

### 2.4 AnalysisResult (分析结果)
*   **核心字段**:
    *   `TargetType`: "song" (歌曲总览) 或 "chart" (谱面详情)。
    *   `TargetID`: 对应的 SongID 或 ChartID。
    *   `Summary`: 简明摘要。
    *   `DifficultyAnalysis`: 难度分析文本。
    *   `RatingAdvice`: 推分建议。
    *   `ReasoningLog`: LLM 推理过程日志（用于调试）。

## 3. 依赖关系 (Dependencies)

*   `gorm.io/gorm`: ORM 框架，用于数据库交互。

## 4. 开发进度 (Status)

*   [x] 核心实体定义 (`Song`, `Chart`, `Comment`, `Video`)。
*   [x] **分析结果模型升级**: 支持 `TargetType` 字段，实现细粒度存储。
*   [x] **高级数据支持**: 包含拟合定数 (`FitDiff`) 等统计字段。

## 5. 计划 (Plan)

*   [ ] **数据库迁移 (Migration)**: 引入 `golang-migrate` 或 Gorm AutoMigrate 的版本控制，管理 Schema 变更。
*   [ ] **全文索引**: 为 `Comment.Content` 和 `Song.Title` 添加全文索引 (SQLite FTS5 / MySQL FullText)，加速模糊搜索。
*   [ ] **软删除策略**: 完善数据清理逻辑。
