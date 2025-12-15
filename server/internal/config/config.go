package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type Config struct {
	ServerPort  string         `mapstructure:"server_port"`
	DatabaseURL string         `mapstructure:"database_url"`
	LLM         LLMConfig      `mapstructure:"llm"`
	Log         logger.Config  `mapstructure:"log"`
	Bilibili    BilibiliConfig `mapstructure:"bilibili"`
}

type LLMConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

type BilibiliConfig struct {
	Cookie string `mapstructure:"cookie"`
	Proxy  string `mapstructure:"proxy"`
}

func Load() (*Config, error) {
	v := viper.New()

	// 设置默认值
	v.SetDefault("server_port", ":8080")
	v.SetDefault("database_url", "maiecho.db")
	v.SetDefault("llm.base_url", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	v.SetDefault("llm.model", "qwen-plus")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output_path", "logs/maiecho.log")
	v.SetDefault("log.llm_log_path", "logs/llm_conversations.log")
	v.SetDefault("log.encoding", "console")

	// 读取环境变量
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./server/config")

	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，则继续使用默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error("读取配置文件失败", "module", "config", "error", err)
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		logger.Info("配置文件不存在，使用默认值和环境变量", "module", "config")
	} else {
		logger.Info("配置文件加载成功", "module", "config")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Error("解析配置失败", "module", "config", "error", err)
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必填字段
	if v.GetString("llm.api_key") == "" {
		logger.Error("配置验证失败：llm.api_key 缺失", "module", "config")
		return nil, fmt.Errorf("llm.api_key 是必填项")
	}

	logger.Info("配置加载完成", "module", "config", "serverPort", cfg.ServerPort)
	return &cfg, nil
}
