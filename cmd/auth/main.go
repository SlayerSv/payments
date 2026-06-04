package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/auth/grpcserver"
	"github.com/SlayerSv/payments/internal/auth/repo"
	"github.com/SlayerSv/payments/internal/auth/repo/postgres"
	"github.com/SlayerSv/payments/internal/auth/service"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/shared/logger"
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

	logger, cleanup := logger.NewVictoriaLogger("gateway")
	defer cleanup()
	slog.SetDefault(logger)

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
		slog.Info("Starting dabase", slog.String("number of try", strconv.Itoa(i+1)), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		slog.Error("Starting database failed", slog.String("error", err.Error()))
		return
	}
	defer dbpool.Close()
	slog.Info("Successful connection to PostgreSQL")
	repo.StartMigrations(connStr)

	tp, err := tracing.InitTracer("authorization")
	if err != nil {
		slog.Error("Init tracing", slog.String("error", err.Error()))
		return
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	client, err := bao.NewBaoClient()
	if err != nil {
		slog.Error("Connecting to secret manager", slog.String("error", err.Error()))
		return
	}
	publicKey, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		slog.Error("Retreiving public key", slog.String("error", err.Error()))
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
