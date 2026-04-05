package services

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ReadJSONL 读取JSONL文件
func ReadJSONL(filePath string, filter func(map[string]interface{}) bool) ([]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []map[string]interface{}
	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以支持更长的 JSON 行（最大 10MB）
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item map[string]interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue // 跳过解析失败的行
		}

		// 应用过滤条件
		if filter != nil && !filter(item) {
			continue
		}

		results = append(results, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetJSONLLineCount 获取JSONL文件行数
func GetJSONLLineCount(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}

	return count, scanner.Err()
}

// ReadSessionMessages 读取会话消息
func ReadSessionMessages(sessionFilePath string) ([]map[string]interface{}, error) {
	return ReadJSONL(sessionFilePath, nil)
}

// FindSessionFile 根据sessionID查找会话文件
func FindSessionFile(sessionID, projectsPath string) (string, error) {
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionFilePath := filepath.Join(projectsPath, entry.Name(), sessionID+".jsonl")
		if _, err := os.Stat(sessionFilePath); err == nil {
			return sessionFilePath, nil
		}
	}

	return "", os.ErrNotExist
}
