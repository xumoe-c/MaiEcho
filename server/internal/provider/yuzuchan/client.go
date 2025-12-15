package yuzuchan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/logger"
)

const BaseURL = "https://www.yuzuchan.moe/api/maimaidx"

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

func (c *Client) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 规避WAF
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	return c.httpClient.Do(req)
}

type AliasResponse struct {
	StatusCode int         `json:"status_code"`
	Content    []AliasItem `json:"content"`
}

type SingleAliasResponse struct {
	StatusCode int       `json:"status_code"`
	Content    AliasItem `json:"content"`
}

type AliasItem struct {
	SongID int      `json:"SongID"`
	Name   string   `json:"Name"`
	Alias  []string `json:"Alias"`
}

func (c *Client) FetchAliases() ([]AliasItem, error) {
	logger.Info("正在从 YuzuChan 获取别名数据", "module", "provider.yuzuchan")
	resp, err := c.doRequest(BaseURL + "/maimaidxalias")
	if err != nil {
		logger.Error("获取别名失败", "module", "provider.yuzuchan", "error", err)
		return nil, fmt.Errorf("获取别名失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("获取别名返回异常状态码", "module", "provider.yuzuchan", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("意外的状态码: %d", resp.StatusCode)
	}

	var result AliasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("解析别名响应失败", "module", "provider.yuzuchan", "error", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logger.Info("成功从 YuzuChan 获取别名数据", "module", "provider.yuzuchan", "count", len(result.Content))
	return result.Content, nil
}

func (c *Client) FetchAliasBySongID(songID int) (*AliasItem, error) {
	url := fmt.Sprintf("%s/getsongsalias?song_id=%d", BaseURL, songID)
	resp, err := c.doRequest(url)
	if err != nil {
		logger.Error("获取歌曲别名失败", "module", "provider.yuzuchan", "songID", songID, "error", err)
		return nil, fmt.Errorf("获取歌曲 %d 的别名失败: %w", songID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			logger.Debug("歌曲别名不存在", "module", "provider.yuzuchan", "songID", songID)
			return nil, nil
		}
		logger.Error("获取歌曲别名返回异常状态码", "module", "provider.yuzuchan", "songID", songID, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("意外的状态码: %d", resp.StatusCode)
	}

	var result SingleAliasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("解析歌曲别名响应失败", "module", "provider.yuzuchan", "songID", songID, "error", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logger.Info("成功获取歌曲别名", "module", "provider.yuzuchan", "songID", songID, "name", result.Content.Name)
	return &result.Content, nil
}
