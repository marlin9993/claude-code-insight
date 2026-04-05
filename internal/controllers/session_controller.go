package controllers

import (
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SessionController 会话控制器
type SessionController struct {
	cfg *config.Config
}

// NewSessionController 创建会话控制器
func NewSessionController(cfg *config.Config) *SessionController {
	return &SessionController{cfg: cfg}
}

// GetSession 获取会话详情
func (sc *SessionController) GetSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	config := sc.cfg

	// 查找会话文件
	sessionFilePath, err := services.FindSessionFile(sessionID, config.Claude.ProjectsPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	// 读取会话数据
	messages, err := services.ReadSessionMessages(sessionFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 过滤掉 file-history-snapshot 类型
	conversationMessages := filterConversationMessages(messages)

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"messages":  conversationMessages,
		"total":     len(conversationMessages),
	})
}

// GetSessionMessages 分页获取会话消息
func (sc *SessionController) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("sessionId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	config := sc.cfg
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(config.Pagination.DefaultPageSize)))

	// 查找会话文件
	sessionFilePath, err := services.FindSessionFile(sessionID, config.Claude.ProjectsPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	// 读取所有消息
	allMessages, err := services.ReadSessionMessages(sessionFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 过滤掉 file-history-snapshot 类型
	conversationMessages := filterConversationMessages(allMessages)
	total := len(conversationMessages)

	// 分页
	skip := (page - 1) * pageSize
	end := skip + pageSize
	if end > total {
		end = total
	}

	var paginatedMessages []map[string]interface{}
	if skip < total {
		paginatedMessages = conversationMessages[skip:end]
	} else {
		paginatedMessages = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": paginatedMessages,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		},
	})
}

// filterConversationMessages 过滤掉file-history-snapshot类型的消息
func filterConversationMessages(messages []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, msg := range messages {
		if msgType, ok := msg["type"].(string); ok && msgType != "file-history-snapshot" {
			result = append(result, msg)
		}
	}
	return result
}
