package handlers

import (
	"net/http"

	apierrors "bilibili/internal/errors"
	"bilibili/internal/handlers/middleware"
	"bilibili/pkg/utils"
	"github.com/gin-gonic/gin"
)

func RespondAPIError(c *gin.Context, apiErr *apierrors.APIError) {
	if apiErr == nil {
		apiErr = apierrors.NewInternalError("Internal server error")
	}
	response := *apiErr
	response.RequestID = middleware.GetRequestID(c)
	c.JSON(response.GetHTTPStatus(), response)
}

func RespondBadRequest(c *gin.Context, message string) {
	RespondAPIError(c, apierrors.NewBadRequest(message))
}

func RespondNotFound(c *gin.Context, message string) {
	RespondAPIError(c, apierrors.NewNotFound(message))
}

func RespondTaskNotFound(c *gin.Context, message string) {
	RespondAPIError(c, apierrors.NewAPIError(apierrors.ErrCodeTaskNotFound, message, ""))
}

func RespondTaskInvalidState(c *gin.Context, message string) {
	RespondAPIError(c, apierrors.NewAPIError(apierrors.ErrCodeTaskInvalidState, message, ""))
}

func RespondServiceUnavailable(c *gin.Context, message string) {
	RespondAPIError(c, apierrors.NewServiceUnavailable(message))
}

func RespondInternalError(c *gin.Context, message string, err error) {
	fields := map[string]interface{}{
		"request_id": middleware.GetRequestID(c),
		"path":       c.Request.URL.Path,
	}
	if err != nil {
		utils.LogErrorFields(err, fields, message)
	}
	RespondAPIError(c, apierrors.NewInternalError(message))
}

func RespondWithStatus(c *gin.Context, status int, code, message string) {
	apiErr := apierrors.NewAPIError(code, message, "")
	if status == http.StatusInternalServerError {
		apiErr.Code = apierrors.ErrCodeInternalServerError
	}
	response := *apiErr
	response.RequestID = middleware.GetRequestID(c)
	c.JSON(status, response)
}
