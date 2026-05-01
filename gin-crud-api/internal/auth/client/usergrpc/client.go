package usergrpc

import (
	"context"

	userv1 "github.com/yourusername/gin-crud-api/gen/go/user/v1"
	"github.com/yourusername/gin-crud-api/internal/auth/domain"
	apperrors "github.com/yourusername/gin-crud-api/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	users userv1.UserServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{users: userv1.NewUserServiceClient(conn)}
}

func (c *Client) ValidateCredentials(ctx context.Context, email, password string) (*domain.UserIdentity, error) {
	resp, err := c.users.ValidateCredentials(ctx, &userv1.ValidateCredentialsRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, mapError(err)
	}
	if resp.GetUser() == nil {
		return nil, apperrors.ErrUserNotFound
	}

	return toIdentity(resp.GetUser()), nil
}

func (c *Client) GetUserByID(ctx context.Context, id string) (*domain.UserIdentity, error) {
	resp, err := c.users.GetUserByID(ctx, &userv1.GetUserByIDRequest{Id: id})
	if err != nil {
		return nil, mapError(err)
	}
	if resp.GetUser() == nil {
		return nil, apperrors.ErrUserNotFound
	}

	return toIdentity(resp.GetUser()), nil
}

func toIdentity(user *userv1.User) *domain.UserIdentity {
	return &domain.UserIdentity{
		ID:       user.GetId(),
		Email:    user.GetEmail(),
		Username: user.GetUsername(),
		FullName: user.GetFullName(),
		IsActive: user.GetIsActive(),
	}
}

func mapError(err error) error {
	switch status.Code(err) {
	case codes.Unauthenticated:
		return apperrors.ErrInvalidCredentials
	case codes.NotFound:
		return apperrors.ErrUserNotFound
	case codes.InvalidArgument:
		return apperrors.ErrInvalidRequest
	default:
		return apperrors.NewInternalServer(err)
	}
}
