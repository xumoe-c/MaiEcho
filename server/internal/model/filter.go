package model

// SongFilter 定义了歌曲查询的过滤条件
type SongFilter struct {
	Version  string  `form:"version"`
	MinDS    float64 `form:"min_ds"`
	MaxDS    float64 `form:"max_ds"`
	Type     string  `form:"type"` // DX or Standard
	Genre    string  `form:"genre"`
	IsNew    *bool   `form:"is_new"`
	Keyword  string  `form:"keyword"`
	Page     int     `form:"page,default=1"`
	PageSize int     `form:"page_size,default=20"`
}

// SongListResponse 定义了歌曲列表的返回结构
type SongListResponse struct {
	Total int64  `json:"total"`
	Items []Song `json:"items"`
}
