package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type StaticHandler struct {
	service *service.StaticService
}

const (
	maxStaticResourcePathCount = 200
	maxStaticResourceFileBytes = 10 << 20
)

func resolveSafeStaticUploadPath(staticDir, fileID, ext string) (string, bool) {
	if fileID == "" {
		return "", false
	}
	cleanID := filepath.Clean(fileID)
	if cleanID != fileID || strings.Contains(cleanID, "..") || filepath.IsAbs(cleanID) || strings.ContainsAny(fileID, `/\\`) {
		return "", false
	}
	staticDir = filepath.Clean(staticDir)
	fileName := cleanID + ext
	destPath := filepath.Join(staticDir, fileName)
	if !strings.HasPrefix(destPath, staticDir+string(os.PathSeparator)) && destPath != staticDir {
		return "", false
	}
	return destPath, true
}

func NewStaticHandler(service *service.StaticService) *StaticHandler {
	return &StaticHandler{service: service}
}

func (h *StaticHandler) ListResources(c *gin.Context) {
	defer recoverServicePanic(c)

	result, err := h.service.ListResources()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) GetResource(c *gin.Context) {
	defer recoverServicePanic(c)

	id := c.Param("id")
	result, err := h.service.GetResource(id)
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) ListStores(c *gin.Context) {
	defer recoverServicePanic(c)

	result, err := h.service.ListStores()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

func (h *StaticHandler) ListTypefaces(c *gin.Context) {
	defer recoverServicePanic(c)

	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

// ListFont returns font list for frontend compatibility
func (h *StaticHandler) ListFont(c *gin.Context) {
	defer recoverServicePanic(c)

	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	response.Success(c, result)
}

// DefaultFont returns default font for frontend compatibility
func (h *StaticHandler) DefaultFont(c *gin.Context) {
	defer recoverServicePanic(c)

	result, err := h.service.ListTypefaces()
	if err != nil {
		response.Error(c, "500000", err.Error())
		return
	}
	// Return first font as default, or empty object if none
	if len(result) > 0 {
		response.Success(c, result[0])
		return
	}
	response.Success(c, map[string]interface{}{})
}

// XpackModel returns xpack model status
func (h *StaticHandler) XpackModel(c *gin.Context) {
	defer recoverServicePanic(c)
	response.Success(c, false)
}

func (h *StaticHandler) Upload(c *gin.Context) {
	defer recoverServicePanic(c)
	fileID := c.Param("fileId")
	if fileID == "" {
		response.Error(c, "500000", "fileId is required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "500000", "Failed to read upload file: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".gif": true, ".svg": true, ".png": true, ".jpeg": true, ".jpg": true}
	if !allowedExts[ext] {
		response.Error(c, "500000", "File type not allowed. Only gif, svg, png, jpeg, jpg are supported")
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		response.Error(c, "500000", "Failed to read file content: "+err.Error())
		return
	}

	if ext == ".svg" {
		svgStr := string(content)
		if strings.Contains(strings.ToUpper(svgStr), "<!DOCTYPE") ||
			strings.Contains(strings.ToUpper(svgStr), "<!ENTITY") {
			response.Error(c, "500000", "SVG with DOCTYPE/ENTITY is not allowed for security reasons")
			return
		}
	}

	staticDir := os.Getenv("STATIC_RESOURCE_DIR")
	if staticDir == "" {
		staticDir = "/opt/dataease2.0/data/static-resource"
	}
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		response.Error(c, "500000", "Failed to create static directory: "+err.Error())
		return
	}

	destPath, ok := resolveSafeStaticUploadPath(staticDir, fileID, ext)
	if !ok {
		response.Error(c, "500000", "Invalid fileId")
		return
	}
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		response.Error(c, "500000", "Failed to save file: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *StaticHandler) FindResourceAsBase64(c *gin.Context) {
	defer recoverServicePanic(c)
	var req struct {
		ResourcePathList []string `json:"resourcePathList"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if len(req.ResourcePathList) > maxStaticResourcePathCount {
		response.Error(c, "500000", fmt.Sprintf("too many files requested (max %d)", maxStaticResourcePathCount))
		return
	}

	staticDir := os.Getenv("STATIC_RESOURCE_DIR")
	if staticDir == "" {
		staticDir = "/opt/dataease2.0/data/static-resource"
	}

	result := make(map[string]string)
	for _, path := range req.ResourcePathList {
		fileName := path
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			fileName = path[idx+1:]
		}
		cleanName := filepath.Clean(fileName)
		if cleanName != fileName || strings.Contains(cleanName, "..") {
			result[path] = ""
			continue
		}
		filePath := filepath.Join(staticDir, cleanName)
		file, err := os.Open(filePath)
		if err != nil {
			result[path] = ""
			continue
		}
		content, err := io.ReadAll(io.LimitReader(file, maxStaticResourceFileBytes+1))
		_ = file.Close()
		if err != nil {
			result[path] = ""
			continue
		}
		if len(content) > maxStaticResourceFileBytes {
			result[path] = ""
			continue
		}
		result[path] = base64.StdEncoding.EncodeToString(content)
	}
	response.Success(c, result)
}

func RegisterStaticRoutes(r *gin.RouterGroup, h *StaticHandler) {
	staticGroup := r.Group("/staticResource")
	{
		staticGroup.GET("/list", h.ListResources)
		staticGroup.GET("/:id", h.GetResource)
		staticGroup.POST("/upload/:fileId", h.Upload)
		staticGroup.POST("/findResourceAsBase64", h.FindResourceAsBase64)
	}

	storeGroup := r.Group("/store")
	{
		storeGroup.GET("/list", h.ListStores)
	}

	// Xpack model endpoint
	r.GET("/xpackModel", h.XpackModel)
}
