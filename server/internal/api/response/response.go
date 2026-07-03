package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Success 成功响应
func Success(c echo.Context, statusCode int, data interface{}) error {
	return c.JSON(statusCode, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		Timestamp: currentTimestamp(),
	})
}

// Error 错误响应
func Error(c echo.Context, statusCode int, message string) error {
	code := statusCodeToCode(statusCode)
	return c.JSON(statusCode, Response{
		Code:      code,
		Message:   message,
		Timestamp: currentTimestamp(),
	})
}

// ErrorWithCode 带自定义错误码的错误响应
func ErrorWithCode(c echo.Context, statusCode int, errCode int, message string) error {
	return c.JSON(statusCode, Response{
		Code:      errCode,
		Message:   message,
		Timestamp: currentTimestamp(),
	})
}

// 分页响应
type PaginatedResponse struct {
	Items    interface{} `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	HasMore  bool        `json:"has_more"`
}

// PaginatedSuccess 分页成功响应
func PaginatedSuccess(c echo.Context, items interface{}, total, page, pageSize int) error {
	hasMore := page*pageSize < total
	return Success(c, http.StatusOK, &PaginatedResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	})
}

// 当前时间戳
func currentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// 状态码转错误码
func statusCodeToCode(statusCode int) int {
	switch {
	case statusCode >= 400 && statusCode < 500:
		return statusCode - 3000
	case statusCode >= 500:
		return statusCode - 1000
	default:
		return 0
	}
}
