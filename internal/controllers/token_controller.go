package controllers

import (
	"github.com/marlin9993/claude-code-insight/internal/cache"
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TokenController Token统计控制器
type TokenController struct {
	cfg *config.Config
}

// NewTokenController 创建Token统计控制器
func NewTokenController(cfg *config.Config) *TokenController {
	return &TokenController{cfg: cfg}
}

// GetSessionTokenStats 获取会话token统计（带缓存）
func (tc *TokenController) GetSessionTokenStats(c *gin.Context) {
	sessionID := c.Param("sessionId")
	config := tc.cfg

	// 尝试从缓存获取
	cacheKey := "session:" + sessionID
	if cached, found := cache.TokenStatsCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	// 计算统计数据
	stats, err := services.CalculateSessionTokens(sessionID, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 存入缓存
	cache.TokenStatsCache.Set(cacheKey, stats)

	c.JSON(http.StatusOK, stats)
}

// GetProjectTokenStats 获取项目token统计（带缓存）
func (tc *TokenController) GetProjectTokenStats(c *gin.Context) {
	projectPath := c.Param("projectPath")
	config := tc.cfg
	// 去掉可能存在的多余前导斜杠（Gin 的 *param 会包含前导斜杠）
	for len(projectPath) > 0 && projectPath[0] == '/' {
		projectPath = projectPath[1:]
	}
	// 恢复正确的绝对路径格式
	projectPath = "/" + projectPath

	// 尝试从缓存获取
	cacheKey := "project:" + projectPath
	if cached, found := cache.ProjectTokenStatsCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	// 计算统计数据
	stats, err := services.CalculateProjectTokens(projectPath, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 存入缓存
	cache.ProjectTokenStatsCache.Set(cacheKey, stats)

	c.JSON(http.StatusOK, stats)
}

// GetGlobalTokenStats 获取全局token统计（按天聚合，带缓存）
func (tc *TokenController) GetGlobalTokenStats(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	config := tc.cfg

	// 尝试从缓存获取
	cacheKey := "global:" + startDate + ":" + endDate
	if cached, found := cache.GlobalTokenStatsCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	dailyStats, err := services.CalculateDailyTokenStats(startDate, endDate, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	modelStats, err := services.CalculateModelUsageStats(startDate, endDate, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totals := services.CalculateGlobalTotals(dailyStats)

	result := gin.H{
		"daily":  dailyStats,
		"totals": totals,
		"models": modelStats,
	}

	// 存入缓存
	cache.GlobalTokenStatsCache.Set(cacheKey, result)

	c.JSON(http.StatusOK, result)
}
