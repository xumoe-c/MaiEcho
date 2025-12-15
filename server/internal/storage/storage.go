package storage

import "github.com/xumoe-c/maiecho/server/internal/model"

// Storage 定义了存储接口
type Storage interface {
	CreateSong(song *model.Song) error
	UpsertSong(song *model.Song) error
	GetSong(id uint) (*model.Song, error)
	GetSongByGameID(gameID int) (*model.Song, error)
	GetAllSongs() ([]model.Song, error)
	GetSongs(filter model.SongFilter) ([]model.Song, int64, error)
	SaveSongAliases(songID uint, aliases []string) error
	CreateComment(comment *model.Comment) error
	UpdateComment(comment *model.Comment) error
	GetCommentsByKeyword(keyword string) ([]model.Comment, error)
	GetCommentsBySongID(songID uint) ([]model.Comment, error)
	CreateAnalysisResult(result *model.AnalysisResult) error
	GetAnalysisResultBySongID(songID uint) (*model.AnalysisResult, error)
	GetAnalysisResultsByTarget(targetType string, targetID uint) (*model.AnalysisResult, error)
	CreateVideo(video *model.Video) error
	UpdateSongAliasSuitability(aliasID uint, isSuitable bool) error
}
