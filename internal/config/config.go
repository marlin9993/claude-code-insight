package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config 应用配置
type Config struct {
	Claude     ClaudeConfig
	Server     ServerConfig
	Pagination PaginationConfig
}

// ClaudeConfig Claude Code相关配置
type ClaudeConfig struct {
	HistoryPath  string
	ProjectsPath string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int
	Env         string
	CORSOrigins []string
}

// PaginationConfig 分页配置
type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

// Load 加载配置
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	defaultClaudeDir := filepath.Join(homeDir, ".claude")

	return &Config{
		Claude: ClaudeConfig{
			HistoryPath:  getEnv("CLAUDE_HISTORY_PATH", filepath.Join(defaultClaudeDir, "history.jsonl")),
			ProjectsPath: getEnv("CLAUDE_PROJECTS_PATH", filepath.Join(defaultClaudeDir, "projects")),
		},
		Server: ServerConfig{
			Port:        getEnvInt("PORT", 3000),
			Env:         getEnv("NODE_ENV", "production"),
			CORSOrigins: []string{getEnv("CORS_ORIGIN", "http://localhost:5173")},
		},
		Pagination: PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
	}, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证历史记录文件
	if _, err := os.Stat(c.Claude.HistoryPath); os.IsNotExist(err) {
		return fmt.Errorf("历史记录文件不存在: %s", c.Claude.HistoryPath)
	}

	// 验证项目目录
	if _, err := os.Stat(c.Claude.ProjectsPath); os.IsNotExist(err) {
		return fmt.Errorf("项目目录不存在: %s", c.Claude.ProjectsPath)
	}

	return nil
}

// Now 返回当前时间的ISO格式字符串
func (c *Config) Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
