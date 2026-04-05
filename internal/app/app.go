package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/controllers"
	"github.com/marlin9993/claude-code-insight/internal/services"
	"github.com/marlin9993/claude-code-insight/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 创建并配置 Gin 路由实例
func SetupRouter(cfg *config.Config) (*gin.Engine, *websocket.Hub) {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	hub := websocket.NewHub()
	go hub.Run()

	fileWatcher, err := services.NewFileWatcher(hub, cfg)
	if err != nil {
		log.Printf("创建文件监听器失败: %v", err)
	} else {
		go fileWatcher.Start()
		log.Println("文件监听器已启动")
	}

	r.GET("/ws", hub.HandleWebSocket)

	return r, hub
}

// RegisterAPIRoutes 注册所有 API 路由
func RegisterAPIRoutes(r *gin.Engine, cfg *config.Config) {
	historyCtrl := controllers.NewHistoryController(cfg)
	sessionCtrl := controllers.NewSessionController(cfg)
	tokenCtrl := controllers.NewTokenController(cfg)
	searchCtrl := controllers.NewSearchController(cfg)
	shareCtrl := controllers.NewShareController(cfg)
	backupCtrl := controllers.NewBackupController(cfg)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": cfg.Now(),
		})
	})

	api := r.Group("/api")
	{
		history := api.Group("/history")
		{
			history.GET("", historyCtrl.GetHistory)
			history.GET("/search", historyCtrl.SearchHistory)
			history.GET("/fuzzy", searchCtrl.FuzzySearch)
			history.POST("/search", searchCtrl.AdvancedSearch)
			history.GET("/stats", historyCtrl.GetStats)
		}

		api.GET("/projects", historyCtrl.GetProjects)

		sessions := api.Group("/sessions")
		{
			sessions.GET("/:sessionId", sessionCtrl.GetSession)
			sessions.GET("/:sessionId/messages", sessionCtrl.GetSessionMessages)
		}

		tokens := api.Group("/tokens")
		{
			tokens.GET("/session/:sessionId", tokenCtrl.GetSessionTokenStats)
			tokens.GET("/project/*projectPath", tokenCtrl.GetProjectTokenStats)
			tokens.GET("/global", tokenCtrl.GetGlobalTokenStats)
		}

		shares := api.Group("/shares")
		{
			shares.POST("", shareCtrl.CreateShare)
			shares.GET("", shareCtrl.ListShares)
			shares.GET("/:shareId", shareCtrl.GetShare)
			shares.GET("/:shareId/info", shareCtrl.GetShareInfo)
			shares.DELETE("/:shareId", shareCtrl.DeleteShare)
		}

		backup := api.Group("/backup")
		{
			backup.GET("/download", backupCtrl.DownloadBackup)
		}
	}
}

// RegisterFrontendRoutes 托管磁盘上的前端静态文件和 SPA 回退路由。
func RegisterFrontendRoutes(r *gin.Engine) {
	distPath := getEnv("FRONTEND_DIST", filepath.Join("..", "frontend", "dist"))
	// 转为绝对路径（基于当前工作目录）
	if absPath, err := filepath.Abs(distPath); err == nil {
		distPath = absPath
	}
	if info, err := os.Stat(distPath); err == nil && info.IsDir() {
		r.Static("/assets", filepath.Join(distPath, "assets"))
		r.StaticFile("/vite.svg", filepath.Join(distPath, "vite.svg"))
		r.NoRoute(func(c *gin.Context) {
			path := filepath.Join(distPath, "index.html")
			c.File(path)
		})
		log.Printf("前端静态文件目录: %s", distPath)
	} else {
		log.Printf("前端 dist 目录不存在: %s, 跳过静态文件托管", distPath)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StartServer 启动服务器并显示横幅
func StartServer(r *gin.Engine, port int, banner string) {
	addr := fmt.Sprintf(":%d", port)

	if banner != "" {
		fmt.Println(banner)
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
		os.Exit(1)
	}
}
