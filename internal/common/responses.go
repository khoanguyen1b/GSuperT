package common

import (
	"fmt"
	"runtime"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{Data: data})
}

func Error(c *gin.Context, statusCode int, code string, message string) {
	_, file, line, ok := runtime.Caller(1)
	callerInfo := "unknown"
	if ok {
		callerInfo = fmt.Sprintf("%s:%d", file, line)
	}

	c.Error(fmt.Errorf("[%s] %s | Caller: %s", code, message, callerInfo))
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
