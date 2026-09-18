package errors

import "net/http"

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeTaskNotFound        = "TASK_NOT_FOUND"
	ErrCodeTaskInvalidState    = "TASK_INVALID_STATE"
	ErrCodeBilibiliAPI         = "BILIBILI_API_ERROR"
)

func NewAPIError(code, message, details string) *APIError {
	return &APIError{Code: code, Message: message, Details: details}
}

func NewBadRequest(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, "")
}

func NewUnauthorized(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, "")
}

func NewNotFound(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, "")
}

func NewConflict(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, "")
}

func NewServiceUnavailable(message string) *APIError {
	return NewAPIError(ErrCodeServiceUnavailable, message, "")
}

func NewInternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternalServerError, message, "")
}

func (e *APIError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeNotFound, ErrCodeTaskNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeTaskInvalidState:
		return http.StatusConflict
	case ErrCodeBilibiliAPI:
		return http.StatusBadGateway
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
