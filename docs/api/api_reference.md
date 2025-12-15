# MaiEcho API 文档

本文档描述了 MaiEcho 服务端提供的 RESTful API 接口。

**Base URL**: `/api/v1`

## 1. 系统 (System)

### 1.1 获取系统状态
*   **GET** `/system/status`
*   **描述**: 检查服务器运行状态及版本信息。
*   **响应**:
    ```json
    {
      "status": "ok",
      "version": "1.0.0",
      "uptime": "1h23m"
    }
    ```

## 2. 乐曲管理 (Songs)

### 2.1 获取乐曲列表
*   **GET** `/songs`
*   **参数**:
    *   `page` (query, int): 页码，默认 1。
    *   `page_size` (query, int): 每页数量，默认 20。
    *   `keyword` (query, string): 搜索关键词（标题或别名）。
*   **响应**:
    ```json
    {
      "data": [
        {
          "id": 1001,
          "title": "Pandora Paradoxxx",
          "artist": "削除",
          "type": "DX",
          "cover_url": "..."
        }
      ],
      "total": 1500
    }
    ```

### 2.2 获取乐曲详情
*   **GET** `/songs/:id`
*   **描述**: 获取指定乐曲的详细信息，包括所有谱面数据。
*   **参数**:
    *   `id` (path, int): 乐曲 GameID。
*   **响应**:
    ```json
    {
      "id": 1001,
      "title": "Pandora Paradoxxx",
      "charts": [
        {
          "difficulty": "Master",
          "level": "13+",
          "ds": 13.7
        }
      ]
    }
    ```

### 2.3 同步乐曲数据
*   **POST** `/songs/sync`
*   **描述**: 触发从 Diving-Fish API 同步乐曲列表和定数数据。
*   **响应**: `200 OK`

### 2.4 刷新别名
*   **POST** `/songs/aliases/refresh`
*   **描述**: 触发从 Yuzuchan API 刷新乐曲别名库。
*   **响应**: `200 OK`

## 3. 数据采集 (Collection)

### 3.1 触发单曲采集
*   **POST** `/collect`
*   **描述**: 立即触发针对指定乐曲的评论采集任务。
*   **Body**:
    ```json
    {
      "game_id": 1001
    }
    ```
*   **响应**: `200 OK`

### 3.2 触发批量采集 (Backfill)
*   **POST** `/collect/backfill`
*   **描述**: 启动后台任务，对所有未采集或数据过期的乐曲进行批量采集。
*   **响应**: `200 OK`

## 4. 智能分析 (Analysis)

### 4.1 触发单曲分析
*   **POST** `/analysis/songs/:id`
*   **描述**: 触发针对指定乐曲的 LLM 分析流程。
*   **参数**:
    *   `id` (path, int): 乐曲 GameID。
*   **响应**: `200 OK` (异步任务) 或 `200 OK` (同步等待，视配置而定)

### 4.2 触发批量分析
*   **POST** `/analysis/batch`
*   **Body**:
    ```json
    {
      "game_ids": [1001, 1002, 1003]
    }
    ```
*   **响应**: `200 OK`

### 4.3 获取分析结果 (聚合)
*   **GET** `/analysis/songs/:id`
*   **描述**: 获取指定乐曲的完整分析报告，包含歌曲总览和各谱面详情。
*   **参数**:
    *   `id` (path, int): 乐曲 GameID。
*   **响应**:
    ```json
    {
      "song_result": {
        "summary": "...",
        "rating_advice": "...",
        "target_type": "song"
      },
      "chart_results": [
        {
          "target_type": "chart",
          "target_id": 5001,
          "summary": "[DX Master] 本谱面...",
          "difficulty_analysis": "...",
          "rating_advice": "..."
        },
        {
          "target_type": "chart",
          "target_id": 5002,
          "summary": "[Std Master] 旧谱面...",
          ...
        }
      ]
    }
    ```
