package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	Server ServerConfig `json:"server"`
	AI     AIConfig     `json:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// AIConfig AI配置
type AIConfig struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	// 设置默认配置
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		AI: AIConfig{
			APIURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
			APIKey: "",
			Model:  "glm-4.7",
		},
	}

	// 尝试读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置文件
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil
}

// LoadDefault 加载默认配置文件
func LoadDefault() (*Config, error) {
	return Load("./configs/config.json")
}
