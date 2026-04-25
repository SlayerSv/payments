package clients

import (
	pb "github.com/SlayerSv/payments/gen/auth"
	transpb "github.com/SlayerSv/payments/gen/trans"
	walletpb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth   pb.AuthServiceClient
	User   pb.UserServiceClient
	Wallet walletpb.WalletServiceClient
	Trans  transpb.TransServiceClient
}

func InitClients(authAddr, userAddr, walletAddr, transAddr, serviceToken string) (*Clients, error) {
	// Создаем интерцептор
	interceptor := interceptors.NewClientInterceptor(serviceToken)

	// Настройки подключения (интерцептор вешаем прямо сюда)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor),
	}

	// Коннектимся к Auth
	authConn, err := grpc.NewClient(authAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	// Коннектимся к User
	userConn, err := grpc.NewClient(userAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	// Коннектимся к Wallet
	walletConn, err := grpc.NewClient(walletAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	// Коннектимся к Trans
	transConn, err := grpc.NewClient(transAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Auth:   pb.NewAuthServiceClient(authConn),
		Trans:  transpb.NewTransServiceClient(transConn),
		User:   pb.NewUserServiceClient(userConn),
		Wallet: walletpb.NewWalletServiceClient(walletConn),
	}, nil
}
