package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/wallet/grpcserver"
	repository "github.com/SlayerSv/payments/internal/wallet/repository/postgres"
	"github.com/SlayerSv/payments/internal/wallet/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()
	connStr := os.Getenv("WALLET_DB_CONN")

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
	log.Println("Успешное подключение к PostgreSQL!")
	db := repository.NewWallet(dbpool)
	service := service.NewWallet(db)
	walletserv := grpcserver.NewWallet(service)
	lis, err := net.Listen("tcp", "localhost:50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptors.ServerInterceptor("trans", nil)))
	pb.RegisterWalletServiceServer(srv, walletserv)
	srv.Serve(lis)
}
