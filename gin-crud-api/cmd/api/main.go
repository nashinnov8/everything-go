package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/gin-crud-api/internal/config"
	"github.com/yourusername/gin-crud-api/internal/database"
	"github.com/yourusername/gin-crud-api/internal/handler"
	"github.com/yourusername/gin-crud-api/internal/repository/postgres"
	"github.com/yourusername/gin-crud-api/internal/router"
	"github.com/yourusername/gin-crud-api/internal/service"
	"github.com/yourusername/gin-crud-api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\\n", err)
		os.Exit(1)
	}

	// Initialize Logger
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\\n", err)
		os.Exit(1)
	}

	db, err := database.NewPostgresConnection(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Dbname:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})

	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	defer database.Close(db)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, log)

	// Initialize handlers
	userHanlder := handler.NewUserHandler(userService, log)

	// Initialize router
	r := router.NewRouter(userHanlder, log)
	r.SetupRoutes()

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r.GetEngine(),
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}
