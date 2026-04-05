package services

import (
	"strings"
)

// SearchInMessages 在消息内容中搜索关键词
func SearchInMessages(messages []map[string]interface{}, keyword string) bool {
	keyword = strings.ToLower(keyword)

	for _, msg := range messages {
		if containsInMessageContent(msg, keyword) {
			return true
		}
	}

	return false
}

// containsInMessageContent 递归搜索消息内容
func containsInMessageContent(msg map[string]interface{}, keyword string) bool {
	message, ok := msg["message"].(map[string]interface{})
	if !ok {
		return false
	}

	content, ok := message["content"]
	if !ok {
		return false
	}

	// 处理字符串内容
	if contentStr, ok := content.(string); ok {
		return strings.Contains(strings.ToLower(contentStr), keyword)
	}

	// 处理数组内容（结构化消息）
	if contentArr, ok := content.([]interface{}); ok {
		for _, item := range contentArr {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// 检查 text 字段
				if text, ok := itemMap["text"].(string); ok {
					if strings.Contains(strings.ToLower(text), keyword) {
						return true
					}
				}
				// 检查 thinking 字段
				if thinking, ok := itemMap["thinking"].(string); ok {
					if strings.Contains(strings.ToLower(thinking), keyword) {
						return true
					}
				}
				// 递归检查嵌套结构
				if searchInMap(itemMap, keyword) {
					return true
				}
			}
		}
	}

	return false
}

// searchInMap 递归搜索 map 中的所有字符串值
func searchInMap(m map[string]interface{}, keyword string) bool {
	for _, v := range m {
		switch val := v.(type) {
		case string:
			if strings.Contains(strings.ToLower(val), keyword) {
				return true
			}
		case map[string]interface{}:
			if searchInMap(val, keyword) {
				return true
			}
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if searchInMap(itemMap, keyword) {
						return true
					}
				}
			}
		}
	}
	return false
}

// MatchKeyword 检查关键词是否匹配（描述或内容）
func MatchKeyword(item map[string]interface{}, keyword, projectsPath string, searchContent bool) bool {
	keyword = strings.ToLower(keyword)

	// 检查描述字段
	if display, ok := item["display"].(string); ok {
		if strings.Contains(strings.ToLower(display), keyword) {
			return true
		}
	}

	// 检查消息内容
	if searchContent {
		if sessionID, ok := item["sessionId"].(string); ok {
			sessionPath, _ := FindSessionFile(sessionID, projectsPath)
			if sessionPath != "" {
				messages, err := ReadSessionMessages(sessionPath)
				if err == nil && SearchInMessages(messages, keyword) {
					return true
				}
			}
		}
	}

	return false
}
