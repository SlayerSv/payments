package grpcserver

import (
	"context"

	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthProvider interface {
	Register(ctx context.Context, email string) error
	Login(ctx context.Context, email, password string) (string, error)
	Restore(ctx context.Context, email string) error
}

type UserProvider interface {
	UpdateUser(ctx context.Context, id uuid.UUID, name *string, pass *string) (models.User, error)
}

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedUserServiceServer
	auth AuthProvider
	user UserProvider
}

func NewAuthServer(authLogic AuthProvider, userLogic UserProvider) *AuthServer {
	return &AuthServer{
		auth: authLogic,
		user: userLogic,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := s.auth.Register(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{Status: "ok"}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (s *AuthServer) Restore(ctx context.Context, req *pb.RestoreRequest) (*pb.RestoreResponse, error) {
	err := s.auth.Restore(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &pb.RestoreResponse{Status: "ok"}, nil
}

func (a *AuthServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	uid, _ := uuid.Parse(ctx.Value("user_id").(string))
	user, err := a.user.UpdateUser(ctx, uid, req.NewName, req.NewPassword)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:        user.ID.String(),
		Email:     user.Email,
		Name:      *user.Name,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
