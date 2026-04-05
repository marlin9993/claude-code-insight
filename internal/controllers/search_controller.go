package controllers

import (
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SearchController 搜索控制器
type SearchController struct {
	cfg *config.Config
}

// NewSearchController 创建搜索控制器
func NewSearchController(cfg *config.Config) *SearchController {
	return &SearchController{cfg: cfg}
}

// SearchRequest 高级搜索请求
type SearchRequest struct {
	Keyword        string `json:"keyword"`
	Project        string `json:"project,omitempty"`
	StartDate      string `json:"startDate,omitempty"`
	EndDate        string `json:"endDate,omitempty"`
	SearchContent  bool   `json:"searchContent"`
	MessageCountMin int   `json:"messageCountMin,omitempty"`
	MessageCountMax int   `json:"messageCountMax,omitempty"`
}

// AdvancedSearch 高级搜索历史记录（POST 请求）
func (sc *SearchController) AdvancedSearch(c *gin.Context) {
	var req SearchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	config := sc.cfg
	// 读取全部历史记录
	historyItems, err := services.ReadJSONL(config.Claude.HistoryPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 应用过滤条件
	var results []map[string]interface{}
	for _, item := range historyItems {
		// 项目过滤
		if req.Project != "" {
			if project, ok := item["project"].(string); ok {
				if project != req.Project {
					continue
				}
			} else {
				continue
			}
		}

		// 日期范围过滤
		if timestampFloat, ok := item["timestamp"].(float64); ok {
			timestamp := int64(timestampFloat)
			if req.StartDate != "" {
				startTime := parseDateToTime(req.StartDate)
				if !startTime.IsZero() && timestamp < startTime.Unix()*1000 {
					continue
				}
			}
			if req.EndDate != "" {
				endTime := parseDateToTime(req.EndDate)
				if !endTime.IsZero() {
					endTime = endTime.Add(24 * time.Hour)
					if timestamp > endTime.Unix()*1000 {
						continue
					}
				}
			}
		}

		// 关键词搜索（描述 + 内容）
		if req.Keyword != "" {
			if !services.MatchKeyword(item, req.Keyword, config.Claude.ProjectsPath, req.SearchContent) {
				continue
			}
		}

		// 消息数量范围过滤
		if req.MessageCountMin > 0 || req.MessageCountMax > 0 {
			if sessionID, ok := item["sessionId"].(string); ok {
				sessionPath, _ := services.FindSessionFile(sessionID, config.Claude.ProjectsPath)
				if sessionPath != "" {
					count, err := services.GetJSONLLineCount(sessionPath)
					if err != nil {
						continue
					}
					if req.MessageCountMin > 0 && count < req.MessageCountMin {
						continue
					}
					if req.MessageCountMax > 0 && count > req.MessageCountMax {
						continue
					}
				}
			}
		}

		results = append(results, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   results,
		"total":  len(results),
	})
}

// parseDateToTime 解析日期字符串为 time.Time
func parseDateToTime(dateStr string) time.Time {
	// 支持多种日期格式
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// 尝试简单解析 YYYY-MM-DD
	var year, month, day int
	if _, err := fmt.Sscanf(dateStr, "%d-%d-%d", &year, &month, &day); err == nil {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	}

	// 如果解析失败，返回零值
	return time.Time{}
}

// parseDateToTimestamp 解析日期字符串为时间戳（毫秒），兼容现有代码
func parseDateToTimestamp(dateStr string) int64 {
	t := parseDateToTime(dateStr)
	if t.IsZero() {
		// 降级使用简单解析
		var year, month, day int
		fmt.Sscanf(dateStr, "%d-%d-%d", &year, &month, &day)
		return int64(year*365*24*3600 + month*30*24*3600 + day*24*3600) * 1000
	}
	return t.Unix() * 1000
}

// FuzzySearch 模糊搜索（兼容原有的简单搜索）
func (sc *SearchController) FuzzySearch(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少关键词参数"})
		return
	}

	searchContent := c.Query("searchContent") == "true"
	config := sc.cfg

	// 读取全部历史记录
	historyItems, err := services.ReadJSONL(config.Claude.HistoryPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 应用关键词过滤
	var results []map[string]interface{}
	keywordLower := strings.ToLower(keyword)

	for _, item := range historyItems {
		// 检查描述字段
		if display, ok := item["display"].(string); ok {
			if strings.Contains(strings.ToLower(display), keywordLower) {
				results = append(results, item)
				continue
			}
		}

		// 检查消息内容（如果启用）
		if searchContent && services.MatchKeyword(item, keyword, config.Claude.ProjectsPath, true) {
			results = append(results, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"total": len(results),
	})
}
