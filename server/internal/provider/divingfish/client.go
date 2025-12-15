package divingfish

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/logger"
	"github.com/xumoe-c/maiecho/server/internal/model"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DivingFishSong 代表从 Diving-Fish API 获取的歌曲数据结构
type DivingFishSong struct {
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Type   string    `json:"type"` // SD, DX
	DS     []float64 `json:"ds"`
	Level  []string  `json:"level"`
	Cids   []int     `json:"cids"`
	Charts []struct {
		Notes []int  `json:"notes"`
		Char  string `json:"charter"`
	} `json:"charts"`
	BasicInfo struct {
		Title       string  `json:"title"`
		Artist      string  `json:"artist"`
		Genre       string  `json:"genre"`
		BPM         float64 `json:"bpm"`
		ReleaseDate string  `json:"release_date"`
		From        string  `json:"from"`
		IsNew       bool    `json:"is_new"`
	} `json:"basic_info"`
}

type ChartStat struct {
	Cnt     float64 `json:"cnt"`
	Diff    string  `json:"diff"`
	FitDiff float64 `json:"fit_diff"`
	Avg     float64 `json:"avg"`
	AvgDX   float64 `json:"avg_dx"`
	StdDev  float64 `json:"std_dev"`
}

type ChartStatsResponse struct {
	Charts map[string][]ChartStat `json:"charts"`
}

func (c *Client) FetchSongs() ([]model.Song, error) {
	// 1. 获取谱面统计数据
	logger.Info("正在从 Diving-Fish 获取谱面统计数据", "module", "provider.divingfish")
	statsURL := "https://www.diving-fish.com/api/maimaidxprober/chart_stats"

	var chartStats map[string][]ChartStat
	statsReq, _ := http.NewRequest("GET", statsURL, nil)
	statsResp, err := c.httpClient.Do(statsReq)
	if err == nil {
		defer statsResp.Body.Close()
		if statsResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(statsResp.Body)
			var statsResponse ChartStatsResponse
			if err := json.Unmarshal(body, &statsResponse); err == nil {
				chartStats = statsResponse.Charts
				logger.Info("成功获取谱面统计数据", "module", "provider.divingfish", "count", len(chartStats))
			} else {
				logger.Error("解析谱面统计数据失败", "module", "provider.divingfish", "error", err)
			}
		} else {
			logger.Warn("获取谱面统计数据失败", "module", "provider.divingfish", "status_code", statsResp.StatusCode)
		}
	} else {
		logger.Error("请求谱面统计数据失败", "module", "provider.divingfish", "error", err)
	}

	// 2. 获取音乐数据
	logger.Info("正在从 Diving-Fish 获取音乐数据", "module", "provider.divingfish")
	url := "https://www.diving-fish.com/api/maimaidxprober/music_data"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Diving-Fish 数据失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("意外的状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var dfSongs []DivingFishSong
	if err := json.Unmarshal(body, &dfSongs); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	logger.Info("成功从 Diving-Fish 获取歌曲", "module", "provider.divingfish", "count", len(dfSongs))

	var songs []model.Song
	for _, dfSong := range dfSongs {
		idInt, _ := strconv.Atoi(dfSong.ID)
		song := model.Song{
			GameID:      idInt,
			Title:       dfSong.Title,
			Type:        dfSong.Type,
			Artist:      dfSong.BasicInfo.Artist,
			Genre:       dfSong.BasicInfo.Genre,
			BPM:         dfSong.BasicInfo.BPM,
			ReleaseDate: dfSong.BasicInfo.ReleaseDate,
			Version:     dfSong.BasicInfo.From,
			IsNew:       dfSong.BasicInfo.IsNew,
			CoverURL:    getCoverURL(idInt),
		}

		// 处理谱面
		difficulties := []string{"Basic", "Advanced", "Expert", "Master", "Re:Master"}

		// 获取该歌曲的统计数据
		songStats := chartStats[dfSong.ID]

		for i, level := range dfSong.Level {
			if i >= len(difficulties) {
				break
			}

			// 通过难度标签查找匹配的统计数据
			var fitDiff, avgAchieve, avgDX, stdDev float64
			var sampleCount int
			for _, stat := range songStats {
				if stat.Diff == level {
					fitDiff = stat.FitDiff
					avgAchieve = stat.Avg
					avgDX = stat.AvgDX
					stdDev = stat.StdDev
					sampleCount = int(stat.Cnt)
					break
				}
			}

			notesJSON, _ := json.Marshal(dfSong.Charts[i].Notes)

			chart := model.Chart{
				Difficulty:     difficulties[i],
				Level:          level,
				DS:             dfSong.DS[i],
				Notes:          string(notesJSON),
				Charter:        dfSong.Charts[i].Char,
				FitDiff:        fitDiff,
				AvgAchievement: avgAchieve,
				AvgDX:          avgDX,
				StdDev:         stdDev,
				SampleCount:    sampleCount,
			}
			song.Charts = append(song.Charts, chart)
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// getCoverURL 根据歌曲 ID 生成封面图片 URL
func getCoverURL(id int) string {
	coverID := id
	if id > 10000 && id <= 11000 {
		coverID -= 10000
	}
	return fmt.Sprintf("https://www.diving-fish.com/covers/%05d.png", coverID)
}
