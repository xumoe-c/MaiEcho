package agent

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type Mapper struct {
	storage storage.Storage
	llm     *llm.Client
	prompts *config.PromptConfig
}

func NewMapper(s storage.Storage, l *llm.Client, prompts *config.PromptConfig) *Mapper {
	return &Mapper{
		storage: s,
		llm:     l,
		prompts: prompts,
	}
}

// MapCommentsToSongs 触发评论到歌曲的映射过程
func (m *Mapper) MapCommentsToSongs(ctx context.Context) error {
	songs, err := m.storage.GetAllSongs()
	if err != nil {
		return fmt.Errorf("获取歌曲失败: %w", err)
	}

	logger.Info("开始进行评论到歌曲的映射", "module", "agent.mapper", "songCount", len(songs))

	count := 0
	for _, song := range songs {
		// 收集所有关键词以进行搜索（标题 + 别名）
		keywords := []string{song.Title}
		for _, alias := range song.Aliases {
			if alias.Alias != "" {
				keywords = append(keywords, alias.Alias)
			}
		}

		// 关键词去重
		uniqueKeywords := make(map[string]bool)
		var searchList []string
		for _, k := range keywords {
			if _, exists := uniqueKeywords[k]; !exists {
				uniqueKeywords[k] = true
				// 仅添加长度 >=2 的关键词以避免过多噪音
				if utf8.RuneCountInString(k) >= 2 {
					searchList = append(searchList, k)
				}
			}
		}

		for _, keyword := range searchList {
			comments, err := m.storage.GetCommentsByKeyword(keyword)
			if err != nil {
				logger.Error("获取关键词的评论失败", "module", "agent.mapper", "keyword", keyword, "error", err)
				continue
			}

			for i := range comments {
				comment := &comments[i]

				// 如果已经关联了歌曲，跳过
				if comment.SongID != nil {
					continue
				}

				// 简单匹配标题或内容
				match := false
				if strings.Contains(strings.ToLower(comment.Content), strings.ToLower(keyword)) {
					match = true
				}
				if strings.Contains(strings.ToLower(comment.SourceTitle), strings.ToLower(keyword)) {
					match = true
				}

				if match {
					// 对于短关键词或别名，使用 LLM 进行验证
					isRelevant := true
					// 如果关键词较短或是别名（别名可能也有歧义），则进行验证。目前我们还是坚持长度规则。
					if utf8.RuneCountInString(keyword) <= 4 && m.llm != nil {
						relevant, err := m.verifyMatchWithLLM(ctx, keyword, comment)
						if err != nil {
							logger.Error("LLM 验证评论失败", "module", "agent.mapper", "commentID", comment.ID, "error", err)
							isRelevant = false
						} else {
							isRelevant = relevant
						}
					}

					if isRelevant {
						sID := song.ID
						comment.SongID = &sID
						if err := m.storage.UpdateComment(comment); err != nil {
							logger.Error("关联评论到歌曲失败", "module", "agent.mapper", "commentID", comment.ID, "songTitle", song.Title, "error", err)
						} else {
							count++
						}
					}
				}
			}
		}
	}
	logger.Info("评论到歌曲的映射完成", "module", "agent.mapper", "associatedCount", count)
	return nil
}

func (m *Mapper) verifyMatchWithLLM(ctx context.Context, keyword string, comment *model.Comment) (bool, error) {
	systemPrompt := m.prompts.Agent.Mapper.VerifyMatch.System

	data := struct {
		Keyword     string
		SourceTitle string
		Content     string
	}{
		Keyword:     keyword,
		SourceTitle: comment.SourceTitle,
		Content:     comment.Content,
	}

	userPrompt, err := ExecuteTemplate(m.prompts.Agent.Mapper.VerifyMatch.User, data)
	if err != nil {
		return false, fmt.Errorf("执行用户提示模板失败: %w", err)
	}

	resp, err := m.llm.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return false, err
	}

	cleaned := strings.TrimSpace(strings.ToUpper(resp))
	return strings.Contains(cleaned, "YES"), nil
}
