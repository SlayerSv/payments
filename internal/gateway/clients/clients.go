package clients

import (
	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth pb.AuthServiceClient
	User pb.UserServiceClient
}

func InitClients(authAddr, userAddr, serviceToken string) (*Clients, error) {
	// Создаем интерцептор
	authInterceptor := interceptors.NewClientInterceptor(serviceToken)

	// Настройки подключения (интерцептор вешаем прямо сюда)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
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

	return &Clients{
		Auth: pb.NewAuthServiceClient(authConn),
		User: pb.NewUserServiceClient(userConn),
	}, nil
}
