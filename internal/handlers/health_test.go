package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bilibili/internal/services"
	"bilibili/pkg/storage"
	"github.com/gin-gonic/gin"
)

func TestHealthLiveAndReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	commentService := services.NewCommentService(ctx, storage.NewJSONStorage(t.TempDir()))
	exportService := services.NewExportService(ctx, t.TempDir())
	analysisService := services.NewAnalysisService("", "", "")

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_ = exportService.Shutdown(shutdownCtx)
		_ = commentService.Shutdown(shutdownCtx)
	}()

	handler := NewHealthHandler(commentService, exportService, analysisService)
	r := gin.New()
	r.GET("/live", handler.Live)
	r.GET("/ready", handler.Ready)

	live := httptest.NewRecorder()
	r.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("live status=%d body=%s", live.Code, live.Body.String())
	}

	ready := httptest.NewRecorder()
	r.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if ready.Code != http.StatusOK {
		t.Fatalf("ready status=%d body=%s", ready.Code, ready.Body.String())
	}
}

func TestReadinessFailsForUnwritableStoragePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	blockingFile := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blockingFile, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	commentService := services.NewCommentService(ctx, storage.NewJSONStorage(blockingFile))
	exportService := services.NewExportService(ctx, t.TempDir())
	analysisService := services.NewAnalysisService("", "", "")
	handler := NewHealthHandler(commentService, exportService, analysisService)

	r := gin.New()
	r.GET("/live", handler.Live)
	r.GET("/ready", handler.Ready)

	live := httptest.NewRecorder()
	r.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness must not depend on storage: %d", live.Code)
	}

	ready := httptest.NewRecorder()
	r.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready status=%d body=%s", ready.Code, ready.Body.String())
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
	defer shutdownCancel()
	_ = exportService.Shutdown(shutdownCtx)
	_ = commentService.Shutdown(shutdownCtx)
}
