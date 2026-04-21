package main

import (
	"context"
	"log"
	"net"
	"os"

	transpb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/trans/clients"
	"github.com/SlayerSv/payments/internal/trans/grpcserver"
	repository "github.com/SlayerSv/payments/internal/trans/repository/postgres"
	"github.com/SlayerSv/payments/internal/trans/service"
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
	log.Println("Успешное подключение к PostgreSQL!")
	db := repository.NewTransaction(dbpool)
	userClient, err := clients.NewUserClient("localhost:50051", "trans")
	walletClient, err := clients.NewWalletClient("localhost:50053", "trans")
	service := service.NewTransaction(db, userClient, walletClient)
	transserv := grpcserver.NewTrans(service)
	lis, err := net.Listen("tcp", "localhost:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptors.ServerInterceptor("gateway", nil)))
	transpb.RegisterTransServiceServer(srv, transserv)
	srv.Serve(lis)
}
