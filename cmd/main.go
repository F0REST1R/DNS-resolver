package main

import (
	"context"
	"dns-resolver/internal/api"
	"dns-resolver/internal/config"
	dnsresolver "dns-resolver/internal/dns_resolver"
	"dns-resolver/internal/repository"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	logger := log.New(os.Stdout, "DNS_RESOLVER: ", log.LstdFlags|log.Lshortfile)

	db, err := config.ProdDB()
	if err != nil {
		logger.Fatalf("failed to connect DB: %v", err)
	}

	repo := repository.NewDB(db)
	resolver := dnsresolver.NewResolver(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go resolver.DNSUpdater(ctx, 5*time.Minute)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	validate := validator.New()
	e.Validator = &CustomValidator{validator: validate}

	api.NewHandler(resolver).RegisterRoutes(e)

	go func() {
		port := "8080"

		logger.Printf("Starting server on :%s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")
	cancel() // Останавливаем все горутины

	// Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server shutdown error: %v", err)
	}

	logger.Println("Server gracefully stopped")
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
