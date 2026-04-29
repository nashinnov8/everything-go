package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
	"github.com/yourusername/gin-crud-api/internal/handler"
	"github.com/yourusername/gin-crud-api/internal/middleware"
	"github.com/yourusername/gin-crud-api/pkg/logger"
)

type Router struct {
	engine      *gin.Engine
	userHandler *handler.UserHandler
	logger      logger.Logger
}

func NewRouter(userHandler *handler.UserHandler, logger logger.Logger) *Router {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Global middleware
	engine.Use(middleware.Logger(logger))
	engine.Use(middleware.Recovery(logger))

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return &Router{
		engine:      engine,
		userHandler: userHandler,
		logger:      logger,
	}
}

// Setup Routes configures all Routes
func (r *Router) SetupRoutes() {
	// Health check
	r.engine.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.engine.Group("api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", r.userHandler.Create)
			users.GET("", r.userHandler.List)
			users.GET("/:id", r.userHandler.GetByID)
			users.PUT("/:id", r.userHandler.Update)
			users.DELETE("/:id", r.userHandler.Delete)
		}
	}
}

// GetEngine returns the gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
