package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/auth/grpcserver"
	"github.com/SlayerSv/payments/internal/auth/repo"
	"github.com/SlayerSv/payments/internal/auth/repo/postgres"
	"github.com/SlayerSv/payments/internal/auth/service"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()
	connStr := os.Getenv("AUTH_DB_CONN")

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
	repo.StartMigrations(connStr)

	client, err := bao.NewBaoClient()
	if err != nil {
		log.Fatalf("Не удалось подлючиться к опенбао: %v\n", err)
	}
	publicKey, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		log.Fatalf("Не удалось достать публичный ключ: %v\n", err)
	}
	lis, _ := net.Listen("tcp", ":50051")

	// Настраиваем сервер с ОДНИМ интерцептором
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptors.ServerInterceptor([]string{"gateway", "trans"}, publicKey),
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
