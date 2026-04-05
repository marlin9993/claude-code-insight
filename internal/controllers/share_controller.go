package controllers

import (
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ShareController 分享控制器
type ShareController struct {
	cfg *config.Config
}

// NewShareController 创建分享控制器
func NewShareController(cfg *config.Config) *ShareController {
	loadSharesFromDisk(cfg)
	return &ShareController{cfg: cfg}
}

// ShareLink 分享链接
type ShareLink struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	IsPublic  bool      `json:"isPublic"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	shares  = make(map[string]*ShareLink)
	shareMu sync.RWMutex
)

func getShareStorePath(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	return filepath.Join(filepath.Dir(cfg.Claude.HistoryPath), "claude-code-insight-shares.json")
}

func loadSharesFromDisk(cfg *config.Config) {
	path := getShareStorePath(cfg)
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("警告: 读取分享存储失败: %v", err)
		}
		return
	}

	var persisted []*ShareLink
	if err := json.Unmarshal(data, &persisted); err != nil {
		log.Printf("警告: 解析分享存储失败: %v", err)
		return
	}

	shareMu.Lock()
	defer shareMu.Unlock()

	shares = make(map[string]*ShareLink, len(persisted))
	for _, share := range persisted {
		if share != nil && share.ID != "" {
			shares[share.ID] = share
		}
	}
}

func saveSharesToDiskLocked(cfg *config.Config) {
	path := getShareStorePath(cfg)
	if path == "" {
		return
	}

	persisted := make([]*ShareLink, 0, len(shares))
	for _, share := range shares {
		persisted = append(persisted, share)
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		log.Printf("警告: 序列化分享存储失败: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("警告: 写入分享存储失败: %v", err)
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateShare 创建分享链接
func (sc *ShareController) CreateShare(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId" binding:"required"`
		IsPublic  bool   `json:"isPublic"`
		ExpiresIn int    `json:"expiresIn"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	config := sc.cfg
	_, err := services.FindSessionFile(req.SessionID, config.Claude.ProjectsPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	share := &ShareLink{
		ID:        generateID(),
		SessionID: req.SessionID,
		IsPublic:  req.IsPublic,
		CreatedAt: time.Now(),
	}

	if req.ExpiresIn > 0 {
		share.ExpiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
	}

	shareMu.Lock()
	shares[share.ID] = share
	saveSharesToDiskLocked(sc.cfg)
	shareMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"shareId":   share.ID,
		"sessionId": req.SessionID,
		"url":       "/api/shares/" + share.ID,
		"expiresAt": share.ExpiresAt,
	})
}

// GetShare 获取分享的会话
func (sc *ShareController) GetShare(c *gin.Context) {
	shareID := c.Param("shareId")

	shareMu.RLock()
	share, exists := shares[shareID]
	shareMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "分享链接不存在"})
		return
	}

	if !share.ExpiresAt.IsZero() && time.Now().After(share.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "分享链接已过期"})
		return
	}

	config := sc.cfg
	sessionPath, err := services.FindSessionFile(share.SessionID, config.Claude.ProjectsPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话文件不存在"})
		return
	}

	messages, err := services.ReadSessionMessages(sessionPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取会话失败"})
		return
	}

	historyItems, err := services.ReadJSONL(config.Claude.HistoryPath, func(item map[string]interface{}) bool {
		sessionID, ok := item["sessionId"].(string)
		return ok && sessionID == share.SessionID
	})
	if err == nil && len(historyItems) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"shareId":   shareID,
			"sessionId": share.SessionID,
			"metadata":  historyItems[0],
			"messages":  messages,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shareId":   shareID,
		"sessionId": share.SessionID,
		"messages":  messages,
	})
}

// GetShareInfo 获取分享链接信息
func (sc *ShareController) GetShareInfo(c *gin.Context) {
	shareID := c.Param("shareId")

	shareMu.RLock()
	share, exists := shares[shareID]
	shareMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "分享链接不存在"})
		return
	}

	if !share.ExpiresAt.IsZero() && time.Now().After(share.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "分享链接已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shareId":   share.ID,
		"sessionId": share.SessionID,
		"isPublic":  share.IsPublic,
		"createdAt": share.CreatedAt,
		"expiresAt": share.ExpiresAt,
	})
}

// DeleteShare 删除分享链接
func (sc *ShareController) DeleteShare(c *gin.Context) {
	shareID := c.Param("shareId")

	shareMu.Lock()
	defer shareMu.Unlock()

	if _, exists := shares[shareID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "分享链接不存在"})
		return
	}

	delete(shares, shareID)
	saveSharesToDiskLocked(sc.cfg)

	c.JSON(http.StatusOK, gin.H{
		"message": "分享链接已删除",
	})
}

// ListShares 列出某个会话的所有分享链接
func (sc *ShareController) ListShares(c *gin.Context) {
	sessionID := c.Query("sessionId")

	shareMu.RLock()
	defer shareMu.RUnlock()

	var result []ShareLink
	for _, share := range shares {
		if sessionID == "" || share.SessionID == sessionID {
			result = append(result, *share)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"shares": result,
		"total":  len(result),
	})
}

// CleanupExpiredShares 清理过期的分享链接
func CleanupExpiredShares(cfg *config.Config) {
	shareMu.Lock()
	defer shareMu.Unlock()

	now := time.Now()
	for id, share := range shares {
		if !share.ExpiresAt.IsZero() && now.After(share.ExpiresAt) {
			delete(shares, id)
		}
	}
	saveSharesToDiskLocked(cfg)
}
