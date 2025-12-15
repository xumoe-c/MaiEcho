package agent

import (
	"strings"

	"github.com/xumoe-c/maiecho/server/internal/config"
)

// KnowledgeBase 存储音游术语及其解释
type KnowledgeBase struct {
	terms       map[string]string
	guideHeader string
}

// NewKnowledgeBase 使用常用术语初始化知识库
func NewKnowledgeBase(cfg *config.PromptConfig) *KnowledgeBase {
	header := "\n术语指南:\n"
	if cfg != nil && cfg.Agent.Knowledge.GuideHeader != "" {
		header = cfg.Agent.Knowledge.GuideHeader
	}

	return &KnowledgeBase{
		guideHeader: header,
		terms: map[string]string{
			// 评价类
			"鸟":   "SSS评价 (100.0000% - 100.4999%)",
			"鸟加":  "SSS+评价 (100.5000%及以上)",
			"AP":  "All Perfect (所有Note均为Perfect判定)",
			"FC":  "Full Combo (全连)",
			"理论值": "101.0000% (所有Note均为Perfect，其中所有Break Note均为Critical判定)",
			"越级":  "挑战超过自己当前水平的谱面",
			"诈称":  "实际难度高于标定等级",
			"逆诈称": "实际难度低于标定等级",
			"手癖":  "由于错误的肌肉记忆导致的习惯性失误",

			// 谱面配置类
			"纵连":  "连续的纵向按键配置",
			"交互":  "左右手交替点击",
			"海底潭": "指乐曲《海底潭》的红谱，以特定的配置闻名",
			"流星雨": "指乐曲《PANDORA PARADOXXX》中的一段密集Note下落",
			"转圈":  "需要沿着屏幕边缘滑动的Slide或Tap配置",
			"蹭键":  "利用判定区特性，用非正规的手法触发Note",
			"糊":   "指玩家看不清谱面，乱拍",
			"底力":  "玩家的基础实力（如读谱速度、手速、耐力）",
			"位移":  "需要身体或手部大幅度移动的配置",
			"出张":  "手部跨越到屏幕另一侧去处理Note",
		},
	}
}

// GetRelevantTerms 返回在内容中找到的术语及其解释
func (kb *KnowledgeBase) GetRelevantTerms(content string) map[string]string {
	relevant := make(map[string]string)
	for term, explanation := range kb.terms {
		if strings.Contains(content, term) {
			relevant[term] = explanation
		}
	}
	return relevant
}

// FormatTerms 将相关术语格式化为字符串以供 Prompt 使用
func (kb *KnowledgeBase) FormatTerms(terms map[string]string) string {
	if len(terms) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(kb.guideHeader)
	for term, explanation := range terms {
		sb.WriteString("- " + term + ": " + explanation + "\n")
	}
	return sb.String()
}
