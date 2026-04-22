package handler

import (
	"strconv"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/pkg/response"
	"dataease/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StoreHandler struct {
	repo *repository.FavoriteRepository
}

func NewStoreHandler(repo *repository.FavoriteRepository) *StoreHandler {
	return &StoreHandler{repo: repo}
}

type StoreExecuteRequest struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type StoreQueryRequest struct {
	Type    string `json:"type"`
	Keyword string `json:"keyword"`
}

func (h *StoreHandler) Execute(c *gin.Context) {
	defer recoverServicePanic(c)
	var req StoreExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	userID := getUserID(c)
	if userID <= 0 {
		response.Error(c, "500000", "Unauthorized")
		return
	}

	favorited, err := h.repo.IsFavorited(req.ID, userID)
	if err != nil {
		response.Error(c, "500000", "Failed to check favorite: "+err.Error())
		return
	}

	if favorited {
		if err := h.repo.DeleteFavorite(req.ID, userID); err != nil {
			response.Error(c, "500000", "Failed to remove favorite: "+err.Error())
			return
		}
	} else {
		resourceType := resourceTypeFromString(req.Type)
		store := &auto.CoreStore{
			ID:           int64(uuid.New().ID()),
			ResourceID:   req.ID,
			UID:          userID,
			ResourceType: resourceType,
			Time:         time.Now().UnixMilli(),
		}
		if err := h.repo.CreateFavorite(store); err != nil {
			response.Error(c, "500000", "Failed to add favorite: "+err.Error())
			return
		}
	}
	response.Success(c, nil)
}

func (h *StoreHandler) Favorited(c *gin.Context) {
	defer recoverServicePanic(c)
	idStr := c.Param("id")
	var id int64
	if _, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		id, _ = strconv.ParseInt(idStr, 10, 64)
	}
	if id <= 0 {
		response.Error(c, "500000", "Invalid id")
		return
	}
	userID := getUserID(c)
	if userID <= 0 {
		response.Success(c, false)
		return
	}
	favorited, err := h.repo.IsFavorited(id, userID)
	if err != nil {
		response.Error(c, "500000", "Failed to check favorite: "+err.Error())
		return
	}
	response.Success(c, favorited)
}

func (h *StoreHandler) Query(c *gin.Context) {
	defer recoverServicePanic(c)
	var req StoreQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "500000", "Invalid request: "+err.Error())
		return
	}
	userID := getUserID(c)
	if userID <= 0 {
		response.Success(c, []interface{}{})
		return
	}
	resourceType := resourceTypeFromString(req.Type)
	rows, err := h.repo.QueryFavorites(userID, resourceType, req.Keyword)
	if err != nil {
		response.Error(c, "500000", "Failed to query favorites: "+err.Error())
		return
	}
	response.Success(c, rows)
}

func RegisterStoreRoutes(r gin.IRouter, h *StoreHandler, skipQuery bool) {
	g := r.Group("/store")
	{
		g.POST("/execute", h.Execute)
		g.GET("/favorited/:id", h.Favorited)
		if !skipQuery {
			g.POST("/query", h.Query)
		}
	}
}

func getUserID(c *gin.Context) int64 {
	if uid, exists := c.Get("userId"); exists {
		if id, ok := uid.(int64); ok {
			return id
		}
	}
	return 0
}

func resourceTypeFromString(typeStr string) int32 {
	switch typeStr {
	case "PANEL", "DATA_VIZ":
		return 1
	case "SCREEN", "DATA_VIZ_SCREEN":
		return 2
	case "REPORT":
		return 3
	default:
		return 1
	}
}
