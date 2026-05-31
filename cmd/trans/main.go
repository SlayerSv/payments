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
	"github.com/SlayerSv/payments/internal/shared/metrics"
	"github.com/SlayerSv/payments/internal/shared/tracing"
	"github.com/SlayerSv/payments/internal/trans/clients"
	"github.com/SlayerSv/payments/internal/trans/grpcserver"
	"github.com/SlayerSv/payments/internal/trans/repository"
	"github.com/SlayerSv/payments/internal/trans/repository/postgres"
	"github.com/SlayerSv/payments/internal/trans/service"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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

	tp, err := tracing.InitTracer("transactions")
	if err != nil {
		log.Fatalf("Error init tracing: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			// 1. Метрики Прометея
			// Они должны быть в самом начале цепочки, чтобы замерить полное время
			// выполнения запроса, включая то время, что тратится на проверку JWT.
			grpc_prometheus.UnaryServerInterceptor,
			interceptors.ServerInterceptor([]string{"gateway"}, publicKey),
		),
	)
	transpb.RegisterTransServiceServer(srv, transserv)
	grpc_prometheus.Register(srv)

	metrics.InitMetricsServer(os.Getenv("TRANS_METRICS_PORT"))

	srv.Serve(lis)
}
