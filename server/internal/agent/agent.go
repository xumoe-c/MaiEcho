package agent

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/xumoe-c/maiecho/server/internal/logger"
)

// Agent 定义了分析智能体的接口
type Agent interface {
	Analyze(input string) (string, error)
}

// ExecuteTemplate 执行简单的文本模板
func ExecuteTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		logger.Error("解析模板失败", "module", "agent", "error", err)
		return "", fmt.Errorf("解析模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}
	return buf.String(), nil
}
