package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CorpusPath       string `yaml:"corpus_path"`
	DefaultResponse  string `yaml:"default_response"`
	MatchThreshold   int    `yaml:"match_threshold"`
	EnableAutoBackup bool   `yaml:"enable_auto_backup"` // 新增：是否启用自动备份
	MaxBackupFiles   int    `yaml:"max_backup_files"`   // 新增：最大备份文件数
}

// 在 config/config.go 中添加
func (c *Config) Validate() error {
	if c.CorpusPath == "" {
		return fmt.Errorf("corpus_path 不能为空")
	}

	if _, err := os.Stat(c.CorpusPath); os.IsNotExist(err) {
		return fmt.Errorf("语料库目录不存在: %s", c.CorpusPath)
	}

	return nil
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.CorpusPath == "" {
		config.CorpusPath = "./corpus"
	}
	if config.DefaultResponse == "" {
		config.DefaultResponse = "抱歉，我没明白您的意思。"
	}
	if config.MatchThreshold == 0 {
		config.MatchThreshold = 70
	}
	if config.MaxBackupFiles == 0 {
		config.MaxBackupFiles = 10 // 默认保留10个备份
	}

	return &config, nil
}
