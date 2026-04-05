package controllers

import (
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HistoryController 历史记录控制器
type HistoryController struct {
	cfg *config.Config
}

// NewHistoryController 创建历史记录控制器
func NewHistoryController(cfg *config.Config) *HistoryController {
	return &HistoryController{cfg: cfg}
}

// GetHistory 获取历史记录列表（分页）
func (hc *HistoryController) GetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	config := hc.cfg
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(config.Pagination.DefaultPageSize)))
	page, pageSize = normalizePagination(page, pageSize, config.Pagination.DefaultPageSize, config.Pagination.MaxPageSize)

	log.Printf("[GetHistory] 请求参数: page=%d, pageSize=%d", page, pageSize)

	allHistoryItems, err := services.ReadJSONL(config.Claude.HistoryPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total := len(allHistoryItems)
	log.Printf("[GetHistory] 读取到 %d 条历史记录", total)

	var items []map[string]interface{}
	for _, item := range allHistoryItems {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		timestampI, _ := items[i]["timestamp"].(float64)
		timestampJ, _ := items[j]["timestamp"].(float64)
		return int64(timestampI) > int64(timestampJ)
	})

	skip := (page - 1) * pageSize
	end := skip + pageSize
	if skip > len(items) {
		skip = len(items)
	}
	if end > len(items) {
		end = len(items)
	}

	historyItems := items[skip:end]
	log.Printf("[GetHistory] 返回数据: skip=%d, end=%d, 返回 %d 条", skip, end, len(historyItems))

	c.JSON(http.StatusOK, gin.H{
		"data": historyItems,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		},
	})
}

// SearchHistory 搜索历史记录
func (hc *HistoryController) SearchHistory(c *gin.Context) {
	keyword := c.Query("keyword")
	project := c.Query("project")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	config := hc.cfg
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(config.Pagination.DefaultPageSize)))
	page, pageSize = normalizePagination(page, pageSize, config.Pagination.DefaultPageSize, config.Pagination.MaxPageSize)

	filter := func(item map[string]interface{}) bool {
		if keyword != "" {
			searchLower := strings.ToLower(keyword)
			display, ok := item["display"].(string)
			displayMatch := ok && strings.Contains(strings.ToLower(display), searchLower)

			proj, ok := item["project"].(string)
			projectMatch := ok && strings.Contains(strings.ToLower(proj), searchLower)

			if !displayMatch && !projectMatch {
				return false
			}
		}

		if project != "" {
			proj, ok := item["project"].(string)
			if !ok || !strings.Contains(proj, project) {
				return false
			}
		}

		timestamp, _ := item["timestamp"].(float64)
		if startDate != "" {
			startTime := parseDate(startDate)
			if !startTime.IsZero() && int64(timestamp) < startTime.UnixMilli() {
				return false
			}
		}

		if endDate != "" {
			endTime := parseDate(endDate)
			if !endTime.IsZero() && int64(timestamp) >= endTime.Add(24*time.Hour).UnixMilli() {
				return false
			}
		}

		return true
	}

	allItems, err := services.ReadJSONL(config.Claude.HistoryPath, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 按 sessionId 去重，保留每个 session 最新的那条记录
	sessionMap := make(map[string]map[string]interface{})
	for _, item := range allItems {
		sessionID, ok := item["sessionId"].(string)
		if !ok || sessionID == "" {
			continue
		}
		existing, exists := sessionMap[sessionID]
		if !exists {
			sessionMap[sessionID] = item
		} else {
			ts1, _ := item["timestamp"].(float64)
			ts2, _ := existing["timestamp"].(float64)
			if int64(ts1) > int64(ts2) {
				sessionMap[sessionID] = item
			}
		}
	}

	// 转为切片并按时间降序排序
	uniqueItems := make([]map[string]interface{}, 0, len(sessionMap))
	for _, item := range sessionMap {
		uniqueItems = append(uniqueItems, item)
	}
	sort.Slice(uniqueItems, func(i, j int) bool {
		timestampI, _ := uniqueItems[i]["timestamp"].(float64)
		timestampJ, _ := uniqueItems[j]["timestamp"].(float64)
		return int64(timestampI) > int64(timestampJ)
	})

	total := len(uniqueItems)

	skip := (page - 1) * pageSize
	end := skip + pageSize
	if skip > total {
		skip = total
	}
	if end > total {
		end = total
	}

	paginatedItems := uniqueItems[skip:end]

	c.JSON(http.StatusOK, gin.H{
		"data": paginatedItems,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		},
	})
}

// GetStats 获取统计信息
func (hc *HistoryController) GetStats(c *gin.Context) {
	config := hc.cfg
	historyItems, err := services.ReadJSONL(config.Claude.HistoryPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	projectCounts := make(map[string]int)
	sessionSet := make(map[string]bool)
	var timestamps []int64

	for _, item := range historyItems {
		project, ok := item["project"].(string)
		if !ok {
			project = "unknown"
		}
		projectCounts[project]++

		sessionID, ok := item["sessionId"].(string)
		if ok {
			sessionSet[sessionID] = true
		}

		timestamp, _ := item["timestamp"].(float64)
		timestamps = append(timestamps, int64(timestamp))
	}

	var earliestTimestamp, latestTimestamp int64
	if len(timestamps) > 0 {
		earliestTimestamp = timestamps[0]
		latestTimestamp = timestamps[0]
		for _, ts := range timestamps {
			if ts < earliestTimestamp {
				earliestTimestamp = ts
			}
			if ts > latestTimestamp {
				latestTimestamp = ts
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"totalSessions": len(sessionSet),
		"totalMessages": len(historyItems),
		"projectCounts": projectCounts,
		"dateRange": gin.H{
			"earliest": formatTimestamp(earliestTimestamp),
			"latest":   formatTimestamp(latestTimestamp),
		},
	})
}

// GetProjects 获取项目列表（分组聚合）
func (hc *HistoryController) GetProjects(c *gin.Context) {
	config := hc.cfg
	historyItems, err := services.ReadJSONL(config.Claude.HistoryPath, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recentSessionsCount, _ := strconv.Atoi(c.DefaultQuery("recentSessionsCount", "5"))

	// 按 (project, sessionId) 去重，保留每个 session 最新的那条记录
	type sessionKey struct {
		project   string
		sessionID string
	}
	uniqueSessions := make(map[sessionKey]map[string]interface{})
	for _, item := range historyItems {
		project, ok := item["project"].(string)
		if !ok || project == "" {
			continue
		}
		sessionID, ok := item["sessionId"].(string)
		if !ok || sessionID == "" {
			continue
		}

		key := sessionKey{project: project, sessionID: sessionID}
		existing, exists := uniqueSessions[key]
		if !exists {
			uniqueSessions[key] = item
		} else {
			ts1, _ := item["timestamp"].(float64)
			ts2, _ := existing["timestamp"].(float64)
			if int64(ts1) > int64(ts2) {
				uniqueSessions[key] = item
			}
		}
	}

	// 按项目分组
	projectMap := make(map[string]*ProjectInfo)
	for _, item := range uniqueSessions {
		project := item["project"].(string)

		if _, exists := projectMap[project]; !exists {
			projectMap[project] = &ProjectInfo{
				ProjectPath: project,
				Sessions:    []map[string]interface{}{},
			}
		}
		projectMap[project].Sessions = append(projectMap[project].Sessions, item)
	}

	var projects []*ProjectInfo
	for _, info := range projectMap {
		info.SessionCount = len(info.Sessions)
		info.ProjectName = filepath.Base(info.ProjectPath)

		// 按 timestamp 降序排序
		sort.Slice(info.Sessions, func(i, j int) bool {
			tsI, _ := info.Sessions[i]["timestamp"].(float64)
			tsJ, _ := info.Sessions[j]["timestamp"].(float64)
			return int64(tsI) > int64(tsJ)
		})

		// 只保留最近 N 条
		if len(info.Sessions) > recentSessionsCount {
			info.Sessions = info.Sessions[:recentSessionsCount]
		}

		// 取最新时间作为项目最后更新时间
		if len(info.Sessions) > 0 {
			ts, _ := info.Sessions[0]["timestamp"].(float64)
			info.LastUpdate = int64(ts)
		}

		projects = append(projects, info)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUpdate > projects[j].LastUpdate
	})

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	ProjectPath  string                   `json:"projectPath"`
	ProjectName  string                   `json:"projectName"`
	SessionCount int                      `json:"sessionCount"`
	LastUpdate   int64                    `json:"lastUpdate"`
	Sessions     []map[string]interface{} `json:"sessions,omitempty"`
}

func parseDate(dateStr string) time.Time {
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

	log.Printf("警告: 无法解析日期字符串: %s", dateStr)
	return time.Time{}
}

func normalizePagination(page, pageSize, defaultPageSize, maxPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func formatTimestamp(ts int64) string {
	t := time.Unix(ts/1000, 0)
	return t.Format(time.RFC3339)
}
