package model

import (
	"gorm.io/gorm"
)

// AnalysisResult 存储对歌曲或谱面的分析结果
type AnalysisResult struct {
	gorm.Model
	TargetType         string `gorm:"index" json:"target_type"` // Song or Chart
	TargetID           uint   `gorm:"index" json:"target_id"`
	Summary            string `json:"summary"`
	RatingAdvice       string `json:"rating_advice"`
	DifficultyAnalysis string `json:"difficulty_analysis"`
	ReasoningLog       string `json:"reasoning_log" gorm:"type:text"` // 存储 LLM 的推理过程
}
