package models

// HistoryItem 历史记录项
type HistoryItem struct {
	Timestamp int64  `json:"timestamp"`
	SessionID string `json:"sessionId"`
	Display   string `json:"display"`
	Project   string `json:"project"`
}

// SessionMessage 会话消息
type SessionMessage struct {
	Type    string `json:"type"`
	Message struct {
		Type      string      `json:"type"`
		Role      string      `json:"role"`
		Content   string      `json:"content"`
		Usage     *TokenUsage `json:"usage,omitempty"`
		Model     string      `json:"model,omitempty"`
	} `json:"message"`
}

// TokenUsage Token使用量
type TokenUsage struct {
	InputTokens            int `json:"input_tokens"`
	OutputTokens           int `json:"output_tokens"`
	CacheReadInputTokens   int `json:"cache_read_input_tokens"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// HistoryResponse 历史记录响应
type HistoryResponse struct {
	Data       []HistoryItem     `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// StatsResponse 统计响应
type StatsResponse struct {
	TotalSessions int               `json:"totalSessions"`
	TotalMessages int               `json:"totalMessages"`
	ProjectCounts map[string]int    `json:"projectCounts"`
	DateRange     DateRange         `json:"dateRange"`
}

// DateRange 日期范围
type DateRange struct {
	Earliest string `json:"earliest"`
	Latest   string `json:"latest"`
}

// SessionResponse 会话响应
type SessionResponse struct {
	SessionID string            `json:"sessionId"`
	Messages  []SessionMessage  `json:"messages"`
	Total     int               `json:"total"`
}

// TokenStatsResponse Token统计响应
type TokenStatsResponse struct {
	SessionID       string  `json:"sessionId,omitempty"`
	ProjectPath     string  `json:"projectPath,omitempty"`
	ProjectName     string  `json:"projectName,omitempty"`
	MessageCount    int     `json:"messageCount,omitempty"`
	SessionCount    int     `json:"sessionCount,omitempty"`
	InputTokens     int     `json:"inputTokens"`
	OutputTokens    int     `json:"outputTokens"`
	CacheReadTokens int     `json:"cacheReadTokens"`
	TotalTokens     int     `json:"totalTokens"`
}

// GlobalTokenStatsResponse 全局Token统计响应
type GlobalTokenStatsResponse struct {
	Daily  []DailyTokenStats   `json:"daily"`
	Totals TokenStatsResponse  `json:"totals"`
}

// DailyTokenStats 每日Token统计
type DailyTokenStats struct {
	Date            string `json:"date"`
	InputTokens     int    `json:"inputTokens"`
	OutputTokens    int    `json:"outputTokens"`
	CacheReadTokens int    `json:"cacheReadTokens"`
	TotalTokens     int    `json:"totalTokens"`
	SessionCount    int    `json:"sessionCount"`
}
