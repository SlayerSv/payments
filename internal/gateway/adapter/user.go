package adapter

import (
	"context"
	"fmt"

	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/models"
	"google.golang.org/protobuf/types/known/emptypb"
)

type User struct {
	userClient pb.UserServiceClient
}

func NewUser(userClient pb.UserServiceClient) *User {
	return &User{userClient: userClient}
}

func (u *User) Get(ctx context.Context) (models.UserDTO, error) {
	resp, err := u.userClient.Get(ctx, &emptypb.Empty{})
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("error calling auth client: %w", err)
	}
	return pbToUser(resp), nil
}

func (u *User) Update(ctx context.Context, name, password *string) (models.UserDTO, error) {
	resp, err := u.userClient.Update(ctx, &pb.UpdateRequest{NewName: name, NewPassword: password})
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("error calling auth client: %w", err)
	}
	return pbToUser(resp), nil
}

func pbToUser(user *pb.UserResponse) models.UserDTO {
	return models.UserDTO{
		ID:        user.Id,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}
}
