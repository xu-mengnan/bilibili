package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bilibili/internal/handlers/middleware"
	"github.com/gin-gonic/gin"
)

func TestRespondBadRequestUsesCanonicalShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/", func(c *gin.Context) {
		RespondBadRequest(c, "bad input")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "BAD_REQUEST" || body["message"] != "bad input" || body["request_id"] != "req-123" {
		t.Fatalf("unexpected response: %#v", body)
	}
	if _, exists := body["error"]; exists {
		t.Fatalf("legacy error field should not be present: %#v", body)
	}
}
