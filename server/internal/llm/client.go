package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(cfg config.LLMConfig) *Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)

	return &Client{
		client: &client,
		model:  cfg.Model,
	}
}

func (c *Client) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	chatCompletion, err := c.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
			Model: c.model,
		},
	)

	if err != nil {
		logger.Error("LLM请求失败", "module", "llm", "model", c.model, "error", err)
		logger.LogLLMConversation(c.model, systemPrompt, userPrompt, "", err)
		return "", fmt.Errorf("LLM请求失败: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		logger.Error("LLM未返回任何选项", "module", "llm", "model", c.model)
		logger.LogLLMConversation(c.model, systemPrompt, userPrompt, "", fmt.Errorf("no choices returned"))
		return "", fmt.Errorf("LLM未返回任何选项")
	}

	response := chatCompletion.Choices[0].Message.Content
	logger.LogLLMConversation(c.model, systemPrompt, userPrompt, response, nil)
	return response, nil
}

// ChatWithReasoning 执行对话并分离推理过程 (<thinking>标签) 和最终内容
func (c *Client) ChatWithReasoning(ctx context.Context, systemPrompt, userPrompt string) (content string, reasoning string, err error) {
	fullResponse, err := c.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		logger.Error("LLM对话失败", "module", "llm", "error", err)
		return "", "", err
	}

	// 提取 <thinking> 内容
	re := regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	matches := re.FindStringSubmatch(fullResponse)

	if len(matches) > 1 {
		reasoning = strings.TrimSpace(matches[1])
		// 从响应中移除 <thinking> 部分，只保留剩下的内容（通常是 JSON）
		content = strings.TrimSpace(re.ReplaceAllString(fullResponse, ""))
		logger.Info("成功提取推理内容", "module", "llm", "reasoningLength", len(reasoning))
	} else {
		// 如果没有找到标签，假设全部是内容
		content = fullResponse
		logger.Debug("LLM响应中未找到thinking标签", "module", "llm")
	}

	return content, reasoning, nil
}
