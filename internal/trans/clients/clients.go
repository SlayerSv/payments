package clients

import (
	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUserClient(userAddr, serviceToken string) (pb.UserServiceClient, error) {
	userInterceptor := interceptors.NewClientInterceptor(serviceToken)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(userInterceptor),
	}
	userConn, err := grpc.NewClient(userAddr, dialOpts...)
	if err != nil {
		return nil, err
	}
	return pb.NewUserServiceClient(userConn), nil
}
