package grpc

import (
	"context"

	"github.com/google/uuid"
	userv1 "github.com/yourusername/gin-crud-api/gen/go/user/v1"
	"github.com/yourusername/gin-crud-api/internal/user/domain"
	"github.com/yourusername/gin-crud-api/internal/user/service"
	apperrors "github.com/yourusername/gin-crud-api/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	userService service.UserService
}

func NewServer(userService service.UserService) *Server {
	return &Server{userService: userService}
}

func (s *Server) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	user, err := s.userService.GetByID(ctx, id)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &userv1.GetUserResponse{User: toProtoUser(user)}, nil
}

func (s *Server) ValidateCredentials(ctx context.Context, req *userv1.ValidateCredentialsRequest) (*userv1.ValidateCredentialsResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.userService.ValidateCredentials(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, toStatusError(err)
	}

	return &userv1.ValidateCredentialsResponse{User: toProtoUser(user)}, nil
}

func toProtoUser(user *domain.UserResponse) *userv1.User {
	return &userv1.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName,
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func toStatusError(err error) error {
	appErr := apperrors.GetError(err)

	switch appErr.Code {
	case apperrors.ErrInvalidCredentials.Code:
		return status.Error(codes.Unauthenticated, appErr.Message)
	case apperrors.ErrUserNotFound.Code:
		return status.Error(codes.NotFound, appErr.Message)
	case apperrors.ErrInvalidRequest.Code, apperrors.ErrInvalidUserId.Code:
		return status.Error(codes.InvalidArgument, appErr.Message)
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
