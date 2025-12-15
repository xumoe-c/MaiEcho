package agent

// AnalystOutput 代表从评论中提取的客观事实
type AnalystOutput struct {
	DifficultyTags  []string `json:"difficulty_tags"` // 例如 "13+", "体力谱"
	KeyPatterns     []string `json:"key_patterns"`    // 例如 "纵连", "流星雨"
	Pros            []string `json:"pros"`
	Cons            []string `json:"cons"`
	Sentiment       string   `json:"sentiment"`        // "Positive", "Neutral", "Negative"
	VersionAnalysis string   `json:"version_analysis"` // 针对不同版本/难度的特定分析
}

// AdvisorOutput 代表最终建议
type AdvisorOutput struct {
	Summary            string `json:"summary"`
	RatingAdvice       string `json:"rating_advice"`
	DifficultyAnalysis string `json:"difficulty_analysis"`
}
