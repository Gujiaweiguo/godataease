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

const (
	CodeSuccess       = "000000"
	CodeInternalError = "500000"
	CodeBadRequest    = "10001"
	CodeUnauthorized  = "20001"
	CodeTooManyReqs   = "429001"
	CodeForbidden     = "70001"
	CodeNotFound      = "50001"
	CodeForbiddenExp  = "403001"
	CodeNotFoundExp   = "404001"
	CodeServerErr     = "40001"
)

const (
	MsgAuthenticationRequired = "authentication required"
	MsgInsufficientPermission = "insufficient permissions"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
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
	Error(c, CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	c.Header("DE-GATEWAY-FLAG", "1")
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
	c.Abort()
}

func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    CodeTooManyReqs,
		Message: message,
	})
	c.Abort()
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
	c.Abort()
}

func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

func ForbiddenExport(c *gin.Context, message string) {
	Error(c, CodeForbiddenExp, message)
}

func NotFoundExport(c *gin.Context, message string) {
	Error(c, CodeNotFoundExp, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, CodeServerErr, message)
}
