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
	"github.com/SlayerSv/payments/internal/shared/metrics"
	"github.com/SlayerSv/payments/internal/shared/tracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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

	tp, err := tracing.InitTracer("authorization")
	if err != nil {
		log.Fatalf("Error init tracing: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

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
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
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

	grpc_prometheus.Register(srv)
	metrics.InitMetricsServer(os.Getenv("AUTH_METRICS_PORT"))

	srv.Serve(lis)
}
