package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bilibili/api"
	"bilibili/pkg/utils"
)

func main() {
	utils.LogInfo("Starting Bilibili Comment Scraper...")

	// 创建 root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置路由
	router, services := api.SetupRoutes(ctx)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动 HTTP 服务器（goroutine）
	errChan := make(chan error, 1)
	go func() {
		utils.LogInfo("Server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed: %w", err)
		}
	}()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		utils.LogError("Server error: " + err.Error())
		os.Exit(1)
	case sig := <-sigChan:
		utils.LogInfo("Received signal: " + sig.String())
	}

	// 优雅关闭
	utils.LogInfo("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 1. 关闭 HTTP 服务器（不再接受新请求）
	utils.LogInfo("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.LogError("HTTP server shutdown error: " + err.Error())
	}

	// 2. 关闭所有服务
	api.ShutdownServices(shutdownCtx, services)

	utils.LogInfo("Shutdown complete")
}
