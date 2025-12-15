# Storage 模块 (Data Persistence)

## 1. 结构 (Structure)
*   `database.go`: 数据库连接与初始化。
*   `storage.go`: 存储接口定义与实现。

## 2. 功能 (Functionality)
*   **数据库连接**: 管理 SQLite (或 PostgreSQL) 连接池。
*   **CRUD 操作**: 提供对 Song, Comment, AnalysisResult 等实体的增删改查方法。
*   **别名管理**: 支持保存和查询歌曲别名 (`SaveSongAliases`)。
*   **关联查询**: 支持通过 SongID 查询关联评论 (`GetCommentsBySongID`)。
*   **细粒度查询**: 支持通过 `TargetType` 和 `TargetID` 查询特定的分析结果 (`GetAnalysisResultsByTarget`)。
*   **自动迁移**: 使用 GORM AutoMigrate 自动同步表结构。

## 3. 依赖关系 (Dependencies)
*   `gorm.io/gorm`: ORM 库。
*   `gorm.io/driver/sqlite`: SQLite 驱动。

## 4. 开发进度 (Status)
*   [x] SQLite 基础支持。
*   [x] 核心实体的 CRUD 方法。
*   [x] **多态存储支持**: 适配 `AnalysisResult` 的 `TargetType` 字段查询。

## 5. 计划 (Plan)
*   [ ] 引入 PostgreSQL 支持。
*   [ ] 实现更复杂的查询构建器 (Query Builder)。
*   [ ] 增加 Redis 缓存层。
