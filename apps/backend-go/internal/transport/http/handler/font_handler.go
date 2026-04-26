package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

const defaultFontDir = "/opt/dataease2.0/data/font/"

type FontHandler struct {
	repo *repository.TypefaceRepository
}

func NewFontHandler(repo *repository.TypefaceRepository) *FontHandler {
	return &FontHandler{repo: repo}
}

type FontDTO struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	FileName      string  `json:"fileName"`
	FileTransName string  `json:"fileTransName"`
	IsDefault     bool    `json:"isDefault"`
	IsBuiltIn     bool    `json:"isBuiltin"`
	Size          float64 `json:"size"`
	SizeType      string  `json:"sizeType"`
}

func (h *FontHandler) List(c *gin.Context) {
	defer recoverServicePanic(c)
	fonts, err := h.repo.ListFonts()
	if err != nil {
		response.Error(c, "500000", "Failed to list fonts: "+err.Error())
		return
	}
	dtos := make([]FontDTO, 0, len(fonts))
	for _, f := range fonts {
		dtos = append(dtos, fontToDTO(&f))
	}
	response.Success(c, dtos)
}

func (h *FontHandler) Create(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto FontDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	existing, err := h.repo.FindFontByName(dto.Name)
	if err == nil && existing != nil {
		response.Error(c, "500000", "存在重名字库")
		return
	}
	font := &auto.CoreFont{
		ID:            int64(uuid.New().ID()),
		Name:          dto.Name,
		FileName:      dto.FileName,
		FileTransName: dto.FileTransName,
		IsDefault:     dto.IsDefault,
		IsBuiltIn:     false,
		Size:          dto.Size,
		SizeType:      dto.SizeType,
		UpdateTime:    time.Now().UnixMilli(),
	}
	if err := h.repo.CreateFont(font); err != nil {
		response.Error(c, "500000", "Failed to create font: "+err.Error())
		return
	}
	response.Success(c, fontToDTO(font))
}

func (h *FontHandler) Edit(c *gin.Context) {
	defer recoverServicePanic(c)
	var dto FontDTO
	if err := c.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	if dto.ID == 0 {
		h.Create(c)
		return
	}
	font, err := h.repo.GetFontByID(dto.ID)
	if err != nil {
		response.Error(c, "500000", "Font not found")
		return
	}
	font.Name = dto.Name
	font.FileName = dto.FileName
	font.FileTransName = dto.FileTransName
	font.IsDefault = dto.IsDefault
	font.Size = dto.Size
	font.SizeType = dto.SizeType
	font.UpdateTime = time.Now().UnixMilli()

	if dto.IsDefault {
		_ = h.repo.ClearDefaultFonts(dto.ID)
	}
	if err := h.repo.UpdateFont(font); err != nil {
		response.Error(c, "500000", "Failed to update font: "+err.Error())
		return
	}
	response.Success(c, fontToDTO(font))
}

func (h *FontHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	idStr := c.Param("id")
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
		response.Error(c, "500000", "Invalid id")
		return
	}
	font, err := h.repo.GetFontByID(id)
	if err != nil {
		response.Error(c, "500000", "Font not found")
		return
	}
	if font.FileTransName != "" {
		fontDir := os.Getenv("FONT_DIR")
		if fontDir == "" {
			fontDir = "/opt/dataease2.0/data/font/"
		}
		_ = os.Remove(filepath.Join(fontDir, font.FileTransName))
	}
	if err := h.repo.DeleteFont(id); err != nil {
		response.Error(c, "500000", "Failed to delete font: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *FontHandler) UploadFile(c *gin.Context) {
	defer recoverServicePanic(c)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "500000", "Failed to read upload file: "+err.Error())
		return
	}
	defer file.Close()

	filename := header.Filename
	if filename == "" || !strings.HasSuffix(strings.ToLower(filename), ".ttf") {
		response.Error(c, "500000", "非法格式的文件！Only .ttf files are supported")
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, "500000", "Failed to read file: "+err.Error())
		return
	}

	fontDir := os.Getenv("FONT_DIR")
	if fontDir == "" {
		fontDir = defaultFontDir
	}
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		response.Error(c, "500000", "Failed to create font directory: "+err.Error())
		return
	}

	fileUUID := uuid.New().String()
	ext := ".ttf"
	fileTransName := fileUUID + ext
	destPath := filepath.Join(fontDir, fileTransName)
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		response.Error(c, "500000", "Failed to save font file: "+err.Error())
		return
	}

	length := int64(len(content))
	var size float64
	var sizeType string
	if float64(length)/1024/1024 > 1 {
		if float64(length)/1024/1024/1024 > 1 {
			sizeType = "GB"
			size = float64(length) / 1024 / 1024 / 1024
		} else {
			sizeType = "MB"
			size = float64(length) / 1024 / 1024
		}
	} else {
		sizeType = "KB"
		size = float64(length) / 1024
	}
	size = float64(int(size*100)) / 100

	fontName := strings.TrimSuffix(filename, filepath.Ext(filename))

	result := FontDTO{
		FileTransName: fileTransName,
		FileName:      filename,
		Name:          fontName,
		Size:          size,
		SizeType:      sizeType,
	}
	response.Success(c, result)
}

func (h *FontHandler) DefaultFont(c *gin.Context) {
	defer recoverServicePanic(c)
	fonts, err := h.repo.ListDefaultFonts()
	if err != nil {
		response.Error(c, "500000", "Failed to list default fonts: "+err.Error())
		return
	}
	dtos := make([]FontDTO, 0, len(fonts))
	for _, f := range fonts {
		dtos = append(dtos, fontToDTO(&f))
	}
	response.Success(c, dtos)
}

func (h *FontHandler) Download(c *gin.Context) {
	defer recoverServicePanic(c)
	fileName := c.Param("file")
	if fileName == "" {
		response.Error(c, "500000", "Invalid file name")
		return
	}
	cleanName := filepath.Clean(fileName)
	if cleanName != fileName || strings.Contains(cleanName, "..") {
		response.Error(c, "500000", "Invalid file name")
		return
	}
	fontDir := os.Getenv("FONT_DIR")
	if fontDir == "" {
		fontDir = defaultFontDir
	}
	fontDir = filepath.Clean(fontDir)
	filePath := filepath.Join(fontDir, cleanName)
	if !strings.HasPrefix(filePath, fontDir+string(os.PathSeparator)) && filePath != fontDir {
		response.Error(c, "500000", "Invalid file name")
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error(c, "500000", "Font file not found")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+cleanName)
	c.File(filePath)
}

func RegisterFontRoutes(r gin.IRouter, h *FontHandler) {
	g := r.Group("/typeface")
	{
		g.GET("/listFont", h.List)
		g.POST("/create", h.Create)
		g.POST("/edit", h.Edit)
		g.POST("/delete/:id", h.Delete)
		g.POST("/uploadFile", h.UploadFile)
		g.GET("/defaultFont", h.DefaultFont)
	}
}

func RegisterFontDownloadRoute(r gin.IRouter, h *FontHandler) {
	g := r.Group("/typeface")
	{
		g.GET("/download/:file", h.Download)
	}
}

func fontToDTO(f *auto.CoreFont) FontDTO {
	return FontDTO{
		ID:            f.ID,
		Name:          f.Name,
		FileName:      f.FileName,
		FileTransName: f.FileTransName,
		IsDefault:     f.IsDefault,
		IsBuiltIn:     f.IsBuiltIn,
		Size:          f.Size,
		SizeType:      f.SizeType,
	}
}
