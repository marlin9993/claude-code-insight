package services

import (
	"github.com/marlin9993/claude-code-insight/internal/config"
	"fmt"
	"strings"
	"time"
)

// toInt 安全地将 interface{} 转换为 int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		return 0
	}
}

// TokenUsage Token使用量
type TokenUsage struct {
	InputTokens          int `json:"input_tokens"`
	OutputTokens         int `json:"output_tokens"`
	CacheReadInputTokens int `json:"cache_read_input_tokens"`
}

type ModelUsageStats struct {
	Model           string `json:"model"`
	InputTokens     int    `json:"inputTokens"`
	OutputTokens    int    `json:"outputTokens"`
	CacheReadTokens int    `json:"cacheReadTokens"`
	TotalTokens     int    `json:"totalTokens"`
	SessionCount    int    `json:"sessionCount"`
}

// CalculateSessionTokens 计算会话token统计
func CalculateSessionTokens(sessionID string, cfg *config.Config) (map[string]interface{}, error) {
	sessionFilePath, err := FindSessionFile(sessionID, cfg.Claude.ProjectsPath)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	messages, err := ReadSessionMessages(sessionFilePath)
	if err != nil {
		return nil, err
	}

	// 过滤掉 file-history-snapshot 类型
	conversationMessages := filterMessages(messages)

	var totalInput, totalOutput, totalCacheRead int

	for _, msg := range conversationMessages {
		tokens := extractMessageTokens(msg)
		totalInput += tokens.InputTokens
		totalOutput += tokens.OutputTokens
		totalCacheRead += tokens.CacheReadInputTokens
	}

	return map[string]interface{}{
		"sessionId":       sessionID,
		"messageCount":    len(conversationMessages),
		"inputTokens":     totalInput,
		"outputTokens":    totalOutput,
		"cacheReadTokens": totalCacheRead,
		"totalTokens":     totalInput + totalOutput + totalCacheRead,
	}, nil
}

// CalculateProjectTokens 计算项目token统计
func CalculateProjectTokens(projectPath string, cfg *config.Config) (map[string]interface{}, error) {
	historyItems, err := ReadJSONL(cfg.Claude.HistoryPath, func(item map[string]interface{}) bool {
		return item["project"] == projectPath
	})
	if err != nil {
		return nil, err
	}

	var totalInput, totalOutput, totalCacheRead int
	sessionIDs := uniqueSessionIDs(historyItems)
	sessionCount := len(sessionIDs)

	for _, sessionID := range sessionIDs {
		sessionTokens, err := CalculateSessionTokens(sessionID, cfg)
		if err != nil {
			continue
		}

		totalInput += toInt(sessionTokens["inputTokens"])
		totalOutput += toInt(sessionTokens["outputTokens"])
		totalCacheRead += toInt(sessionTokens["cacheReadTokens"])
	}

	pathParts := strings.Split(projectPath, "/")
	projectName := pathParts[len(pathParts)-1]

	return map[string]interface{}{
		"projectPath":     projectPath,
		"projectName":     projectName,
		"sessionCount":    sessionCount,
		"inputTokens":     totalInput,
		"outputTokens":    totalOutput,
		"cacheReadTokens": totalCacheRead,
		"totalTokens":     totalInput + totalOutput + totalCacheRead,
	}, nil
}

// CalculateDailyTokenStats 计算每日token统计
func CalculateDailyTokenStats(startDate, endDate string, cfg *config.Config) ([]map[string]interface{}, error) {
	historyItems, err := ReadJSONL(cfg.Claude.HistoryPath, nil)
	if err != nil {
		return nil, err
	}

	// 按日期聚合
	dailyStats := make(map[string]*map[string]interface{})
	seenByDay := make(map[string]map[string]struct{})

	for _, item := range historyItems {
		sessionID, _ := item["sessionId"].(string)
		timestamp, _ := item["timestamp"].(float64)

		sessionTokens, err := CalculateSessionTokens(sessionID, cfg)
		if err != nil {
			continue
		}

		date := formatTimestamp(int64(timestamp))

		if _, exists := dailyStats[date]; !exists {
			dailyStats[date] = &map[string]interface{}{
				"date":            date,
				"inputTokens":     0,
				"outputTokens":    0,
				"cacheReadTokens": 0,
				"totalTokens":     0,
				"sessionCount":    0,
			}
		}
		if _, exists := seenByDay[date]; !exists {
			seenByDay[date] = make(map[string]struct{})
		}
		if _, exists := seenByDay[date][sessionID]; exists {
			continue
		}
		seenByDay[date][sessionID] = struct{}{}

		stats := dailyStats[date]
		(*stats)["inputTokens"] = toInt((*stats)["inputTokens"]) + toInt(sessionTokens["inputTokens"])
		(*stats)["outputTokens"] = toInt((*stats)["outputTokens"]) + toInt(sessionTokens["outputTokens"])
		(*stats)["cacheReadTokens"] = toInt((*stats)["cacheReadTokens"]) + toInt(sessionTokens["cacheReadTokens"])
		(*stats)["totalTokens"] = toInt((*stats)["totalTokens"]) + toInt(sessionTokens["totalTokens"])
		(*stats)["sessionCount"] = toInt((*stats)["sessionCount"]) + 1
	}

	// 转换为数组并排序
	var result []map[string]interface{}
	for _, stats := range dailyStats {
		result = append(result, *stats)
	}

	// 按日期排序
	sortDailyStats(result)

	// 过滤日期范围
	if startDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		var filtered []map[string]interface{}
		for _, stat := range result {
			date, _ := time.Parse("2006-01-02", stat["date"].(string))
			if !date.Before(start) {
				filtered = append(filtered, stat)
			}
		}
		result = filtered
	}

	if endDate != "" {
		end, _ := time.Parse("2006-01-02", endDate)
		var filtered []map[string]interface{}
		for _, stat := range result {
			date, _ := time.Parse("2006-01-02", stat["date"].(string))
			if !date.After(end) {
				filtered = append(filtered, stat)
			}
		}
		result = filtered
	}

	return result, nil
}

// CalculateGlobalTotals 计算全局总计
func CalculateGlobalTotals(dailyStats []map[string]interface{}) map[string]interface{} {
	totals := map[string]interface{}{
		"inputTokens":     0,
		"outputTokens":    0,
		"cacheReadTokens": 0,
		"totalTokens":     0,
		"sessionCount":    0,
	}

	for _, day := range dailyStats {
		totals["inputTokens"] = toInt(totals["inputTokens"]) + toInt(day["inputTokens"])
		totals["outputTokens"] = toInt(totals["outputTokens"]) + toInt(day["outputTokens"])
		totals["cacheReadTokens"] = toInt(totals["cacheReadTokens"]) + toInt(day["cacheReadTokens"])
		totals["totalTokens"] = toInt(totals["totalTokens"]) + toInt(day["totalTokens"])
		totals["sessionCount"] = toInt(totals["sessionCount"]) + toInt(day["sessionCount"])
	}

	return totals
}

// CalculateModelUsageStats 按模型聚合全局 token 统计
func CalculateModelUsageStats(startDate, endDate string, cfg *config.Config) ([]map[string]interface{}, error) {
	historyItems, err := ReadJSONL(cfg.Claude.HistoryPath, nil)
	if err != nil {
		return nil, err
	}

	sessionIDs := uniqueSessionIDs(historyItems)
	models := make(map[string]*ModelUsageStats)
	modelSessions := make(map[string]map[string]struct{})

	for _, sessionID := range sessionIDs {
		sessionFilePath, err := FindSessionFile(sessionID, cfg.Claude.ProjectsPath)
		if err != nil {
			continue
		}

		messages, err := ReadSessionMessages(sessionFilePath)
		if err != nil {
			continue
		}

		conversationMessages := filterMessages(messages)
		sessionDate := sessionDateFromMessages(conversationMessages)
		if !isWithinDateRange(sessionDate, startDate, endDate) {
			continue
		}

		for _, msg := range conversationMessages {
			modelName := extractMessageModel(msg)
			if modelName == "" {
				continue
			}

			tokens := extractMessageTokens(msg)
			stats, exists := models[modelName]
			if !exists {
				stats = &ModelUsageStats{Model: modelName}
				models[modelName] = stats
			}

			stats.InputTokens += tokens.InputTokens
			stats.OutputTokens += tokens.OutputTokens
			stats.CacheReadTokens += tokens.CacheReadInputTokens
			stats.TotalTokens += tokens.InputTokens + tokens.OutputTokens + tokens.CacheReadInputTokens

			if _, exists := modelSessions[modelName]; !exists {
				modelSessions[modelName] = make(map[string]struct{})
			}
			modelSessions[modelName][sessionID] = struct{}{}
		}
	}

	var result []map[string]interface{}
	for modelName, stats := range models {
		stats.SessionCount = len(modelSessions[modelName])
		result = append(result, map[string]interface{}{
			"model":           stats.Model,
			"inputTokens":     stats.InputTokens,
			"outputTokens":    stats.OutputTokens,
			"cacheReadTokens": stats.CacheReadTokens,
			"totalTokens":     stats.TotalTokens,
			"sessionCount":    stats.SessionCount,
		})
	}

	sortModelUsageStats(result)
	return result, nil
}

// extractMessageTokens 提取消息的token使用量
func extractMessageTokens(msg map[string]interface{}) TokenUsage {
	message, ok := msg["message"].(map[string]interface{})
	if !ok {
		return TokenUsage{}
	}

	usage, ok := message["usage"].(map[string]interface{})
	if !ok {
		return TokenUsage{}
	}

	result := TokenUsage{}
	if inputTokens, ok := usage["input_tokens"].(float64); ok {
		result.InputTokens = int(inputTokens)
	}
	if outputTokens, ok := usage["output_tokens"].(float64); ok {
		result.OutputTokens = int(outputTokens)
	}
	if cacheTokens, ok := usage["cache_read_input_tokens"].(float64); ok {
		result.CacheReadInputTokens = int(cacheTokens)
	}

	return result
}

func extractMessageModel(msg map[string]interface{}) string {
	message, ok := msg["message"].(map[string]interface{})
	if !ok {
		return ""
	}

	model, _ := message["model"].(string)
	return strings.TrimSpace(model)
}

// filterMessages 过滤掉file-history-snapshot类型的消息
func filterMessages(messages []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, msg := range messages {
		if msgType, ok := msg["type"].(string); ok && msgType != "file-history-snapshot" {
			result = append(result, msg)
		}
	}
	return result
}

// formatTimestamp 格式化时间戳为日期字符串
func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp/1000, 0)
	return t.Format("2006-01-02")
}

func sessionDateFromMessages(messages []map[string]interface{}) string {
	for _, msg := range messages {
		if timestamp, ok := msg["timestamp"].(string); ok && timestamp != "" {
			if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil {
				return parsed.Format("2006-01-02")
			}
		}

		if createdAt, ok := msg["created_at"].(string); ok && createdAt != "" {
			if parsed, err := time.Parse(time.RFC3339, createdAt); err == nil {
				return parsed.Format("2006-01-02")
			}
		}
	}

	return ""
}

func isWithinDateRange(dateValue, startDate, endDate string) bool {
	if dateValue == "" {
		return startDate == "" && endDate == ""
	}

	date, err := time.Parse("2006-01-02", dateValue)
	if err != nil {
		return false
	}

	if startDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err == nil && date.Before(start) {
			return false
		}
	}

	if endDate != "" {
		end, err := time.Parse("2006-01-02", endDate)
		if err == nil && date.After(end) {
			return false
		}
	}

	return true
}

// sortDailyStats 按日期排序每日统计
func sortDailyStats(stats []map[string]interface{}) {
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			dateI, _ := time.Parse("2006-01-02", stats[i]["date"].(string))
			dateJ, _ := time.Parse("2006-01-02", stats[j]["date"].(string))
			if dateI.After(dateJ) {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
}

func sortModelUsageStats(stats []map[string]interface{}) {
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if toInt(stats[i]["totalTokens"]) < toInt(stats[j]["totalTokens"]) {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
}

func uniqueSessionIDs(items []map[string]interface{}) []string {
	seen := make(map[string]struct{})
	var sessionIDs []string

	for _, item := range items {
		sessionID, ok := item["sessionId"].(string)
		if !ok || sessionID == "" {
			continue
		}
		if _, exists := seen[sessionID]; exists {
			continue
		}
		seen[sessionID] = struct{}{}
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs
}
