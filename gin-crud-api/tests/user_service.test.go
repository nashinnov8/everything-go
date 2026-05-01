package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/gin-crud-api/internal/user/domain"
	"github.com/yourusername/gin-crud-api/internal/user/service"
	apperrors "github.com/yourusername/gin-crud-api/pkg/errors"
	"github.com/yourusername/gin-crud-api/pkg/logger"
	"golang.org/x/crypto/bcrypt"
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

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

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

func TestUserService_ValidateCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("debug", "console")
	userService := service.NewUserService(mockRepo, log)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	assert.NoError(t, err)

	storedUser := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Username: "testuser",
		Password: string(hashedPassword),
		FullName: "Test User",
		IsActive: true,
	}

	mockRepo.On("GetByEmail", ctx, email).Return(storedUser, nil)

	user, err := userService.ValidateCredentials(ctx, email, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, storedUser.ID, user.ID)
	assert.Equal(t, storedUser.Email, user.Email)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ValidateCredentials_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log, _ := logger.New("debug", "console")
	userService := service.NewUserService(mockRepo, log)

	ctx := context.Background()
	email := "test@example.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	assert.NoError(t, err)

	storedUser := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Username: "testuser",
		Password: string(hashedPassword),
		FullName: "Test User",
		IsActive: true,
	}

	mockRepo.On("GetByEmail", ctx, email).Return(storedUser, nil)

	user, err := userService.ValidateCredentials(ctx, email, "wrong-password")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}
