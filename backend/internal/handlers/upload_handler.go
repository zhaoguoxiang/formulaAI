package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxUploadSize   = 10 << 20 // 10 MB
	uploadDir       = "uploads"
	uploadURLPrefix = "/uploads"
)

// InitUploads ensures the upload directory exists.
func InitUploads() error {
	return os.MkdirAll(uploadDir, 0o755)
}

// HandleUpload receives a multipart image upload and saves it to disk.
// POST /api/upload
func HandleUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/gif" && contentType != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type: " + contentType})
		return
	}

	// Validate file magic bytes to prevent spoofing Content-Type
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	buf = buf[:n]
	detectedType := contentTypesDetect(buf)
	allowedMagic := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedMagic[detectedType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file content does not match image format"})
		return
	}

	// Seek back to start since multipart.File implements io.Seeker
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to rewind file"})
		return
	}

	// Organize by year/month
	now := time.Now()
	subDir := filepath.Join(uploadDir, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()))
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		slog.Error("failed to create upload directory", "dir", subDir, "error", err)
		serverError(c, "failed to save file", err)
		return
	}

	ext := filepath.Ext(header.Filename)
	saveName := uuid.New().String() + ext
	savePath := filepath.Join(subDir, saveName)

	dst, err := os.Create(savePath)
	if err != nil {
		slog.Error("failed to create upload file", "path", savePath, "error", err)
		serverError(c, "failed to save file", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		slog.Error("failed to write upload file", "path", savePath, "error", err)
		serverError(c, "failed to save file", err)
		return
	}

	relativeURL := filepath.Join(uploadURLPrefix, fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()), saveName)

	slog.Info("upload received", "path", relativeURL, "size_bytes", header.Size)

	// Vditor expects this response format
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": gin.H{
			"originalURL": relativeURL,
			"url":         relativeURL,
		},
	})
}

func contentTypesDetect(data []byte) string {
	return http.DetectContentType(data)
}
