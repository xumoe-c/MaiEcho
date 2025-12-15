package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type RelevanceAnalyzer struct {
	llm     *llm.Client
	prompts *config.PromptConfig
}

func NewRelevanceAnalyzer(llm *llm.Client, prompts *config.PromptConfig) *RelevanceAnalyzer {
	return &RelevanceAnalyzer{
		llm:     llm,
		prompts: prompts,
	}
}

type AliasCheckResult struct {
	IsSuitable bool   `json:"is_suitable"`
	Reason     string `json:"reason"`
}

func (a *RelevanceAnalyzer) CheckAliasSuitability(ctx context.Context, title, artist, alias string) (bool, error) {
	// 简单的预过滤：如果别名太短，直接返回 false，节省 token
	if len(alias) < 2 {
		return false, nil
	}

	prompt := a.prompts.Agent.Relevance.CheckAlias
	userPrompt := strings.ReplaceAll(prompt.User, "{{.Title}}", title)
	userPrompt = strings.ReplaceAll(userPrompt, "{{.Artist}}", artist)
	userPrompt = strings.ReplaceAll(userPrompt, "{{.Alias}}", alias)

	resp, err := a.llm.Chat(ctx, prompt.System, userPrompt)
	if err != nil {
		return false, err
	}

	// 清理可能的 markdown 标记
	resp = strings.TrimPrefix(resp, "```json")
	resp = strings.TrimPrefix(resp, "```")
	resp = strings.TrimSpace(resp)

	var result AliasCheckResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		logger.Error("解析别名检查结果失败", "module", "agent.relevance", "response", resp, "error", err)
		return false, err
	}

	logger.Info("别名检查", "module", "agent.relevance", "alias", alias, "suitable", result.IsSuitable, "reason", result.Reason)
	return result.IsSuitable, nil
}

type TitleCheckResult struct {
	IsRelevant bool    `json:"is_relevant"`
	Confidence float64 `json:"confidence"`
}

func (a *RelevanceAnalyzer) CheckTitleRelevance(ctx context.Context, title, artist string, aliases []string, videoTitle string) (bool, error) {
	prompt := a.prompts.Agent.Relevance.CheckTitle

	aliasesStr := fmt.Sprintf("%v", aliases)

	userPrompt := strings.ReplaceAll(prompt.User, "{{.Title}}", title)
	userPrompt = strings.ReplaceAll(userPrompt, "{{.Artist}}", artist)
	userPrompt = strings.ReplaceAll(userPrompt, "{{.Aliases}}", aliasesStr)
	userPrompt = strings.ReplaceAll(userPrompt, "{{.VideoTitle}}", videoTitle)

	resp, err := a.llm.Chat(ctx, prompt.System, userPrompt)
	if err != nil {
		return false, err
	}

	resp = strings.TrimPrefix(resp, "```json")
	resp = strings.TrimPrefix(resp, "```")
	resp = strings.TrimSpace(resp)

	var result TitleCheckResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		logger.Error("解析标题相关性检查结果失败", "module", "agent.relevance", "response", resp, "error", err)
		return false, err
	}

	return result.IsRelevant, nil
}
