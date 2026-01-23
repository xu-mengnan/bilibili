package errors

import "net/http"

// APIError 统一错误响应结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// 错误码定义
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeTaskNotFound        = "TASK_NOT_FOUND"
	ErrCodeTaskInvalidState    = "TASK_INVALID_STATE"
	ErrCodeBilibiliAPI         = "BILIBILI_API_ERROR"
)

// NewAPIError 创建错误
func NewAPIError(code, message, details string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBadRequest 400 错误
func NewBadRequest(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, "")
}

// NewUnauthorized 401 错误
func NewUnauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, "")
}

// NewNotFound 404 错误
func NewNotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, "")
}

// NewConflict 409 错误
func NewConflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, "")
}

// NewInternalError 500 错误
func NewInternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternalServerError, message, "")
}

// GetHTTPStatus 获取错误对应的 HTTP 状态码
func (e *APIError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeTaskNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
