package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-request-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	if w.Header().Get("X-Request-ID") != "client-request-123" {
		t.Fatalf("response request id=%q", w.Header().Get("X-Request-ID"))
	}
	if w.Body.String() != "client-request-123" {
		t.Fatalf("context request id=%q", w.Body.String())
	}
}

func TestRequestIDRejectsUnsafeClientValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "unsafe request id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() == "unsafe request id" || w.Body.String() == "" {
		t.Fatalf("unsafe request id was not replaced: %q", w.Body.String())
	}
}
