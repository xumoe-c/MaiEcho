package model

import (
	"gorm.io/gorm"
)

// Song 代表音乐游戏中的一首歌曲
type Song struct {
	gorm.Model
	GameID      int         `gorm:"uniqueIndex" json:"id"` // 来自 Diving-Fish 的 ID
	Title       string      `gorm:"index" json:"title"`
	Type        string      `json:"type"` // DX or 标准
	Artist      string      `json:"artist"`
	Genre       string      `json:"genre"`
	BPM         float64     `json:"bpm"`
	ReleaseDate string      `json:"release_date"`
	Version     string      `json:"version"` // 乐曲更新版本
	IsNew       bool        `json:"is_new"`
	CoverURL    string      `json:"cover_url"`
	LastScraped *string     `json:"last_scraped"` // 上次采集时间 (ISO8601 字符串或时间戳)
	Charts      []Chart     `json:"charts,omitempty"`
	Aliases     []SongAlias `json:"aliases,omitempty"`
}

type SongAlias struct {
	gorm.Model
	SongID     uint   `gorm:"index" json:"song_id"`
	Alias      string `gorm:"index" json:"alias"`
	IsSuitable *bool  `json:"is_suitable"` // nil: unchecked, true: suitable, false: unsuitable
}

// Chart 代表歌曲的特定难度谱面
type Chart struct {
	gorm.Model
	SongID         uint    `gorm:"index" json:"song_id"`
	Difficulty     string  `json:"difficulty"` // Basic, Advanced, Expert, Master, Re:Master
	Level          string  `json:"level"`      // 13, 13+, 14, etc.
	DS             float64 `json:"ds"`         // Internal decimal level
	Notes          string  `json:"notes"`      // JSON array: [tap, hold, slide, touch, break]
	Charter        string  `json:"charter"`    // NotesDesigner
	FitDiff        float64 `json:"fit_diff"`
	AvgAchievement float64 `json:"avg_achievement"`
	AvgDX          float64 `json:"avg_dx"`
	StdDev         float64 `json:"std_dev"`
	SampleCount    int     `json:"sample_count"`
	Song           Song    `json:"-"`
}
