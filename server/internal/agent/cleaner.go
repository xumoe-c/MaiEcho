package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type Cleaner struct {
	htmlTagRegex    *regexp.Regexp
	noiseKeywords   []string
	validShortTerms []string
	llm             *llm.Client
	prompts         *config.PromptConfig
}

func NewCleaner(llmClient *llm.Client, prompts *config.PromptConfig) *Cleaner {
	return &Cleaner{
		htmlTagRegex: regexp.MustCompile(`<[^>]*>`),
		// 噪音关键词：与谱面分析无关的内容
		noiseKeywords: []string{
			"拼车", "排队", "机况", "出勤", "打卡", "机器", "按键", "屏幕",
			"第一", "前排", "沙发", "打卡", "围观", "吃瓜",
			"求好友", "互粉", "扩列", "佬", "?", "不可发送单个标点符号", "258元回答你的问题",
		},
		// 有效短语白名单：即使长度很短也保留
		validShortTerms: []string{
			"AP", "FC", "SSS", "SSS+", "FDX", "FDX+", "鸟", "鸟加", "全连", "收了", "理论值",
			"越级", "诈称", "逆诈称", "手癖", "局所难", "个人差",
		},
		llm:     llmClient,
		prompts: prompts,
	}
}

// Clean 移除 HTML 标签、多余的空白字符，并过滤掉非常短的内容
func (c *Cleaner) Clean(content string) string {
	// 1. 移除 HTML 标签
	content = c.htmlTagRegex.ReplaceAllString(content, "")

	// 2. 将多个空格/换行符替换为单个空格
	content = strings.Join(strings.Fields(content), " ")

	// 3. 去除首尾空白
	content = strings.TrimSpace(content)

	return content
}

// IsValid 检查内容是否值得分析
func (c *Cleaner) IsValid(content string) bool {
	if content == "" {
		return false
	}

	// 1. 检查噪音关键词
	for _, keyword := range c.noiseKeywords {
		if strings.Contains(content, keyword) {
			return false
		}
	}

	// 2. 检查长度
	runeCount := utf8.RuneCountInString(content)

	// 如果包含有效术语，允许短内容
	if runeCount < 5 {
		upperContent := strings.ToUpper(content)
		for _, term := range c.validShortTerms {
			if strings.Contains(upperContent, term) {
				return true
			}
		}
		return false
	}

	return true
}

// CleanWithLLM 使用 LLM 进行语义清洗
func (c *Cleaner) CleanWithLLM(ctx context.Context, comments []string) ([]string, error) {
	if len(comments) == 0 {
		return []string{}, nil
	}

	// 构造 Prompt
	tmpl, err := template.New("cleaner").Parse(c.prompts.Agent.Cleaner.User)
	if err != nil {
		return nil, fmt.Errorf("解析 Cleaner Prompt 失败: %w", err)
	}

	// 将评论列表转换为带序号的字符串，方便 LLM 理解
	var commentListBuilder strings.Builder
	for i, comment := range comments {
		commentListBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, comment))
	}

	var promptBody bytes.Buffer
	if err := tmpl.Execute(&promptBody, map[string]interface{}{
		"Comments": commentListBuilder.String(),
	}); err != nil {
		return nil, fmt.Errorf("执行 Cleaner Prompt 失败: %w", err)
	}

	// 调用 LLM
	// 注意：Cleaner 不需要推理过程，只需要结果
	response, err := c.llm.Chat(ctx, c.prompts.Agent.Cleaner.System, promptBody.String())
	if err != nil {
		return nil, fmt.Errorf("LLM 清洗请求失败: %w", err)
	}

	// 解析 JSON
	var validComments []string
	// 尝试清理 Markdown 代码块标记 (```json ... ```)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	if err := json.Unmarshal([]byte(response), &validComments); err != nil {
		logger.Error("LLM 清洗结果解析失败，降级处理", "module", "agent.cleaner", "error", err, "responseLength", len(response))
		// 降级：如果解析失败，返回原始列表的前 50% (假设前排评论质量较高) 或者全部返回
		// 这里为了安全起见，返回空列表让上层处理，或者返回原始列表
		return comments, nil // 降级为不清洗
	}

	return validComments, nil
}
