package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/auth/grpcserver"
	"github.com/SlayerSv/payments/internal/auth/repo/postgres"
	"github.com/SlayerSv/payments/internal/auth/service"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/hashicorp/vault/api"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()
	connStr := os.Getenv("AUTH_DB_CONN")

	// 1. Создаем пул соединений
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось создать пул соединений: %v\n", err)
	}
	defer dbpool.Close()

	// 2. Проверяем соединение
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Ошибка пинга пула: %v\n", err)
	}
	fmt.Println("Успешное подключение к PostgreSQL!")

	config := api.DefaultConfig()
	config.Address = "http://localhost:8200" // Адрес OpenBao
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Не удалось создать клиент опенбао: %v\n", err)
	}
	client.SetToken("myroot")
	publicKey, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		log.Fatalf("Не удалось достать публичный ключ: %v\n", err)
	}
	lis, _ := net.Listen("tcp", "localhost:50051")

	// Настраиваем сервер с ОДНИМ интерцептором
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptors.ServerInterceptor("gateway", publicKey),
		),
	)
	authRepo := postgres.NewAuth(dbpool)
	userRepo := postgres.NewUser(dbpool)
	otpRepo := postgres.NewOTP(dbpool)
	authService := service.NewAuth(authRepo, otpRepo, userRepo, client)
	userService := service.NewUser(userRepo, publicKey)
	authServer := grpcserver.NewAuthServer(authService, userService)
	pb.RegisterAuthServiceServer(srv, authServer)
	pb.RegisterUserServiceServer(srv, authServer)
	srv.Serve(lis)
}
