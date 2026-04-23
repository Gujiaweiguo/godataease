package handler

import (
	"errors"
	"io"

	"dataease/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func recoverServicePanic(c *gin.Context) {
	recoverServicePanicWithMessage(c, "Service unavailable")
}

func recoverServicePanicWithMessage(c *gin.Context, message string) {
	if r := recover(); r != nil {
		response.Error(c, "500000", message)
	}
}

func recoverDatasourceServicePanic(c *gin.Context) {
	recoverServicePanicWithMessage(c, "Failed: repository is unavailable")
}

func shouldBindOptionalJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindBodyWith(target, binding.JSON); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
