package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/yourusername/gin-crud-api/docs"
	userv1 "github.com/yourusername/gin-crud-api/gen/go/user/v1"
	"github.com/yourusername/gin-crud-api/internal/config"
	"github.com/yourusername/gin-crud-api/internal/database"
	"github.com/yourusername/gin-crud-api/internal/router"
	"github.com/yourusername/gin-crud-api/internal/user/repository/postgres"
	userservice "github.com/yourusername/gin-crud-api/internal/user/service"
	usergrpc "github.com/yourusername/gin-crud-api/internal/user/transport/grpc"
	userhttp "github.com/yourusername/gin-crud-api/internal/user/transport/http"
	"github.com/yourusername/gin-crud-api/pkg/logger"
	"google.golang.org/grpc"
)

// @title Gin CRUD API
// @version 1.0
// @description CRUD API built with Gin and PostgreSQL.
// @BasePath /api/v1
// @schemes http https

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
	userService := userservice.NewUserService(userRepo, log)

	// Initialize handlers
	userHandler := userhttp.NewUserHandler(userService, log)

	// Initialize router
	r := router.NewRouter(userHandler, log)
	r.SetupRoutes()

	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("Failed to listen for gRPC server", "error", err)
	}

	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, usergrpc.NewServer(userService))

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r.GetEngine(),
	}

	// Start internal gRPC server for service-to-service communication.
	go func() {
		log.Info("Starting gRPC server", "address", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal("Failed to start gRPC server", "error", err)
		}
	}()

	// Start server in goroutine
	go func() {
		log.Info("Starting HTTP server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", "error", err)
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

	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	select {
	case <-grpcStopped:
	case <-ctx.Done():
		grpcServer.Stop()
	}

	log.Info("Server exited")
}
