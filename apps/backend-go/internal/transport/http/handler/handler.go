package handler

import (
	"errors"
	"io"

	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func recoverServicePanic(c *gin.Context) {
	if r := recover(); r != nil {
		response.Error(c, "500000", "Service unavailable")
	}
}

func shouldBindOptionalJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
