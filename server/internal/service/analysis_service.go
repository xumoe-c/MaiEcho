package service

import (
	"context"

	"github.com/xumoe-c/maiecho/server/internal/agent"
	"github.com/xumoe-c/maiecho/server/internal/config"
	"github.com/xumoe-c/maiecho/server/internal/llm"
	"github.com/xumoe-c/maiecho/server/internal/model"
	"github.com/xumoe-c/maiecho/server/internal/storage"
)

type AnalysisService struct {
	analyzer *agent.Analyzer
	storage  storage.Storage
}

func NewAnalysisService(s storage.Storage, client *llm.Client, prompts *config.PromptConfig) *AnalysisService {
	return &AnalysisService{
		analyzer: agent.NewAnalyzer(s, client, prompts),
		storage:  s,
	}
}

func (s *AnalysisService) AnalyzeSong(ctx context.Context, songID uint) error {
	return s.analyzer.AnalyzeSong(ctx, songID)
}

func (s *AnalysisService) AnalyzeSongByGameID(ctx context.Context, gameID int) error {
	song, err := s.storage.GetSongByGameID(gameID)
	if err != nil {
		return err
	}
	return s.analyzer.AnalyzeSong(ctx, song.ID)
}

// AggregatedAnalysisResult 聚合了歌曲和谱面的分析结果
type AggregatedAnalysisResult struct {
	SongResult   *model.AnalysisResult   `json:"song_result"`
	ChartResults []*model.AnalysisResult `json:"chart_results"`
}

func (s *AnalysisService) GetAggregatedAnalysisResultByGameID(gameID int) (*AggregatedAnalysisResult, error) {
	song, err := s.storage.GetSongByGameID(gameID)
	if err != nil {
		return nil, err
	}

	// 获取歌曲级别的分析结果
	songResult, err := s.storage.GetAnalysisResultBySongID(song.ID)
	if err != nil {
		// 如果没有歌曲级别的结果，可能还没有分析过，或者只有谱面级别的结果（虽然不太可能）
		// 这里我们允许 songResult 为 nil，或者返回错误
		// return nil, err
	}

	// 获取该歌曲下所有谱面的分析结果
	var chartResults []*model.AnalysisResult
	for _, chart := range song.Charts {
		res, err := s.storage.GetAnalysisResultsByTarget("chart", chart.ID)
		if err == nil && res != nil {
			chartResults = append(chartResults, res)
		}
	}

	return &AggregatedAnalysisResult{
		SongResult:   songResult,
		ChartResults: chartResults,
	}, nil
}

func (s *AnalysisService) GetAnalysisResult(songID uint) (*model.AnalysisResult, error) {
	return s.storage.GetAnalysisResultBySongID(songID)
}

func (s *AnalysisService) GetAnalysisResultByGameID(gameID int) (*model.AnalysisResult, error) {
	song, err := s.storage.GetSongByGameID(gameID)
	if err != nil {
		return nil, err
	}
	return s.storage.GetAnalysisResultBySongID(song.ID)
}
