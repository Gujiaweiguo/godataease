package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    "000000",
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code string, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithData(c *gin.Context, code string, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, "10001", message)
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    "20001",
		Message: message,
	})
	c.Abort()
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    "70001",
		Message: message,
	})
	c.Abort()
}

func NotFound(c *gin.Context, message string) {
	Error(c, "50001", message)
}

// Java-compatible error codes for export module
func ForbiddenExport(c *gin.Context, message string) {
	Error(c, "403001", message)
}

func NotFoundExport(c *gin.Context, message string) {
	Error(c, "404001", message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, "40001", message)
}
