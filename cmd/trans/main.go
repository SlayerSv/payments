package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	transpb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/trans/clients"
	"github.com/SlayerSv/payments/internal/trans/grpcserver"
	"github.com/SlayerSv/payments/internal/trans/repository"
	"github.com/SlayerSv/payments/internal/trans/repository/postgres"
	"github.com/SlayerSv/payments/internal/trans/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()

	connStr := os.Getenv("TRANS_DB_CONN")
	var dbpool *pgxpool.Pool
	var err error
	for i := 0; i < 5; i++ {
		dbpool, err = pgxpool.New(context.Background(), connStr)
		if err == nil {
			err = dbpool.Ping(context.Background())
			if err == nil {
				break
			}
		}
		log.Printf("База еще не готова (попытка %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе после 5 попыток: %v", err)
	}
	defer dbpool.Close()
	log.Println("Успешное подключение к PostgreSQL!")
	repository.StartMigrations(connStr)
	db := postgres.NewTransaction(dbpool)

	userClient, err := clients.NewUserClient(os.Getenv("USER_ADDR"), "trans")
	walletClient, err := clients.NewWalletClient(os.Getenv("WALLET_ADDR"), "trans")
	service := service.NewTransaction(db, userClient, walletClient)
	transserv := grpcserver.NewTrans(service)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	client, err := bao.NewBaoClient()
	if err != nil {
		log.Fatalf("Не удалось подлючиться к опенбао: %v\n", err)
	}
	publicKey, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		log.Fatalf("Не удалось достать публичный ключ: %v\n", err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptors.ServerInterceptor([]string{"gateway"}, publicKey)))
	transpb.RegisterTransServiceServer(srv, transserv)

	srv.Serve(lis)
}
