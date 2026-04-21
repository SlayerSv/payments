package clients

import (
	pb "github.com/SlayerSv/payments/gen/auth"
	walletpb "github.com/SlayerSv/payments/gen/wallet"
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

func NewWalletClient(walletAddr, serviceToken string) (walletpb.WalletServiceClient, error) {
	walletInterceptor := interceptors.NewClientInterceptor(serviceToken)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(walletInterceptor),
	}
	WalletConn, err := grpc.NewClient(walletAddr, dialOpts...)
	if err != nil {
		return nil, err
	}
	return walletpb.NewWalletServiceClient(WalletConn), nil
}
