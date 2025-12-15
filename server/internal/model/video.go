package model

import (
	"time"

	"gorm.io/gorm"
)

// Video 代表来自外部平台的音乐相关视频
type Video struct {
	gorm.Model
	Source      string    `json:"source"`                         // 例如： "Bilibili"
	ExternalID  string    `gorm:"uniqueIndex" json:"external_id"` // 例如： "BV..."
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	PublishTime time.Time `json:"publish_time"`
}
