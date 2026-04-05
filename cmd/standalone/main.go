package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/marlin9993/claude-code-insight/internal/app"
	"github.com/marlin9993/claude-code-insight/internal/config"

	"flag"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var frontendFS embed.FS

func main() {
	// 解析命令行参数
	port := flag.Int("p", 3000, "服务端口")
	flag.Parse()

	// 验证前端静态文件
	if err := validateFrontend(); err != nil {
		log.Fatalf("✗ 前端验证失败: %v\n请确保已运行 build.sh 构建前端", err)
	}
	log.Println("✓ 前端静态文件验证成功")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("✗ 配置加载失败: %v", err)
	}
	cfg.Server.Port = *port
	cfg.Server.CORSOrigins = []string{"*"} // 同源部署，允许所有来源

	// 验证 Claude 目录
	if err := cfg.Validate(); err != nil {
		log.Fatalf("✗ Claude Code 目录验证失败: %v", err)
	}
	log.Println("✓ Claude Code 目录验证成功")

	// 设置路由（包含 WebSocket Hub）
	r, _ := app.SetupRouter(cfg)

	// 注册 API 路由
	app.RegisterAPIRoutes(r, cfg)

	// 设置前端静态文件服务
	setupStaticServer(r)

	// 启动服务器
	banner := fmt.Sprintf(`
╔═══════════════════════════════════════════════════════╗
║   Claude Code Insight - 一体化版本                   ║
╠═══════════════════════════════════════════════════════╣
║   服务地址: http://localhost:%d                     ║
║   前端页面: http://localhost:%d                     ║
║   API 基础路径: /api                                  ║
║   健康检查: http://localhost:%d/health              ║
╚═══════════════════════════════════════════════════════╝
`, *port, *port, *port)

	app.StartServer(r, *port, banner)
}

// validateFrontend 验证前端静态文件是否已嵌入
func validateFrontend() error {
	// 尝试读取 index.html
	subFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return fmt.Errorf("无法访问前端文件系统: %w", err)
	}

	// 检查 index.html 是否存在
	if _, err := subFS.Open("index.html"); err != nil {
		return fmt.Errorf("index.html 不存在: %w", err)
	}

	return nil
}

// setupStaticServer 设置静态文件服务和 SPA 路由回退
func setupStaticServer(r *gin.Engine) {
	// 创建前端文件系统的子文件系统
	frontendSubFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		log.Fatalf("无法创建前端文件系统: %v", err)
	}

	// 读取并缓存 index.html（避免 http.FileServer 的 301 重定向）
	indexHTML, err := fs.ReadFile(frontendSubFS, "index.html")
	if err != nil {
		log.Fatalf("无法读取 index.html: %v", err)
	}

	// 根路径 - 返回缓存的 index.html
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})

	// 静态文件服务
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(frontendSubFS))
	})

	// SPA 路由回退 - 对于所有非 API 请求，返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 路由，返回 404 JSON
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API 接口不存在"})
			return
		}

		// 对于所有其他路由，返回缓存的 index.html（SPA 路由回退）
		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
}
