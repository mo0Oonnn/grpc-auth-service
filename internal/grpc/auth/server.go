package authgrpc

import (
	"context"
	"errors"

	ssov1 "github.com/mo0Oonnn/github.com-mo0Oonnn-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mo0Oonnn/sso/internal/grpc/validation"
	"github.com/mo0Oonnn/sso/internal/services/auth"
	"github.com/mo0Oonnn/sso/internal/storage"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(
		ctx context.Context,
		userID int64,
	) (isAdmin bool, err error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

// Register registers gRPC server.
func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

// Login checks the user credentials and if that is valid returns a JWT token.
// If user credentials are invalid, returns an error.
func (s *ServerAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

// validateLogin validates login request.
func validateLogin(req *ssov1.LoginRequest) error {
	if !validation.IsEmail(req.GetEmail()) {
		return status.Errorf(codes.InvalidArgument, "invalid email")
	} else if !validation.IsValidPassword(req.GetPassword()) {
		return status.Errorf(codes.InvalidArgument, "invalid password")
	} else if !validation.IsValidAppID(req.GetAppId()) {
		return status.Errorf(codes.InvalidArgument, "invalid app id")
	}
	return nil
}

// Register creates a new user.
func (s *ServerAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

// validateRegister validates register request.
func validateRegister(req *ssov1.RegisterRequest) error {
	if !validation.IsEmail(req.GetEmail()) {
		return status.Errorf(codes.InvalidArgument, "invalid email")
	} else if !validation.IsValidPassword(req.GetPassword()) {
		return status.Errorf(codes.InvalidArgument, "invalid password")
	}
	return nil
}

// IsAdmin checks if the user is an admin.
//
// If the user doesn't exist, returns an error.
func (s *ServerAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

// validateIsAdmin validates IsAdmin request.
func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if !validation.IsValidUserID(req.GetUserId()) {
		return status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	return nil
}
