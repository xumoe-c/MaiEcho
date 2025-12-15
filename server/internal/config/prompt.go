package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type PromptConfig struct {
	Agent AgentPrompts `mapstructure:"agent"`
}

type AgentPrompts struct {
	Cleaner   PromptPair       `mapstructure:"cleaner"`
	Analyst   PromptPair       `mapstructure:"analyst"`
	Advisor   PromptPair       `mapstructure:"advisor"`
	Mapper    MapperPrompts    `mapstructure:"mapper"`
	Knowledge KnowledgePrompts `mapstructure:"knowledge"`
	Relevance RelevancePrompts `mapstructure:"relevance"`
}

type RelevancePrompts struct {
	CheckAlias PromptPair `mapstructure:"check_alias"`
	CheckTitle PromptPair `mapstructure:"check_title"`
}

type MapperPrompts struct {
	VerifyMatch PromptPair `mapstructure:"verify_match"`
}

type KnowledgePrompts struct {
	GuideHeader string `mapstructure:"guide_header"`
}

type PromptPair struct {
	System string `mapstructure:"system"`
	User   string `mapstructure:"user"`
}

func LoadPrompts() (*PromptConfig, error) {
	v := viper.New()
	v.SetConfigName("prompts")
	v.SetConfigType("yaml")
	v.AddConfigPath("./server/prompts")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("读取提示词文件失败", "module", "config.prompt", "error", err)
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}

	var cfg PromptConfig
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Error("解析提示词失败", "module", "config.prompt", "error", err)
		return nil, fmt.Errorf("解析提示词失败: %w", err)
	}

	logger.Info("提示词加载成功", "module", "config.prompt")
	return &cfg, nil
}
