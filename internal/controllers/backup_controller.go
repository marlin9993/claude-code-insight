package controllers

import (
	"archive/zip"
	"github.com/marlin9993/claude-code-insight/internal/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// BackupController 备份控制器
type BackupController struct {
	cfg *config.Config
}

// NewBackupController 创建备份控制器
func NewBackupController(cfg *config.Config) *BackupController {
	return &BackupController{cfg: cfg}
}

// DownloadBackup 下载 Claude Code 历史与会话备份
func (bc *BackupController) DownloadBackup(c *gin.Context) {
	filename := fmt.Sprintf("claude-code-backup-%s.zip", time.Now().UTC().Format("20060102-150405"))

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Cache-Control", "no-store")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	if err := addFileToZip(zipWriter, bc.cfg.Claude.HistoryPath, "history.jsonl"); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = zipWriter.Close()
		return
	}

	if err := addDirToZip(zipWriter, bc.cfg.Claude.ProjectsPath, "projects"); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = zipWriter.Close()
		return
	}
}

func addFileToZip(zipWriter *zip.Writer, sourcePath, archivePath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = archivePath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func addDirToZip(zipWriter *zip.Writer, sourceDir, archiveRoot string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		archivePath := filepath.ToSlash(filepath.Join(archiveRoot, relPath))

		if info.IsDir() {
			_, err := zipWriter.Create(archivePath + "/")
			return err
		}

		return addFileToZip(zipWriter, path, archivePath)
	})
}
