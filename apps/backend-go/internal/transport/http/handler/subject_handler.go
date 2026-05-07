package handler

import (
	"encoding/base64"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type SubjectHandler struct {
	repo *repository.SubjectRepository
}

func NewSubjectHandler(repo *repository.SubjectRepository) *SubjectHandler {
	return &SubjectHandler{repo: repo}
}

type SubjectRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Details    string `json:"details"`
	DeleteFlag bool   `json:"deleteFlag"`
	CoverURL   string `json:"coverUrl"`
}

func (h *SubjectHandler) Query(c *gin.Context) {
	defer recoverServicePanic(c)
	subjects, err := h.repo.List()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to query subjects: "+err.Error())
		return
	}
	response.Success(c, subjects)
}

func (h *SubjectHandler) QueryWithGroup(c *gin.Context) {
	defer recoverServicePanic(c)
	subjects, err := h.repo.ListAll()
	if err != nil {
		response.Error(c, response.CodeInternalError, "Failed to query subjects: "+err.Error())
		return
	}
	// Group into chunks of 4
	pageSize := 4
	var groups [][]auto.VisualizationSubject
	for i := 0; i < len(subjects); i += pageSize {
		end := i + pageSize
		if end > len(subjects) {
			end = len(subjects)
		}
		groups = append(groups, subjects[i:end])
	}
	if groups == nil {
		groups = [][]auto.VisualizationSubject{}
	}
	response.Success(c, groups)
}

func (h *SubjectHandler) Update(c *gin.Context) {
	defer recoverServicePanic(c)
	var req SubjectRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.Error(c, response.CodeInternalError, "Invalid request: "+err.Error())
		return
	}

	now := time.Now().UnixMilli()

	if req.ID == "" {
		// Create
		existing, err := h.repo.FindByName(req.Name)
		if err == nil && existing != nil {
			response.Error(c, response.CodeInternalError, "名称已经存在")
			return
		}
		id := base64.RawURLEncoding.EncodeToString([]byte(uuid.New().String()))[:16]
		subject := &auto.VisualizationSubject{
			ID:         id,
			Name:       req.Name,
			Type:       role.DataScopeSelf,
			Details:    req.Details,
			DeleteFlag: false,
			CoverURL:   req.CoverURL,
			CreateTime: now,
		}
		if err := h.repo.Create(subject); err != nil {
			response.Error(c, response.CodeInternalError, "Failed to create subject: "+err.Error())
			return
		}
	} else {
		// Update
		existing, err := h.repo.FindByNameExcludeID(req.Name, req.ID)
		if err == nil && existing != nil {
			response.Error(c, response.CodeInternalError, "名称已经存在")
			return
		}
		subject, err := h.repo.GetByID(req.ID)
		if err != nil {
			response.Error(c, response.CodeInternalError, "Subject not found")
			return
		}
		subject.Name = req.Name
		subject.Details = req.Details
		subject.CoverURL = req.CoverURL
		subject.UpdateTime = now
		if err := h.repo.Update(subject); err != nil {
			response.Error(c, response.CodeInternalError, "Failed to update subject: "+err.Error())
			return
		}
	}
	response.Success(c, nil)
}

func (h *SubjectHandler) Delete(c *gin.Context) {
	defer recoverServicePanic(c)
	id := c.Param("id")
	if id == "" {
		response.Error(c, response.CodeInternalError, errInvalidID)
		return
	}
	if err := h.repo.Delete(id); err != nil {
		response.Error(c, response.CodeInternalError, "Failed to delete subject: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func RegisterSubjectRoutes(r gin.IRouter, h *SubjectHandler) {
	g := r.Group("/visualizationSubject")
	{
		g.POST("/query", h.Query)
		g.POST("/querySubjectWithGroup", h.QueryWithGroup)
		g.POST("/update", h.Update)
		g.POST("/delete/:id", h.Delete)
	}
}
