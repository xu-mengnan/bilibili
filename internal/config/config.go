package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `json:"server"`
	AI      AIConfig      `json:"ai"`
	Storage StorageConfig `json:"storage"`
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

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir      string `json:"data_dir"`
	AutoSave     bool   `json:"auto_save"`
	SaveInterval int    `json:"save_interval"`
}

// Default 返回安全的默认配置。
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
		},
		AI: AIConfig{
			APIURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
			APIKey: "",
			Model:  "glm-4-flash",
		},
		Storage: StorageConfig{
			DataDir:      "./data",
			AutoSave:     true,
			SaveInterval: 30,
		},
	}
}

// Load 从文件加载配置，并使用环境变量覆盖敏感/运行时配置。
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := applyEnvironment(cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func applyEnvironment(cfg *Config) error {
	if value := strings.TrimSpace(os.Getenv("BILIBILI_HOST")); value != "" {
		cfg.Server.Host = value
	}
	if value := strings.TrimSpace(os.Getenv("BILIBILI_PORT")); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("BILIBILI_PORT 必须是整数: %w", err)
		}
		cfg.Server.Port = port
	}
	if value := strings.TrimSpace(os.Getenv("BILIBILI_DATA_DIR")); value != "" {
		cfg.Storage.DataDir = value
	}
	if value := strings.TrimSpace(os.Getenv("ZHIPU_API_URL")); value != "" {
		cfg.AI.APIURL = value
	}
	if value := strings.TrimSpace(os.Getenv("ZHIPU_API_KEY")); value != "" {
		cfg.AI.APIKey = value
	}
	if value := strings.TrimSpace(os.Getenv("ZHIPU_MODEL")); value != "" {
		cfg.AI.Model = value
	}
	return nil
}

func validate(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port 必须在 1-65535 之间")
	}
	if strings.TrimSpace(cfg.Server.Host) == "" {
		return fmt.Errorf("server.host 不能为空")
	}
	if strings.TrimSpace(cfg.Storage.DataDir) == "" {
		return fmt.Errorf("storage.data_dir 不能为空")
	}
	return nil
}

// LoadDefault 加载默认配置文件。
func LoadDefault() (*Config, error) {
	return Load("./configs/config.json")
}
