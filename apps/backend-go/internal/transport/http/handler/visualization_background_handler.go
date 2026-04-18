package handler

import (
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type VisualizationBackgroundHandler struct {
	repo *repository.VisualizationBackgroundRepository
}

func NewVisualizationBackgroundHandler(repo *repository.VisualizationBackgroundRepository) *VisualizationBackgroundHandler {
	return &VisualizationBackgroundHandler{repo: repo}
}

// BackgroundVO is the response VO matching Java VisualizationBackgroundVO.
type BackgroundVO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Classification string `json:"classification"`
	Content        string `json:"content"`
	Remark         string `json:"remark"`
	Sort           int32  `json:"sort"`
	UploadTime     int64  `json:"uploadTime"`
	BaseURL        string `json:"baseUrl"`
	URL            string `json:"url"`
}

func (h *VisualizationBackgroundHandler) FindAll(c *gin.Context) {
	defer recoverServicePanic(c)
	backgrounds, err := h.repo.FindAll()
	if err != nil {
		response.Error(c, "500000", "Failed to query backgrounds: "+err.Error())
		return
	}

	result := make(map[string][]BackgroundVO)
	for _, bg := range backgrounds {
		vo := BackgroundVO{
			ID:             bg.ID,
			Name:           bg.Name,
			Classification: bg.Classification,
			Content:        bg.Content,
			Remark:         bg.Remark,
			Sort:           bg.Sort,
			UploadTime:     bg.UploadTime,
			BaseURL:        bg.BaseURL,
			URL:            bg.URL,
		}
		result[bg.Classification] = append(result[bg.Classification], vo)
	}
	response.Success(c, result)
}

func RegisterVisualizationBackgroundRoutes(r gin.IRouter, h *VisualizationBackgroundHandler) {
	g := r.Group("/visualizationBackground")
	{
		g.GET("/findAll", h.FindAll)
	}
}
