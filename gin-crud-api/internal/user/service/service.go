package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/gin-crud-api/internal/user/domain"
)

type UserService interface {
	Create(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
	ValidateCredentials(ctx context.Context, email, password string) (*domain.UserResponse, error)
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
