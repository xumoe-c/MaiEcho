package storage

import (
	"time"

	"github.com/xumoe-c/maiecho/server/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	err = db.AutoMigrate(
		&model.Song{},
		&model.Chart{},
		&model.SongAlias{},
		&model.Comment{},
		&model.AnalysisResult{},
		&model.Video{},
	)
	if err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

// 实现基本的 CRUD 操作，可以放在这里或单独的文件中
func (d *Database) CreateSong(song *model.Song) error {
	return d.DB.Create(song).Error
}

func (d *Database) UpsertSong(song *model.Song) error {
	var existing model.Song
	// 使用 Find 替代 First 避免 "record not found" 日志
	result := d.DB.Where("game_id = ?", song.GameID).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return d.DB.Create(song).Error
	}

	// 更新已有记录
	song.ID = existing.ID

	return d.DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧的谱面，确保没有重复或过时的数据
		if err := tx.Where("song_id = ?", existing.ID).Delete(&model.Chart{}).Error; err != nil {
			return err
		}
		// 保存歌曲（更新字段并插入新的谱面）
		return tx.Save(song).Error
	})
}

func (d *Database) GetSong(id uint) (*model.Song, error) {
	var song model.Song
	err := d.DB.Preload("Charts").Preload("Aliases").First(&song, id).Error
	return &song, err
}

func (d *Database) GetSongByGameID(gameID int) (*model.Song, error) {
	var song model.Song
	err := d.DB.Preload("Charts").Preload("Aliases").Where("game_id = ?", gameID).First(&song).Error
	return &song, err
}

func (d *Database) GetAllSongs() ([]model.Song, error) {
	var songs []model.Song
	err := d.DB.Preload("Aliases").Find(&songs).Error
	return songs, err
}

func (d *Database) GetSongs(filter model.SongFilter) ([]model.Song, int64, error) {
	var songs []model.Song
	var total int64

	query := d.DB.Model(&model.Song{})

	if filter.Version != "" {
		query = query.Where("version = ?", filter.Version)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Genre != "" {
		query = query.Where("genre = ?", filter.Genre)
	}
	if filter.IsNew != nil {
		query = query.Where("is_new = ?", *filter.IsNew)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		// Search in Title, Artist OR Aliases
		// 使用 Group 条件来确保 OR 逻辑正确，并且正确处理子查询
		query = query.Where(
			d.DB.Where("title LIKE ?", keyword).
				Or("artist LIKE ?", keyword).
				Or("id IN (?)", d.DB.Model(&model.SongAlias{}).Select("song_id").Where("alias LIKE ?", keyword)),
		)
	}

	// 谱面定数筛选需要关联查询，这里简化处理，如果需要精确筛选可能需要 Join Chart 表
	// 假设筛选的是该歌曲任意谱面满足条件
	if filter.MinDS > 0 || filter.MaxDS > 0 {
		subQuery := d.DB.Model(&model.Chart{}).Select("song_id")
		if filter.MinDS > 0 {
			subQuery = subQuery.Where("ds >= ?", filter.MinDS)
		}
		if filter.MaxDS > 0 {
			subQuery = subQuery.Where("ds <= ?", filter.MaxDS)
		}
		query = query.Where("id IN (?)", subQuery)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Preload("Charts").Preload("Aliases").
		Offset(offset).Limit(filter.PageSize).
		Find(&songs).Error

	return songs, total, err
}

func (d *Database) SaveSongAliases(songID uint, aliases []string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧的别名
		if err := tx.Where("song_id = ?", songID).Delete(&model.SongAlias{}).Error; err != nil {
			return err
		}

		// 插入新的别名
		for _, alias := range aliases {
			if err := tx.Create(&model.SongAlias{
				SongID: songID,
				Alias:  alias,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *Database) CreateComment(comment *model.Comment) error {
	return d.DB.Create(comment).Error
}

func (d *Database) UpdateComment(comment *model.Comment) error {
	return d.DB.Save(comment).Error
}

func (d *Database) GetCommentsByKeyword(keyword string) ([]model.Comment, error) {
	var comments []model.Comment
	// 使用 LIKE 查询内容和来源标题
	// TODO: 考虑全文索引以提高性能
	err := d.DB.Where("content LIKE ? OR source_title LIKE ?", "%"+keyword+"%", "%"+keyword+"%").Find(&comments).Error
	return comments, err
}

func (d *Database) GetCommentsBySongID(songID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := d.DB.Where("song_id = ?", songID).Find(&comments).Error
	return comments, err
}

func (d *Database) CreateAnalysisResult(result *model.AnalysisResult) error {
	return d.DB.Create(result).Error
}

func (d *Database) GetAnalysisResultBySongID(songID uint) (*model.AnalysisResult, error) {
	return d.GetAnalysisResultsByTarget("song", songID)
}

func (d *Database) GetAnalysisResultsByTarget(targetType string, targetID uint) (*model.AnalysisResult, error) {
	var result model.AnalysisResult
	err := d.DB.Where("target_type = ? AND target_id = ?", targetType, targetID).Order("created_at desc").First(&result).Error
	return &result, err
}

func (d *Database) CreateVideo(video *model.Video) error {
	// 使用 Clauses 处理潜在的重复（例如忽略或更新）
	// 目前，我们只是忽略如果存在以避免错误，或者使用 FirstOrCreate 逻辑
	var existing model.Video
	result := d.DB.Where("external_id = ?", video.ExternalID).First(&existing)
	if result.RowsAffected > 0 {
		// 如果需要，更新字段，或者直接返回 nil
		video.ID = existing.ID
		return d.DB.Save(video).Error
	}
	return d.DB.Create(video).Error
}

func (d *Database) UpdateSongLastScrapedTime(songID uint) error {
	now := time.Now().Format(time.RFC3339)
	return d.DB.Model(&model.Song{}).Where("id = ?", songID).Update("last_scraped", now).Error
}

func (d *Database) UpdateSongAliasSuitability(aliasID uint, isSuitable bool) error {
	return d.DB.Model(&model.SongAlias{}).Where("id = ?", aliasID).Update("is_suitable", isSuitable).Error
}
