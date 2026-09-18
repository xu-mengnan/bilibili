package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"bilibili/api"
	"bilibili/internal/config"
	"bilibili/pkg/utils"
)

func main() {
	utils.LogInfo("Starting Bilibili Comment Scraper...")

	cfg, err := config.LoadDefault()
	if err != nil {
		utils.LogError("Failed to load config, using safe defaults: " + err.Error())
		cfg = config.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router, services := api.SetupRoutes(ctx)
	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		// Streaming analysis can legitimately exceed minutes. Per-request
		// cancellation/provider timeouts bound SSE work instead of WriteTimeout.
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		utils.LogInfo("Server listening on " + addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		utils.LogError("Server error: " + err.Error())
		os.Exit(1)
	case sig := <-sigChan:
		utils.LogInfo("Received signal: " + sig.String())
	}

	utils.LogInfo("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	utils.LogInfo("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.LogError("HTTP server shutdown error: " + err.Error())
	}

	api.ShutdownServices(shutdownCtx, services)
	utils.LogInfo("Shutdown complete")
}
