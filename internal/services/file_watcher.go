package services

import (
	"bufio"
	"github.com/marlin9993/claude-code-insight/internal/config"
	"github.com/marlin9993/claude-code-insight/internal/websocket"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Message 消息结构
type Message struct {
	UUID string `json:"uuid"`
	Type string `json:"type"`
	// 其他字段...
}

// FileWatcher 监听文件变化
type FileWatcher struct {
	watcher          *fsnotify.Watcher
	hub              *websocket.Hub
	cfg              *config.Config
	stopChan         chan struct{}
	trackedSessions  map[string]int64
	trackedSessionsMu sync.RWMutex
}

// NewFileWatcher 创建文件监听器
func NewFileWatcher(hub *websocket.Hub, cfg *config.Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:          watcher,
		hub:              hub,
		cfg:              cfg,
		stopChan:         make(chan struct{}),
		trackedSessions:  make(map[string]int64),
	}

	return fw, nil
}

// Start 启动文件监听
func (fw *FileWatcher) Start() {
	// 监听历史记录文件
	if err := fw.watcher.Add(fw.cfg.Claude.HistoryPath); err != nil {
		log.Printf("无法监听历史记录文件: %v", err)
	}

	// 监听所有项目目录下的会话文件
	fw.watchSessionFiles()

	// 使用轮询检测新会话（因为会话文件在子目录中）
	go fw.pollSessions()

	// 启动 fsnotify 事件循环
	go fw.watchEvents()
}

// watchSessionFiles 监听所有会话文件
func (fw *FileWatcher) watchSessionFiles() {
	projectsPath := fw.cfg.Claude.ProjectsPath
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		log.Printf("无法读取项目目录: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsPath, entry.Name())
		sessions, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, session := range sessions {
			if filepath.Ext(session.Name()) != ".jsonl" {
				continue
			}

			sessionPath := filepath.Join(projectPath, session.Name())
			// 记录初始文件大小
			if info, err := os.Stat(sessionPath); err == nil {
				fw.trackedSessionsMu.Lock()
				fw.trackedSessions[sessionPath] = info.Size()
				fw.trackedSessionsMu.Unlock()
			}
			// 监听会话文件
			if err := fw.watcher.Add(sessionPath); err != nil {
				log.Printf("无法监听会话文件 %s: %v", sessionPath, err)
			} else {
				log.Printf("开始监听会话文件: %s", sessionPath)
			}
		}
	}
}

// Stop 停止文件监听
func (fw *FileWatcher) Stop() {
	close(fw.stopChan)
	fw.watcher.Close()
}

// watchEvents 监听 fsnotify 事件
func (fw *FileWatcher) watchEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// 检查是否是会话文件更新
				if strings.HasSuffix(event.Name, ".jsonl") {
					fw.handleSessionUpdate(event.Name)
				} else {
					// 历史记录文件更新，广播通知
					fw.hub.Broadcast("history_updated", map[string]interface{}{
						"file": event.Name,
						"time": time.Now().Unix(),
					})
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Println("文件监听错误:", err)
		case <-fw.stopChan:
			return
		}
	}
}

// handleSessionUpdate 处理会话文件更新，读取新消息并通过 WebSocket 发送
func (fw *FileWatcher) handleSessionUpdate(sessionPath string) {
	// 提取 sessionId
	sessionId := filepath.Base(sessionPath)
	sessionId = strings.TrimSuffix(sessionId, ".jsonl")

	// 获取当前文件大小
	info, err := os.Stat(sessionPath)
	if err != nil {
		return
	}
	currentSize := info.Size()

	// 获取上次记录的文件大小
	fw.trackedSessionsMu.RLock()
	lastSize, known := fw.trackedSessions[sessionPath]
	fw.trackedSessionsMu.RUnlock()

	if !known {
		fw.trackedSessionsMu.Lock()
		fw.trackedSessions[sessionPath] = currentSize
		fw.trackedSessionsMu.Unlock()
		return
	}

	// 如果文件变小了（被重写），重新读取整个文件
	if currentSize < lastSize {
		lastSize = 0
	}

	// 如果没有新内容，返回
	if currentSize <= lastSize {
		return
	}

	// 打开文件并读取新增的内容
	file, err := os.Open(sessionPath)
	if err != nil {
		log.Printf("无法打开会话文件 %s: %v", sessionPath, err)
		return
	}
	defer file.Close()

	// 定位到上次读取的位置
	if _, err := file.Seek(lastSize, 0); err != nil {
		log.Printf("无法定位文件位置 %s: %v", sessionPath, err)
		return
	}

	// 读取新增的行
	scanner := bufio.NewScanner(file)
	newMessages := []interface{}{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析消息
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}
		newMessages = append(newMessages, msg)
	}

	// 更新文件大小记录
	fw.trackedSessionsMu.Lock()
	fw.trackedSessions[sessionPath] = currentSize
	fw.trackedSessionsMu.Unlock()

	// 如果有新消息，通过 WebSocket 发送
	if len(newMessages) > 0 {
		log.Printf("会话 %s 有 %d 条新消息", sessionId, len(newMessages))
		fw.hub.Broadcast("session_updated", map[string]interface{}{
			"sessionId": sessionId,
			"messages":  newMessages,
		})
	}
}

// pollSessions 轮询检测新会话
func (fw *FileWatcher) pollSessions() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	knownSessions := make(map[string]time.Time)

	for {
		select {
		case <-ticker.C:
			fw.scanSessions(knownSessions)
		case <-fw.stopChan:
			return
		}
	}
}

// scanSessions 扫描会话文件
func (fw *FileWatcher) scanSessions(knownSessions map[string]time.Time) {
	projectsPath := fw.cfg.Claude.ProjectsPath
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查该项目的会话文件
		projectPath := filepath.Join(projectsPath, entry.Name())
		sessions, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, session := range sessions {
			if filepath.Ext(session.Name()) != ".jsonl" {
				continue
			}

			sessionPath := filepath.Join(projectPath, session.Name())
			info, err := os.Stat(sessionPath)
			if err != nil {
				continue
			}

			// 检查是否是新会话或更新的会话
			lastMod, known := knownSessions[sessionPath]
			if !known || info.ModTime().After(lastMod) {
				sessionId := strings.TrimSuffix(session.Name(), ".jsonl")

				// 如果是新发现的会话，添加到监听
				fw.trackedSessionsMu.RLock()
				_, tracked := fw.trackedSessions[sessionPath]
				fw.trackedSessionsMu.RUnlock()

				if !tracked {
					fw.trackedSessionsMu.Lock()
					fw.trackedSessions[sessionPath] = info.Size()
					fw.trackedSessionsMu.Unlock()

					if err := fw.watcher.Add(sessionPath); err != nil {
						log.Printf("无法监听会话文件 %s: %v", sessionPath, err)
					} else {
						log.Printf("开始监听新会话文件: %s", sessionPath)
					}
				}

				// 广播通知
				fw.hub.Broadcast("new_session", map[string]interface{}{
					"sessionId": sessionId,
					"project":   projectPath,
					"time":      info.ModTime().Unix(),
				})
				knownSessions[sessionPath] = info.ModTime()
			}
		}
	}
}
