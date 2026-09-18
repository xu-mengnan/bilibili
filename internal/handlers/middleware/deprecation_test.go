package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeprecatedAPIHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/legacy", DeprecatedAPI("/api/v2/tasks"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/legacy", nil))

	if w.Header().Get("Deprecation") != "true" {
		t.Fatalf("missing Deprecation header")
	}
	if !strings.Contains(w.Header().Get("Link"), "/api/v2/tasks") {
		t.Fatalf("unexpected Link header: %q", w.Header().Get("Link"))
	}
}
