package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"docqa/internal/handlers"
	"docqa/internal/llm"
	"docqa/internal/python"
	"docqa/internal/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	embedURL := getEnv("EMBED_SERVICE_URL", "http://embed:8000")
	uploadDir := getEnv("UPLOAD_DIR", "/data/uploads")
	port := getEnv("PORT", "8080")

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		slog.Error("failed to create upload dir", "err", err)
		os.Exit(1)
	}

	pythonClient := python.NewClient(embedURL)
	llmClient := llm.NewClient()

	h := handlers.New(pythonClient, llmClient, uploadDir)

	r := router.New(h)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Gracefull shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
