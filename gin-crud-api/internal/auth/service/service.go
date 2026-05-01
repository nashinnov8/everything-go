package service

import (
	"context"

	"github.com/yourusername/gin-crud-api/internal/auth/domain"
	apperrors "github.com/yourusername/gin-crud-api/pkg/errors"
)

type UserClient interface {
	ValidateCredentials(ctx context.Context, email, password string) (*domain.UserIdentity, error)
}

type TokenIssuer interface {
	IssueTokenPair(ctx context.Context, user domain.UserIdentity) (*domain.TokenPair, error)
}

type AuthService interface {
	Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error)
}

type authService struct {
	users  UserClient
	tokens TokenIssuer
}

func NewAuthService(users UserClient, tokens TokenIssuer) AuthService {
	return &authService{
		users:  users,
		tokens: tokens,
	}
}

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.ErrInvalidRequest
	}

	user, err := s.users.ValidateCredentials(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	tokens, err := s.tokens.IssueTokenPair(ctx, *user)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		User:   *user,
		Tokens: *tokens,
	}, nil
}
