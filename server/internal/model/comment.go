package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 表示与歌曲或谱面相关的评论
type Comment struct {
	gorm.Model
	Source      string    `gorm:"index" json:"source"` // Bilibili, Tieba, etc.
	SourceTitle string    `json:"source_title"`        // Title of the source (e.g. Video Title)
	ExternalID  string    `gorm:"index" json:"external_id"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	PostDate    time.Time `json:"post_date"`
	SongID      *uint     `gorm:"index" json:"song_id,omitempty"`
	ChartID     *uint     `gorm:"index" json:"chart_id,omitempty"`
	SearchTag   string    `gorm:"index" json:"search_tag"` // The keyword used to find this comment
	Sentiment   float64   `json:"sentiment"`               // -1.0 to 1.0
}
