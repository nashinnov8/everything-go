# Building a CRUD API with Gin and PostgreSQL in Go

A comprehensive guide to building production-ready REST APIs using Gin framework and PostgreSQL, following Go best practices.

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Prerequisites](#prerequisites)
3. [Setting Up the Project](#setting-up-the-project)
4. [Database Layer](#database-layer)
5. [Repository Pattern](#repository-pattern)
6. [Service Layer](#service-layer)
7. [Handler Layer](#handler-layer)
8. [Routing and Middleware](#routing-and-middleware)
9. [Configuration Management](#configuration-management)
10. [Validation](#validation)
11. [Error Handling](#error-handling)
12. [Logging](#logging)
13. [Testing](#testing)
14. [Complete Code Examples](#complete-code-examples)

---

## Project Structure

Following the **Clean Architecture** and **Domain-Driven Design** principles:

```
project-root/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── domain/                  # Business entities
│   │   └── user.go
│   ├── repository/              # Database operations
│   │   ├── repository.go        # Interface definitions
│   │   └── postgres/
│   │       └── user_repository.go
│   ├── service/                 # Business logic
│   │   ├── service.go           # Interface definitions
│   │   └── user_service.go
│   ├── handler/                 # HTTP handlers
│   │   ├── handler.go           # Common handler utilities
│   │   └── user_handler.go
│   ├── middleware/              # Custom middleware
│   │   └── error_handler.go
│   └── database/                # Database connection
│       └── postgres.go
├── pkg/
│   ├── validator/               # Custom validators
│   │   └── validator.go
│   ├── logger/                  # Logger utilities
│   │   └── logger.go
│   └── errors/                  # Custom errors
│       └── errors.go
├── migrations/                  # Database migrations
│   └── 001_create_users_table.sql
├── tests/                       # Integration tests
├── .env                         # Environment variables
├── .env.example                 # Environment template
├── go.mod
├── go.sum
├── Makefile                     # Build automation
└── README.md
```

---

## Prerequisites

### Required Tools

```bash
# Go 1.21 or later
go version

# PostgreSQL installed and running
psql --version

# migrate CLI for database migrations
# https://github.com/golang-migrate/migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Dependencies

```bash
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/spf13/viper
go get -u github.com/go-playground/validator/v10
go get -u go.uber.org/zap
go get -u github.com/stretchr/testify
```

---

## Setting Up the Project

### 1. Initialize Go Module

```bash
mkdir gin-crud-api
cd gin-crud-api
go mod init github.com/yourusername/gin-crud-api
```

### 2. Environment Configuration (.env)

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_MODE=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=gin_crud_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

---

## Database Layer

### PostgreSQL Connection (`internal/database/postgres.go`)

```go
package database

import (
    "context"
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(cfg Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB: %w", err)
    }

    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

    // Verify connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := sqlDB.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

---

## Repository Pattern

### Domain Model (`internal/domain/user.go`)

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

// User represents a user entity
type User struct {
    ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email     string     `json:"email" gorm:"uniqueIndex;not null"`
    Username  string     `json:"username" gorm:"uniqueIndex;not null"`
    Password  string     `json:"-" gorm:"not null"` // Never expose in JSON
    FullName  string     `json:"full_name" gorm:"not null"`
    IsActive  bool       `json:"is_active" gorm:"default:true"`
    CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (User) TableName() string {
    return "users"
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=8"`
    FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
    Email    string `json:"email" validate:"omitempty,email"`
    Username string `json:"username" validate:"omitempty,min=3,max=50"`
    FullName string `json:"full_name" validate:"omitempty,min=2,max=100"`
    IsActive *bool  `json:"is_active"`
}

// UserResponse represents user response (excludes sensitive fields)
type UserResponse struct {
    ID        uuid.UUID `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    FullName  string    `json:"full_name"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
    return UserResponse{
        ID:        u.ID,
        Email:     u.Email,
        Username:  u.Username,
        FullName:  u.FullName,
        IsActive:  u.IsActive,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}
```

### Repository Interface (`internal/repository/repository.go`)

```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/yourusername/gin-crud-api/internal/domain"
)

// UserRepository defines user repository interface
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    GetByUsername(ctx context.Context, username string) (*domain.User, error)
    List(ctx context.Context, offset, limit int) ([]domain.User, int64, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
}
```

### PostgreSQL Implementation (`internal/repository/postgres/user_repository.go`)

```go
package postgres

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/yourusername/gin-crud-api/internal/domain"
    "github.com/yourusername/gin-crud-api/internal/repository"
)

// userRepository implements UserRepository
type userRepository struct {
    db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// GetByID gets user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

// GetByEmail gets user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    var user domain.User
    if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

// GetByUsername gets user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    var user domain.User
    if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

// List lists users with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
    var users []domain.User
    var total int64

    db := r.db.WithContext(ctx).Model(&domain.User{})

    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft-deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// ExistsByEmail checks if email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}

// ExistsByUsername checks if username exists
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.User{}).Where("username = ?", username).Count(&count).Error
    return count > 0, err
}
```

---

## Service Layer

### Service Interface (`internal/service/service.go`)

```go
package service

import (
    "context"

    "github.com/google/uuid"
    "github.com/yourusername/gin-crud-api/internal/domain"
)

// UserService defines user service interface
type UserService interface {
    Create(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
    List(ctx context.Context, page, pageSize int) (*UserListResponse, error)
    Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

// UserListResponse represents paginated user list
type UserListResponse struct {
    Users      []domain.UserResponse `json:"users"`
    Total      int64                 `json:"total"`
    Page       int                   `json:"page"`
    PageSize   int                   `json:"page_size"`
    TotalPages int                   `json:"total_pages"`
}
```

### Service Implementation (`internal/service/user_service.go`)

```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"

    "github.com/yourusername/gin-crud-api/internal/domain"
    "github.com/yourusername/gin-crud-api/internal/repository"
    "github.com/yourusername/gin-crud-api/pkg/errors"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// userService implements UserService
type userService struct {
    userRepo repository.UserRepository
    logger   logger.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, logger logger.Logger) UserService {
    return &userService{
        userRepo: userRepo,
        logger:   logger,
    }
}

// Create creates a new user
func (s *userService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error) {
    // Check if email exists
    exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        s.logger.Error("failed to check email existence", "error", err)
        return nil, errors.ErrInternalServer
    }
    if exists {
        return nil, errors.ErrDuplicateEmail
    }

    // Check if username exists
    exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
    if err != nil {
        s.logger.Error("failed to check username existence", "error", err)
        return nil, errors.ErrInternalServer
    }
    if exists {
        return nil, errors.ErrDuplicateUsername
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        s.logger.Error("failed to hash password", "error", err)
        return nil, errors.ErrInternalServer
    }

    user := &domain.User{
        Email:    req.Email,
        Username: req.Username,
        Password: string(hashedPassword),
        FullName: req.FullName,
        IsActive: true,
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        s.logger.Error("failed to create user", "error", err)
        return nil, errors.ErrInternalServer
    }

    response := user.ToResponse()
    return &response, nil
}

// GetByID gets user by ID
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to get user", "error", err)
        return nil, errors.ErrInternalServer
    }
    if user == nil {
        return nil, errors.ErrUserNotFound
    }

    response := user.ToResponse()
    return &response, nil
}

// List lists users with pagination
func (s *userService) List(ctx context.Context, page, pageSize int) (*UserListResponse, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 10
    }
    if pageSize > 100 {
        pageSize = 100
    }

    offset := (page - 1) * pageSize

    users, total, err := s.userRepo.List(ctx, offset, pageSize)
    if err != nil {
        s.logger.Error("failed to list users", "error", err)
        return nil, errors.ErrInternalServer
    }

    totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

    responses := make([]domain.UserResponse, len(users))
    for i, user := range users {
        responses[i] = user.ToResponse()
    }

    return &UserListResponse{
        Users:      responses,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

// Update updates a user
func (s *userService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to get user", "error", err)
        return nil, errors.ErrInternalServer
    }
    if user == nil {
        return nil, errors.ErrUserNotFound
    }

    // Check email uniqueness if being updated
    if req.Email != "" && req.Email != user.Email {
        exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
        if err != nil {
            s.logger.Error("failed to check email existence", "error", err)
            return nil, errors.ErrInternalServer
        }
        if exists {
            return nil, errors.ErrDuplicateEmail
        }
        user.Email = req.Email
    }

    // Check username uniqueness if being updated
    if req.Username != "" && req.Username != user.Username {
        exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
        if err != nil {
            s.logger.Error("failed to check username existence", "error", err)
            return nil, errors.ErrInternalServer
        }
        if exists {
            return nil, errors.ErrDuplicateUsername
        }
        user.Username = req.Username
    }

    if req.FullName != "" {
        user.FullName = req.FullName
    }

    if req.IsActive != nil {
        user.IsActive = *req.IsActive
    }

    if err := s.userRepo.Update(ctx, user); err != nil {
        s.logger.Error("failed to update user", "error", err)
        return nil, errors.ErrInternalServer
    }

    response := user.ToResponse()
    return &response, nil
}

// Delete deletes a user
func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to get user", "error", err)
        return errors.ErrInternalServer
    }
    if user == nil {
        return errors.ErrUserNotFound
    }

    if err := s.userRepo.Delete(ctx, id); err != nil {
        s.logger.Error("failed to delete user", "error", err)
        return errors.ErrInternalServer
    }

    return nil
}
```

---

## Handler Layer

### Common Handler (`internal/handler/handler.go`)

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/gin-crud-api/pkg/errors"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// BaseHandler provides common handler functionality
type BaseHandler struct {
    logger logger.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(logger logger.Logger) BaseHandler {
    return BaseHandler{logger: logger}
}

// Response represents API response
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
    Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// MetaInfo represents metadata for list responses
type MetaInfo struct {
    Page       int   `json:"page,omitempty"`
    PageSize   int   `json:"page_size,omitempty"`
    Total      int64 `json:"total,omitempty"`
    TotalPages int   `json:"total_pages,omitempty"`
}

// Success responds with success
func (h *BaseHandler) Success(c *gin.Context, statusCode int, data interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Data:    data,
    })
}

// SuccessWithMeta responds with success and metadata
func (h *BaseHandler) SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta *MetaInfo) {
    c.JSON(statusCode, Response{
        Success: true,
        Data:    data,
        Meta:    meta,
    })
}

// Error responds with error
func (h *BaseHandler) Error(c *gin.Context, err error) {
    appErr := errors.GetError(err)
    c.JSON(appErr.StatusCode, Response{
        Success: false,
        Error: &ErrorInfo{
            Code:    appErr.Code,
            Message: appErr.Message,
        },
    })
}

// BindAndValidate binds and validates request
func (h *BaseHandler) BindAndValidate(c *gin.Context, req interface{}) bool {
    if err := c.ShouldBindJSON(req); err != nil {
        h.Error(c, errors.ErrInvalidRequest)
        return false
    }
    return true
}
```

### User Handler (`internal/handler/user_handler.go`)

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/yourusername/gin-crud-api/internal/domain"
    "github.com/yourusername/gin-crud-api/internal/service"
    "github.com/yourusername/gin-crud-api/pkg/errors"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// UserHandler handles user-related requests
type UserHandler struct {
    BaseHandler
    userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService, logger logger.Logger) *UserHandler {
    return &UserHandler{
        BaseHandler: NewBaseHandler(logger),
        userService: userService,
    }
}

// Create godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.CreateUserRequest true "User creation request"
// @Success 201 {object} Response{data=domain.UserResponse}
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
    var req domain.CreateUserRequest
    if !h.BindAndValidate(c, &req) {
        return
    }

    user, err := h.userService.Create(c.Request.Context(), req)
    if err != nil {
        h.Error(c, err)
        return
    }

    h.Success(c, http.StatusCreated, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get a user by their UUID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} Response{data=domain.UserResponse}
// @Failure 404 {object} Response
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.Error(c, errors.ErrInvalidUserID)
        return
    }

    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        h.Error(c, err)
        return
    }

    h.Success(c, http.StatusOK, user)
}

// List godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} Response{data=[]domain.UserResponse,meta=MetaInfo}
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
    page := c.GetInt("page")
    if page == 0 {
        page = 1
    }
    pageSize := c.GetInt("page_size")
    if pageSize == 0 {
        pageSize = 10
    }

    result, err := h.userService.List(c.Request.Context(), page, pageSize)
    if err != nil {
        h.Error(c, err)
        return
    }

    h.SuccessWithMeta(c, http.StatusOK, result.Users, &MetaInfo{
        Page:       result.Page,
        PageSize:   result.PageSize,
        Total:      result.Total,
        TotalPages: result.TotalPages,
    })
}

// Update godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body domain.UpdateUserRequest true "User update request"
// @Success 200 {object} Response{data=domain.UserResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.Error(c, errors.ErrInvalidUserID)
        return
    }

    var req domain.UpdateUserRequest
    if !h.BindAndValidate(c, &req) {
        return
    }

    user, err := h.userService.Update(c.Request.Context(), id, req)
    if err != nil {
        h.Error(c, err)
        return
    }

    h.Success(c, http.StatusOK, user)
}

// Delete godoc
// @Summary Delete user
// @Description Soft delete a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} Response
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        h.Error(c, errors.ErrInvalidUserID)
        return
    }

    if err := h.userService.Delete(c.Request.Context(), id); err != nil {
        h.Error(c, err)
        return
    }

    c.Status(http.StatusNoContent)
}
```

---

## Routing and Middleware

### Router Setup (`internal/router/router.go`)

```go
package router

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"

    "github.com/yourusername/gin-crud-api/internal/handler"
    "github.com/yourusername/gin-crud-api/internal/middleware"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// Router handles HTTP routing
type Router struct {
    engine      *gin.Engine
    userHandler *handler.UserHandler
    logger      logger.Logger
}

// NewRouter creates a new router
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

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() {
    // Health check
    r.engine.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API v1
    v1 := r.engine.Group("/api/v1")
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
```

### Middleware (`internal/middleware/logger.go`)

```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// Logger middleware for logging HTTP requests
func Logger(log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()

        if raw != "" {
            path = path + "?" + raw
        }

        log.Info("HTTP Request",
            "status", statusCode,
            "latency", latency,
            "client_ip", clientIP,
            "method", method,
            "path", path,
        )
    }
}

// Recovery middleware for recovering from panics
func Recovery(log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Error("Panic recovered",
                    "error", err,
                    "path", c.Request.URL.Path,
                )
                c.AbortWithStatusJSON(500, gin.H{
                    "success": false,
                    "error": gin.H{
                        "code":    "INTERNAL_ERROR",
                        "message": "Internal server error",
                    },
                })
            }
        }()
        c.Next()
    }
}
```

---

## Configuration Management

### Config (`internal/config/config.go`)

```go
package config

import (
    "time"

    "github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Mode string `mapstructure:"mode"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
    Host            string        `mapstructure:"host"`
    Port            int           `mapstructure:"port"`
    User            string        `mapstructure:"user"`
    Password        string        `mapstructure:"password"`
    DBName          string        `mapstructure:"dbname"`
    SSLMode         string        `mapstructure:"sslmode"`
    MaxOpenConns    int           `mapstructure:"max_open_conns"`
    MaxIdleConns    int           `mapstructure:"max_idle_conns"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// LogConfig holds logging configuration
type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
    viper.SetDefault("server.host", "localhost")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.mode", "release")

    viper.SetDefault("database.port", 5432)
    viper.SetDefault("database.sslmode", "disable")
    viper.SetDefault("database.max_open_conns", 25)
    viper.SetDefault("database.max_idle_conns", 5)
    viper.SetDefault("database.conn_max_lifetime", "5m")

    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "json")

    viper.AutomaticEnv()

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

---

## Error Handling

### Custom Errors (`pkg/errors/errors.go`)

```go
package errors

import "net/http"

// AppError represents an application error
type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    StatusCode int    `json:"-"`
}

func (e AppError) Error() string {
    return e.Message
}

// Predefined errors
var (
    ErrInternalServer = &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "Internal server error",
        StatusCode: http.StatusInternalServerError,
    }

    ErrInvalidRequest = &AppError{
        Code:       "INVALID_REQUEST",
        Message:    "Invalid request",
        StatusCode: http.StatusBadRequest,
    }

    ErrUserNotFound = &AppError{
        Code:       "USER_NOT_FOUND",
        Message:    "User not found",
        StatusCode: http.StatusNotFound,
    }

    ErrInvalidUserID = &AppError{
        Code:       "INVALID_USER_ID",
        Message:    "Invalid user ID format",
        StatusCode: http.StatusBadRequest,
    }

    ErrDuplicateEmail = &AppError{
        Code:       "DUPLICATE_EMAIL",
        Message:    "Email already exists",
        StatusCode: http.StatusConflict,
    }

    ErrDuplicateUsername = &AppError{
        Code:       "DUPLICATE_USERNAME",
        Message:    "Username already exists",
        StatusCode: http.StatusConflict,
    }
)

// GetError returns AppError from any error
func GetError(err error) *AppError {
    if appErr, ok := err.(*AppError); ok {
        return appErr
    }
    return ErrInternalServer
}
```

---

## Logging

### Logger (`pkg/logger/logger.go`)

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger interface
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Fatal(msg string, keysAndValues ...interface{})
}

// zapLogger implements Logger using zap
type zapLogger struct {
    logger *zap.SugaredLogger
}

// New creates a new logger
func New(level, format string) (Logger, error) {
    var config zap.Config

    if format == "json" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
    }

    logLevel, err := zapcore.ParseLevel(level)
    if err != nil {
        return nil, err
    }
    config.Level = zap.NewAtomicLevelAt(logLevel)

    logger, err := config.Build()
    if err != nil {
        return nil, err
    }

    return &zapLogger{logger: logger.Sugar()}, nil
}

func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
    l.logger.Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
    l.logger.Infow(msg, keysAndValues...)
}

func (l *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
    l.logger.Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
    l.logger.Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
    l.logger.Fatalw(msg, keysAndValues...)
}
```

---

## Testing

### Service Test Example (`tests/user_service_test.go`)

```go
package tests

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/yourusername/gin-crud-api/internal/domain"
    "github.com/yourusername/gin-crud-api/internal/service"
    "github.com/yourusername/gin-crud-api/pkg/logger"
)

// MockUserRepository is a mock implementation
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

// ... implement other mock methods

func TestUserService_Create(t *testing.T) {
    mockRepo := new(MockUserRepository)
    log, _ := logger.New("debug", "console")
    userService := service.NewUserService(mockRepo, log)

    ctx := context.Background()
    req := domain.CreateUserRequest{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
        FullName: "Test User",
    }

    mockRepo.On("ExistsByEmail", ctx, req.Email).Return(false, nil)
    mockRepo.On("ExistsByUsername", ctx, req.Username).Return(false, nil)
    mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

    user, err := userService.Create(ctx, req)

    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Email, user.Email)
    assert.Equal(t, req.Username, user.Username)
    mockRepo.AssertExpectations(t)
}
```

---

## Complete Code Examples

### Main Entry Point (`cmd/api/main.go`)

```go
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
    postgresRepo "github.com/yourusername/gin-crud-api/internal/repository/postgres"
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

    // Initialize logger
    log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create logger: %v\\n", err)
        os.Exit(1)
    }

    // Connect to database
    db, err := database.NewPostgresConnection(database.Config{
        Host:            cfg.Database.Host,
        Port:            cfg.Database.Port,
        User:            cfg.Database.User,
        Password:        cfg.Database.Password,
        DBName:          cfg.Database.DBName,
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
    userRepo := postgresRepo.NewUserRepository(db)

    // Initialize services
    userService := service.NewUserService(userRepo, log)

    // Initialize handlers
    userHandler := handler.NewUserHandler(userService, log)

    // Initialize router
    r := router.NewRouter(userHandler, log)
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
```

### Database Migration (`migrations/001_create_users_table.sql`)

```sql
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## Best Practices Summary

### 1. **Project Structure**
   - Use Clean Architecture principles
   - Separate concerns: handlers, services, repositories
   - Keep domain models independent

### 2. **Dependency Injection**
   - Inject dependencies through constructors
   - Use interfaces for testability
   - Avoid global state

### 3. **Error Handling**
   - Use custom error types
   - Wrap errors with context
   - Return appropriate HTTP status codes

### 4. **Validation**
   - Validate at handler layer
   - Use struct tags for validation
   - Sanitize user input

### 5. **Logging**
   - Use structured logging
   - Include context in logs
   - Different log levels for different environments

### 6. **Security**
   - Hash passwords with bcrypt
   - Use prepared statements (GORM handles this)
   - Implement rate limiting
   - Validate all inputs

### 7. **Database**
   - Use connection pooling
   - Implement migrations
   - Use transactions for complex operations
   - Soft delete instead of hard delete

### 8. **Testing**
   - Write unit tests for services
   - Use mocks for dependencies
   - Test error cases
   - Aim for high coverage

---

## Running the Application

```bash
# Install dependencies
go mod download

# Run migrations
migrate -path migrations -database "postgresql://user:password@localhost:5432/dbname?sslmode=disable" up

# Run the application
go run cmd/api/main.go

# Or with environment variables
DB_PASSWORD=secret go run cmd/api/main.go

# Run tests
go test -v ./...

# Build binary
go build -o api-server cmd/api/main.go
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| POST | /api/v1/users | Create user |
| GET | /api/v1/users | List users |
| GET | /api/v1/users/:id | Get user by ID |
| PUT | /api/v1/users/:id | Update user |
| DELETE | /api/v1/users/:id | Delete user |

---

## Conclusion

This guide demonstrates building a production-ready CRUD API in Go using:
- **Gin** for HTTP routing
- **GORM** with PostgreSQL
- **Clean Architecture** principles
- **Dependency Injection**
- **Comprehensive Error Handling**
- **Structured Logging**

Follow these patterns and adapt them to your specific requirements!
