package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/shared/logger"
	"github.com/SlayerSv/payments/internal/shared/metrics"
	"github.com/SlayerSv/payments/internal/shared/tracing"
	"github.com/SlayerSv/payments/internal/wallet/grpcserver"
	"github.com/SlayerSv/payments/internal/wallet/repository"
	"github.com/SlayerSv/payments/internal/wallet/repository/postgres"
	"github.com/SlayerSv/payments/internal/wallet/service"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load()

	logger, cleanup := logger.NewVictoriaLogger("wallet")
	defer cleanup()
	slog.SetDefault(logger)

	connStr := os.Getenv("WALLET_DB_CONN")

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
		slog.Info("Starting dabase", slog.String("number of try", strconv.Itoa(i+1)), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		slog.Error("Connecting to database failed", slog.String("error", err.Error()))
		return
	}
	defer dbpool.Close()
	slog.Info("Successful connection to PostgreSQL")
	repository.StartMigrations(connStr)

	db := postgres.NewWallet(dbpool)
	service := service.NewWallet(db)
	walletserv := grpcserver.NewWallet(service)
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		slog.Error("Opening tcp connection", slog.String("error", err.Error()), slog.String("port", "50053"))
		return
	}
	client, err := bao.NewBaoClient()
	if err != nil {
		slog.Error("Connecting to secret manager", slog.String("error", err.Error()))
		return
	}
	publicKey, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		slog.Error("Retreiving public key", slog.String("error", err.Error()))
		return
	}
	tp, err := tracing.InitTracer("wallet")
	if err != nil {
		slog.Error("Initializing tracer", slog.String("error", err.Error()))
		return
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			interceptors.ServerInterceptor([]string{"trans", "gateway"}, publicKey),
		),
	)
	pb.RegisterWalletServiceServer(srv, walletserv)
	grpc_prometheus.Register(srv)

	metrics.InitMetricsServer(os.Getenv("WALLET_METRICS_PORT"))
	srv.Serve(lis)
}
