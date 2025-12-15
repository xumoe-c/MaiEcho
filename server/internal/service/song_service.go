package service

import (
	"fmt"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/provider/divingfish"
	"github.com/xumoe-c/maiecho/server/internal/provider/yuzuchan"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type SongService interface {
	GetSong(id uint) (*model.Song, error)
	GetSongByGameID(gameID int) (*model.Song, error)
	CreateSong(song *model.Song) error
	SyncFromDivingFish() error
	RefreshAliases() error
	GetAllSongs() ([]model.Song, error)
	GetSongs(filter model.SongFilter) (*model.SongListResponse, error)
}

type songServiceImpl struct {
	storage          storage.Storage
	divingFishClient *divingfish.Client
	yuzuChanClient   *yuzuchan.Client
}

func NewSongService(s storage.Storage, dfClient *divingfish.Client, yzClient *yuzuchan.Client) SongService {
	return &songServiceImpl{
		storage:          s,
		divingFishClient: dfClient,
		yuzuChanClient:   yzClient,
	}
}

func (s *songServiceImpl) GetAllSongs() ([]model.Song, error) {
	return s.storage.GetAllSongs()
}

func (s *songServiceImpl) GetSongs(filter model.SongFilter) (*model.SongListResponse, error) {
	songs, total, err := s.storage.GetSongs(filter)
	if err != nil {
		return nil, err
	}
	return &model.SongListResponse{
		Total: total,
		Items: songs,
	}, nil
}

func (s *songServiceImpl) GetSong(id uint) (*model.Song, error) {
	return s.storage.GetSong(id)
}

func (s *songServiceImpl) GetSongByGameID(gameID int) (*model.Song, error) {
	return s.storage.GetSongByGameID(gameID)
}

func (s *songServiceImpl) CreateSong(song *model.Song) error {
	return s.storage.CreateSong(song)
}

func (s *songServiceImpl) SyncFromDivingFish() error {
	songs, err := s.divingFishClient.FetchSongs()
	if err != nil {
		return fmt.Errorf("从 diving-fish 获取歌曲失败: %w", err)
	}

	if len(songs) == 0 {
		// 304 Not Modified or just empty
		return nil
	}

	logger.Info("导入歌曲", "module", "service.song", "count", len(songs))
	for _, song := range songs {
		if err := s.storage.UpsertSong(&song); err != nil {
			logger.Error("保存歌曲失败", "module", "service.song", "title", song.Title, "error", err)
		}
	}

	return nil
}

func (s *songServiceImpl) RefreshAliases() error {
	logger.Info("开始刷新别名", "module", "service.song")
	songs, err := s.storage.GetAllSongs()
	if err != nil {
		return fmt.Errorf("获取歌曲失败: %w", err)
	}

	logger.Info("找到需要更新别名的歌曲", "module", "service.song", "count", len(songs))

	count := 0
	for i, song := range songs {
		// 避免过快请求
		if i > 0 {
			time.Sleep(200 * time.Millisecond)
		}

		aliasItem, err := s.yuzuChanClient.FetchAliasBySongID(song.GameID)
		if err != nil {
			// 记录错误但继续
			logger.Error("获取别名失败", "module", "service.song", "gameID", song.GameID, "title", song.Title, "error", err)
			continue
		}

		if aliasItem == nil {
			continue
		}

		if len(aliasItem.Alias) > 0 {
			if err := s.storage.SaveSongAliases(song.ID, aliasItem.Alias); err != nil {
				logger.Error("保存别名失败", "module", "service.song", "songID", song.ID, "error", err)
			} else {
				count++
			}
		}

		// 每处理50首歌曲记录一次进度
		if (i+1)%50 == 0 {
			logger.Info("别名刷新进度", "module", "service.song", "processed", i+1, "total", len(songs), "updated", count)
		}
	}

	logger.Info("已更新歌曲别名", "module", "service.song", "count", count)
	return nil
}
