package main

import (
	"context"
    "crypto"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	transpb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/shared/kafka"
	"github.com/SlayerSv/payments/internal/shared/logger"
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

	logger, cleanup := logger.NewVictoriaLogger("transactions")
	defer cleanup()
	slog.SetDefault(logger)

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
	db := postgres.NewTransaction(dbpool)

	userClient, err := clients.NewUserClient(os.Getenv("USER_ADDR"), "trans")
	walletClient, err := clients.NewWalletClient(os.Getenv("WALLET_ADDR"), "trans")
	service := service.NewTransaction(db, userClient, walletClient)
	transserv := grpcserver.NewTrans(service)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		slog.Error("Opening tcp connection", slog.String("error", err.Error()), slog.String("port", "50052"))
		return
	}

	client, err := bao.NewBaoClient()
	if err != nil {
		slog.Error("Connecting to secret manager", slog.String("error", err.Error()))
		return
	}
    var publicKey crypto.PublicKey
	for i := 0; i < 5; i++ {
		publicKey, err = jwttoken.GetPublicKey(client, "jwt_key")
		if err == nil {
			slog.Info("Successfully retrieved public key from OpenBao")
			break
		}
		
		slog.Warn("Public key not found yet, retrying...", slog.Int("attempt", i+1), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		slog.Error("Failed to retrieve public key after retries", slog.String("error", err.Error()))
		return 
	}

	tp, err := tracing.InitTracer("transactions")
	if err != nil {
		slog.Error("Initializing tracer", slog.String("error", err.Error()))
		return
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

	kafkaBrokersEnv := os.Getenv("KAFKA_BROKERS")
	brokers := strings.Split(kafkaBrokersEnv, ",")

	// Инициализируем консюмера.
	// Указываем топик "wallet-transactions" и группу "audit-group"
	consumer := kafka.NewConsumer(brokers, "wallet-transactions", "audit-group")
	defer consumer.Close()

	ctx := context.Background()

	// Запускаем бесконечное чтение из Кафки
	consumer.StartConsume(ctx, func(ctx context.Context, msg []byte) error {
		// Эта функция будет срабатывать на КАЖДОЕ новое сообщение в Кафке!
		slog.Info("[AUDIT LOG FROM KAFKA]", slog.String("message", string(msg)))
		return nil
	})

	srv.Serve(lis)
}
