package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/gin-crud-api/internal/domain"
	"github.com/yourusername/gin-crud-api/internal/repository"
	"github.com/yourusername/gin-crud-api/pkg/errors"
	"github.com/yourusername/gin-crud-api/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repository.UserRepository
	logger   logger.Logger
}

// Create implements [UserService].
func (u *userService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Check if email existing
	exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)

	if err != nil {
		u.logger.Error("Failed to check email with provided email address", "error", err)
		return nil, errors.ErrInternalServer
	}

	if exists {
		return nil, errors.ErrDuplicateEmail
	}

	// Check if username exists
	exists, err = u.userRepo.ExistsByUsername(ctx, req.Username)

	if err != nil {
		u.logger.Error("Failed to check user name existence", "error", err)
		return nil, errors.ErrInternalServer
	}

	if exists {
		return nil, errors.ErrDuplicateUsername
	}

	// Hashpassword
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		u.logger.Error("failed to hash password", "error", err)
		return nil, errors.ErrInternalServer
	}

	// Create user entity
	user := &domain.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logger.Error("Failed to create user", "error", user)
		return nil, errors.ErrInternalServer
	}

	response := user.ToResponse()
	return &response, nil
}

// Delete implements [UserService].
func (u *userService) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("failed to get user", "error", err)
		return errors.ErrInternalServer
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	if err := u.userRepo.Delete(ctx, id); err != nil {
		u.logger.Error("failed to delete user", "error", user)
		return errors.ErrInternalServer
	}

	return nil
}

// GetByID implements [UserService].
func (u *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("failed to get user", "error", err)
		return nil, errors.ErrInternalServer
	}
	response := user.ToResponse()
	return &response, nil
}

// List implements [UserService].
func (u *userService) List(ctx context.Context, page int, pageSize int) (*UserListResponse, error) {
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

	users, total, err := u.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		u.logger.Error("failed to list users", "error", err)
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

// Update implements [UserService].
func (u *userService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Check user existence
	user, err := u.userRepo.GetByID(ctx, id)

	if err != nil {
		u.logger.Error("Failed to get the user", "error", err)
		return nil, errors.ErrInternalServer
	}

	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// check email uniqueses if being updated
	if req.Email != "" && req.Email != user.Email {
		exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)

		if err != nil {
			u.logger.Error("failed to check email existence", "error", err)
			return nil, errors.ErrInternalServer
		}

		if exists {
			return nil, errors.ErrDuplicateEmail
		}
		user.Email = req.Email
	}
	// Check username uniqueness if being updated
	if req.Username != "" && req.Username != user.Username {
		exists, err := u.userRepo.ExistsByUsername(ctx, req.Username)
		if err != nil {
			u.logger.Error("failed to check username existence", "error", err)
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

	if err := u.userRepo.Update(ctx, user); err != nil {
		u.logger.Error("failed to update user", "error", err)
		return nil, errors.ErrInternalServer
	}

	response := user.ToResponse()
	return &response, nil
}

// Define a NewUserService for injection
func NewUserService(userRepo repository.UserRepository, logger logger.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}
