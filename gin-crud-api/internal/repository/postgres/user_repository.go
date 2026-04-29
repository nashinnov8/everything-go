package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/gin-crud-api/internal/domain"
	"github.com/yourusername/gin-crud-api/internal/repository"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// Create implements [repository.UserRepository].
func (u *userRepository) Create(ctx context.Context, user *domain.User) error {
	return u.db.WithContext(ctx).Create(user).Error
}

// Delete implements [repository.UserRepository].
func (u *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return u.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// ExistsByEmail implements [repository.UserRepository].
func (u *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Count(&count).Error
	return (count > 0), err
}

// ExistsByUsername implements [repository.UserRepository].
func (u *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&domain.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// GetByEmail implements [repository.UserRepository].
func (u *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	if err := u.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByID implements [repository.UserRepository].
func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User

	if err := u.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername implements [repository.UserRepository].
func (u *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User

	if err := u.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// List implements [repository.UserRepository].
func (u *userRepository) List(ctx context.Context, offset int, limit int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	db := u.db.WithContext(ctx).Model(&domain.User{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update implements [repository.UserRepository].
func (u *userRepository) Update(ctx context.Context, user *domain.User) error {
	return u.db.WithContext(ctx).Save(user).Error
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}
