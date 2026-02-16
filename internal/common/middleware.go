package common

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func ErrorLogger() gin.HandlerFunc {
	// Tạo thư mục logs nếu chưa có
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.MkdirAll(logDir, 0755)
	}

	return func(c *gin.Context) {
		// Đọc Body request
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		startTime := time.Now()

		// Tiếp tục xử lý API
		c.Next()

		// Chỉ ghi log nếu lỗi (status >= 400)
		status := c.Writer.Status()
		if status >= 400 {
			// Tạo file log theo ngày: logs/trace_2024-02-16.log
			fileName := fmt.Sprintf("trace_%s.log", time.Now().Format("2006-01-02"))
			filePath := filepath.Join(logDir, fileName)

			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Không thể mở file log: %v", err)
				return
			}
			defer f.Close()

			latency := time.Since(startTime)
			method := c.Request.Method
			path := c.Request.URL.Path
			clientIP := c.ClientIP()

			// Format dòng log trace
			traceLog := fmt.Sprintf("[%s] TRACE | %s | %d | %v | %s | %s | Body: %s\n",
				time.Now().Format("15:04:05"),
				clientIP,
				status,
				latency,
				method,
				path,
				string(body),
			)

			// Ghi log lỗi chi tiết nếu có
			if len(c.Errors) > 0 {
				traceLog += fmt.Sprintf("[%s] DETAILS: %s\n", 
					time.Now().Format("15:04:05"), 
					c.Errors.String(),
				)
			}

			// Nếu là lỗi 500, ghi thêm Stack Trace để debug sâu
			if status == 500 {
				traceLog += fmt.Sprintf("[%s] STACK TRACE:\n%s\n", 
					time.Now().Format("15:04:05"),
					string(debug.Stack()),
				)
			}
			traceLog += "----------------------------------------------------------------------\n"

			// Ghi vào file
			if _, err := f.WriteString(traceLog); err != nil {
				log.Printf("Lỗi ghi file log: %v", err)
			}
			
			// Vẫn in ra console để xem qua docker logs
			fmt.Print(traceLog)
		}
	}
}
