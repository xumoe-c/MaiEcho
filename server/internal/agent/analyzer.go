package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type Analyzer struct {
	storage storage.Storage
	llm     *llm.Client
	cleaner *Cleaner
	mapper  *Mapper
	kb      *KnowledgeBase
	prompts *config.PromptConfig
}

func NewAnalyzer(s storage.Storage, llm *llm.Client, prompts *config.PromptConfig) *Analyzer {
	return &Analyzer{
		storage: s,
		llm:     llm,
		cleaner: NewCleaner(llm, prompts),
		mapper:  NewMapper(s, llm, prompts),
		kb:      NewKnowledgeBase(prompts),
		prompts: prompts,
	}
}

// ChartContext 代表从评论来源推断出的谱面上下文
type ChartContext struct {
	Version    string // "DX", "Std", or "" (unknown)
	Difficulty string // "Master", "Re:Master", "Expert", or "" (unknown)
}

func (a *Analyzer) parseChartContext(title string) ChartContext {
	lowerTitle := strings.ToLower(title)
	ctx := ChartContext{}

	// 解析版本
	if strings.Contains(lowerTitle, "dx") || strings.Contains(lowerTitle, "deluxe") || strings.Contains(lowerTitle, "2p") {
		ctx.Version = "DX"
	} else if strings.Contains(lowerTitle, "std") || strings.Contains(lowerTitle, "standard") || strings.Contains(lowerTitle, "标准") {
		ctx.Version = "Std"
	}

	// 解析难度
	if strings.Contains(lowerTitle, "re:master") || strings.Contains(lowerTitle, "remaster") || strings.Contains(lowerTitle, "白") {
		ctx.Difficulty = "Re:Master"
	} else if strings.Contains(lowerTitle, "master") || strings.Contains(lowerTitle, "紫") || strings.Contains(lowerTitle, "13+") || strings.Contains(lowerTitle, "14") || strings.Contains(lowerTitle, "14+") || strings.Contains(lowerTitle, "15") {
		// 注意：简单的 "13" 可能指代 Expert 13 或 Master 13，这里保守处理，主要依赖文字描述
		ctx.Difficulty = "Master"
	} else if strings.Contains(lowerTitle, "expert") || strings.Contains(lowerTitle, "红") {
		ctx.Difficulty = "Expert"
	}

	return ctx
}

func (a *Analyzer) isOfficialChart(title string) bool {
	lowerTitle := strings.ToLower(title)
	keywords := []string{"自制", "自作", "ugc", "宴", "world's end", "we", "改谱", "fanmade"}
	for _, kw := range keywords {
		if strings.Contains(lowerTitle, kw) {
			return false
		}
	}
	return true
}

// RunMapping 触发评论到歌曲的映射过程
func (a *Analyzer) RunMapping(ctx context.Context) error {
	return a.mapper.MapCommentsToSongs(ctx)
}

// AnalyzeSong 对一首歌曲执行完整的分析流程
func (a *Analyzer) AnalyzeSong(ctx context.Context, songID uint) error {
	// 1. 获取歌曲信息
	song, err := a.storage.GetSong(songID)
	if err != nil {
		return fmt.Errorf("获取歌曲失败: %w", err)
	}

	// 2. 获取评论
	comments, err := a.storage.GetCommentsBySongID(song.ID)
	if err != nil {
		return fmt.Errorf("获取评论失败: %w", err)
	}

	// 降级策略：如果没有关联的评论，尝试关键词搜索
	if len(comments) == 0 {
		logger.Info("未找到歌曲的关联评论，正在尝试关键词搜索", "module", "agent.analyzer", "songTitle", song.Title)

		// 按标题搜索
		byTitle, err := a.storage.GetCommentsByKeyword(song.Title)
		if err != nil {
			logger.Error("按标题获取评论失败", "module", "agent.analyzer", "songTitle", song.Title, "error", err)
		} else {
			comments = append(comments, byTitle...)
		}

		// 按别名搜索
		for _, alias := range song.Aliases {
			if alias.Alias == "" {
				continue
			}
			byAlias, err := a.storage.GetCommentsByKeyword(alias.Alias)
			if err != nil {
				logger.Error("按别名获取评论失败", "module", "agent.analyzer", "alias", alias.Alias, "error", err)
				continue
			}
			comments = append(comments, byAlias...)
		}
	}

	if len(comments) == 0 {
		logger.Info("未找到歌曲的评论", "module", "agent.analyzer", "songTitle", song.Title)
		return nil
	}

	// 3. 清洗和分桶评论
	// Buckets: ChartID -> []string (comments)
	// Key 0 represents "General/Unclassified" bucket
	commentBuckets := make(map[uint][]string)
	seenComments := make(map[string]bool)

	for _, c := range comments {
		// 3.0 预过滤：剔除非官方谱面（自制、宴、UGC等）
		if !a.isOfficialChart(c.SourceTitle) {
			continue
		}

		cleaned := a.cleaner.Clean(c.Content)
		if !a.cleaner.IsValid(cleaned) {
			continue
		}

		// 去重
		if seenComments[cleaned] {
			continue
		}
		seenComments[cleaned] = true

		// 3.0.1 解析上下文并映射到 ChartID
		chartCtx := a.parseChartContext(c.SourceTitle)
		var targetChartID uint = 0 // Default to general

		// 尝试匹配具体的 Chart
		if chartCtx.Difficulty != "" {
			for _, chart := range song.Charts {
				// 匹配难度
				if chart.Difficulty != chartCtx.Difficulty {
					continue
				}

				// 匹配版本 (如果 Song 有 Type 字段区分 DX/Std，或者 Chart 有相关字段)
				// 假设 Song.Type 指示了这首歌是 DX 还是 Std 版本
				// 如果评论指定了 DX，而 Song 是 Std，则不匹配 (反之亦然)
				// 注意：这里简化处理，如果 Song.Type 为空或不明确，主要依赖难度匹配
				if chartCtx.Version != "" {
					// 如果评论明确说了 DX，但歌曲类型是 Standard，可能是在 Standard 歌曲下讨论 DX 谱面（如果有的话），或者这首歌本身就是 DX 版本。
					// 这里我们需要更复杂的逻辑，但暂时假设 Song.Type 能区分
					if chartCtx.Version == "DX" && song.Type == "SD" {
						continue
					}
					if chartCtx.Version == "Std" && song.Type == "DX" {
						continue
					}
				}

				targetChartID = chart.ID
				break
			}
		}

		// 格式: [标题] 评论内容
		commentWithContext := fmt.Sprintf("[%s] %s", c.SourceTitle, cleaned)
		commentBuckets[targetChartID] = append(commentBuckets[targetChartID], commentWithContext)
	}

	// 3.1 LLM 深度清洗 (Semantic Cleaning) - 对每个桶分别清洗太耗时，这里简化为只对总数过多的桶清洗
	// 或者跳过 LLM 清洗，直接进入分析阶段，依靠 Analyst 的能力过滤噪声

	if len(commentBuckets) == 0 {
		return nil
	}

	// 4. 分桶分析
	// 4.1 分析通用桶 (ChartID = 0)
	if generalComments, ok := commentBuckets[0]; ok && len(generalComments) > 0 {
		// 为了简化，我们可以把通用评论也作为上下文提供给具体谱面分析，或者单独出一个“歌曲综述”
	}

	// 4.2 分析具体谱面桶
	for chartID, bucketComments := range commentBuckets {
		if chartID == 0 {
			continue // 稍后处理通用桶
		}

		// 获取对应的 Chart 对象
		var targetChart model.Chart
		for _, ch := range song.Charts {
			if ch.ID == chartID {
				targetChart = ch
				break
			}
		}

		if err := a.analyzeChartBucket(ctx, song, &targetChart, bucketComments); err != nil {
			logger.Error("分析谱面失败", "module", "agent.analyzer", "chartID", chartID, "error", err)
		}
	}

	// 4.3 (可选) 如果没有具体谱面的评论，或者为了生成总览，可以分析通用桶
	// 这里我们暂时保留原有的全量分析逻辑作为“总览”，但只使用通用桶 + 少量随机抽样的具体桶数据；或者简单地跳过总览，只依赖谱面分析结果的聚合。
	// 为了保持兼容性，我们还是生成一个 Song 级别的 AnalysisResult，作为“总览”

	// 收集所有评论用于生成总览 (Summary)
	var allComments []string
	for _, bucket := range commentBuckets {
		allComments = append(allComments, bucket...)
	}

	// 限制总览的评论数量，避免 token 溢出
	if len(allComments) > 50 {
		allComments = allComments[:50]
	}

	// 4. Map-Reduce 分析 (针对 Song 级别)
	validComments := allComments
	var analystOutputs []*AnalystOutput
	var reasoningLogs []string // 收集推理日志
	chunkSize := 50

	for i := 0; i < len(validComments); i += chunkSize {
		end := i + chunkSize
		if end > len(validComments) {
			end = len(validComments)
		}
		chunk := validComments[i:end]
		chunkStr := "- " + strings.Join(chunk, "\n- ")

		// 为该块注入知识库
		relevantTerms := a.kb.GetRelevantTerms(chunkStr)
		termGuide := a.kb.FormatTerms(relevantTerms)

		// 准备别名字符串
		var aliases []string
		for _, alias := range song.Aliases {
			aliases = append(aliases, alias.Alias)
		}
		aliasStr := strings.Join(aliases, ", ")

		// 准备谱面信息字符串 (只包含 Expert, Master, Re:Master)
		chartInfoStr := a.formatChartInfo(song.Charts)

		// 对块运行分析师
		out, reasoning, err := a.runAnalyst(ctx, chunkStr, termGuide, aliasStr, chartInfoStr)
		if err != nil {
			logger.Error("分析块失败", "module", "agent.analyzer", "startIdx", i, "endIdx", end, "error", err)
			continue
		}
		analystOutputs = append(analystOutputs, out)
		if reasoning != "" {
			reasoningLogs = append(reasoningLogs, fmt.Sprintf("--- Chunk %d-%d Analysis ---\n%s", i, end, reasoning))
		}
	}

	if len(analystOutputs) == 0 {
		return fmt.Errorf("所有分析块均失败")
	}

	// 合并结果
	mergedAnalystOutput := a.mergeAnalystOutputs(analystOutputs)

	// 6. 运行顾问（主观建议）
	advisorOutput, err := a.runAdvisor(ctx, song, mergedAnalystOutput)
	if err != nil {
		return fmt.Errorf("顾问运行失败: %w", err)
	}

	// 7. 保存结果
	result := &model.AnalysisResult{
		TargetType:         "song",
		TargetID:           song.ID,
		Summary:            advisorOutput.Summary,
		RatingAdvice:       advisorOutput.RatingAdvice,
		DifficultyAnalysis: advisorOutput.DifficultyAnalysis,
		ReasoningLog:       strings.Join(reasoningLogs, "\n\n"),
	}

	if err := a.storage.CreateAnalysisResult(result); err != nil {
		return fmt.Errorf("保存分析结果失败: %w", err)
	}

	logger.Info("已保存歌曲的分析结果", "module", "agent.analyzer", "songTitle", song.Title)
	return nil
}

func (a *Analyzer) runAnalyst(ctx context.Context, comments, termGuide, aliases, chartInfo string) (*AnalystOutput, string, error) {
	systemData := struct {
		TermGuide string
		Aliases   string
		ChartInfo string
	}{
		TermGuide: termGuide,
		Aliases:   aliases,
		ChartInfo: chartInfo,
	}
	systemPrompt, err := ExecuteTemplate(a.prompts.Agent.Analyst.System, systemData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute system prompt template: %w", err)
	}

	userData := struct {
		Comments string
	}{
		Comments: comments,
	}
	userPrompt, err := ExecuteTemplate(a.prompts.Agent.Analyst.User, userData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute user prompt template: %w", err)
	}

	content, reasoning, err := a.llm.ChatWithReasoning(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, "", err
	}

	// 清理响应（移除 markdown 代码块）
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var output AnalystOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, reasoning, fmt.Errorf("解析分析师输出失败: %w. 响应: %s", err, content)
	}
	return &output, reasoning, nil
}

func (a *Analyzer) runAdvisor(ctx context.Context, song *model.Song, analystData *AnalystOutput) (*AdvisorOutput, error) {
	analystJson, _ := json.Marshal(analystData)

	// 准备别名字符串
	var aliases []string
	for _, alias := range song.Aliases {
		aliases = append(aliases, alias.Alias)
	}
	aliasStr := strings.Join(aliases, ", ")

	systemData := struct {
		Title   string
		Artist  string
		Aliases string
	}{
		Title:   song.Title,
		Artist:  song.Artist,
		Aliases: aliasStr,
	}
	systemPrompt, err := ExecuteTemplate(a.prompts.Agent.Advisor.System, systemData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute system prompt template: %w", err)
	}

	userData := struct {
		AnalysisData string
	}{
		AnalysisData: string(analystJson),
	}
	userPrompt, err := ExecuteTemplate(a.prompts.Agent.Advisor.User, userData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute user prompt template: %w", err)
	}

	response, err := a.llm.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// 清理响应
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	var output AdvisorOutput
	if err := json.Unmarshal([]byte(response), &output); err != nil {
		return nil, fmt.Errorf("解析顾问输出失败: %w. 响应: %s", err, response)
	}
	return &output, nil
}

// analyzeChartBucket 对单个谱面的评论桶进行分析
func (a *Analyzer) analyzeChartBucket(ctx context.Context, song *model.Song, chart *model.Chart, comments []string) error {
	if len(comments) == 0 {
		return nil
	}

	logger.Info("开始分析谱面", "module", "agent.analyzer", "chartID", chart.ID, "difficulty", chart.Difficulty, "level", chart.Level, "commentCount", len(comments))

	// 1. 准备上下文
	chunkStr := "- " + strings.Join(comments, "\n- ")

	// 知识库注入
	relevantTerms := a.kb.GetRelevantTerms(chunkStr)
	termGuide := a.kb.FormatTerms(relevantTerms)

	// 别名
	var aliases []string
	for _, alias := range song.Aliases {
		aliases = append(aliases, alias.Alias)
	}
	aliasStr := strings.Join(aliases, ", ")

	// 谱面数据 (仅针对当前 Chart)
	diff := chart.FitDiff - chart.DS
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	chartInfoStr := fmt.Sprintf("[%s] DS: %.1f, Fit: %.2f (Diff: %s%.2f)",
		chart.Difficulty, chart.DS, chart.FitDiff, sign, diff)

	// 2. 运行分析师 (Analyst)
	// 注意：这里复用了 runAnalyst，它会使用通用的 Analyst Prompt。
	// 理想情况下，我们可以为 Chart Analysis 定制一个 Prompt，但目前通用 Prompt 已经包含了 ChartInfo 和 Version 区分指令，只要我们传入的 comments 是纯净的（或者大部分是针对该谱面的），效果应该不错。
	analystOut, reasoning, err := a.runAnalyst(ctx, chunkStr, termGuide, aliasStr, chartInfoStr)
	if err != nil {
		return fmt.Errorf("分析师运行失败: %w", err)
	}

	// 3. 运行顾问 (Advisor) - 生成针对该谱面的建议
	// 同样复用 runAdvisor，但我们需要让 Advisor 知道这是针对特定谱面的
	// 我们可以临时修改 song.Title 或者在 analystData 中注入信息，但最稳妥的是修改 runAdvisor 接口或 Prompt。这里我们简单地将 Chart 信息附加到 AnalystOutput 的 Summary 中，或者依赖 Advisor 自己从 AnalysisData 中读取
	// 为了简单起见，我们直接使用 runAdvisor，它会生成一份报告。
	advisorOut, err := a.runAdvisor(ctx, song, analystOut)
	if err != nil {
		return fmt.Errorf("顾问运行失败: %w", err)
	}

	// 4. 保存结果 (TargetType = "chart")
	result := &model.AnalysisResult{
		TargetType:         "chart",
		TargetID:           chart.ID,
		Summary:            advisorOut.Summary,
		RatingAdvice:       advisorOut.RatingAdvice,
		DifficultyAnalysis: advisorOut.DifficultyAnalysis,
		ReasoningLog:       reasoning,
	}

	if err := a.storage.CreateAnalysisResult(result); err != nil {
		return fmt.Errorf("保存谱面分析结果失败: %w", err)
	}

	return nil
}

func (a *Analyzer) formatChartInfo(charts []model.Chart) string {
	var infos []string
	for _, c := range charts {
		// 只关注 Expert, Master, Re:Master
		if c.Difficulty != "Expert" && c.Difficulty != "Master" && c.Difficulty != "Re:Master" {
			continue
		}

		diff := c.FitDiff - c.DS
		sign := "+"
		if diff < 0 {
			sign = ""
		}

		// 格式: [DX Master] DS: 13.5, Fit: 13.8 (Diff: +0.3)
		// 注意: Chart 模型中没有直接存储 DX/Std 标记，通常需要从 Song.Type 或 Chart 属性推断，但这里我们假设 Chart 列表已经包含了所有版本。如果 Chart 模型本身没有区分 DX/Std 的字段，
		// 我们可能只能显示难度。根据 model.Song 定义，Type 是在 Song 上的。如果一首歌同时有 DX 和 Std 谱面，通常在 Diving-Fish API 中是作为两个不同的 Song 对象存在的。
		// 所以这里直接用 c.Difficulty 即可。

		info := fmt.Sprintf("[%s] DS: %.1f, Fit: %.2f (Diff: %s%.2f)",
			c.Difficulty, c.DS, c.FitDiff, sign, diff)
		infos = append(infos, info)
	}
	if len(infos) == 0 {
		return "暂无高难度谱面数据"
	}
	return strings.Join(infos, "; ")
}

func (a *Analyzer) mergeAnalystOutputs(outputs []*AnalystOutput) *AnalystOutput {
	merged := &AnalystOutput{
		DifficultyTags: []string{},
		KeyPatterns:    []string{},
		Pros:           []string{},
		Cons:           []string{},
		Sentiment:      "Neutral",
	}

	seenTags := make(map[string]bool)
	seenPatterns := make(map[string]bool)
	seenPros := make(map[string]bool)
	seenCons := make(map[string]bool)

	for _, out := range outputs {
		for _, tag := range out.DifficultyTags {
			if !seenTags[tag] {
				merged.DifficultyTags = append(merged.DifficultyTags, tag)
				seenTags[tag] = true
			}
		}
		for _, pat := range out.KeyPatterns {
			if !seenPatterns[pat] {
				merged.KeyPatterns = append(merged.KeyPatterns, pat)
				seenPatterns[pat] = true
			}
		}
		for _, pro := range out.Pros {
			if !seenPros[pro] {
				merged.Pros = append(merged.Pros, pro)
				seenPros[pro] = true
			}
		}
		for _, con := range out.Cons {
			if !seenCons[con] {
				merged.Cons = append(merged.Cons, con)
				seenCons[con] = true
			}
		}
		if out.Sentiment != "Neutral" {
			merged.Sentiment = out.Sentiment
		}
	}
	return merged
}
