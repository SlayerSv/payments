package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/SlayerSv/payments/internal/gateway/adapter"
	app "github.com/SlayerSv/payments/internal/gateway/api/http"
	"github.com/SlayerSv/payments/internal/gateway/clients"
	"github.com/SlayerSv/payments/internal/gateway/service"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/shared/logger"
	"github.com/SlayerSv/payments/internal/shared/tracing"
	"github.com/SlayerSv/payments/internal/shared/validator"
	"github.com/joho/godotenv"
)

// @title           Payments API
// @version         1.0
// @description     API Server for Payments system
// @termsOfService  http://swagger.io/terms/
// @openapi 3.0.0
// @host            localhost:8081
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	godotenv.Load()
	logger, cleanup := logger.NewVictoriaLogger("gateway")
	defer cleanup()

	// Теперь используем стандартный slog
	slog.SetDefault(logger)

	tp, err := tracing.InitTracer("gateway")
	if err != nil {
		logger.Error("Initializing tracing", slog.String("error", err.Error()))
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	server := &http.Server{
		Addr: ":8081",
	}
	client, err := bao.NewBaoClient()
	if err != nil {
		logger.Error("Connecting to open bao", slog.String("error", err.Error()))
	}
	key, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		logger.Error("Getting public key", slog.String("error", err.Error()))
	}
	clients, err := clients.InitClients(os.Getenv("AUTH_ADDR"), os.Getenv("USER_ADDR"), os.Getenv("WALLET_ADDR"), os.Getenv("TRANS_ADDR"), "gateway")
	if err != nil {
		logger.Error("Creating clients", slog.String("error", err.Error()))
	}
	validate := validator.NewValidator()
	authAdapter := adapter.NewAuth(clients.Auth)
	authService := service.NewAuth(authAdapter)
	userAdapter := adapter.NewUser(clients.User)
	userService := service.NewUser(userAdapter)
	transAdapter := adapter.NewTrans(clients.Trans)
	transService := service.NewTrans(transAdapter)
	walletAdapter := adapter.NewWallet(clients.Wallet)
	walletService := service.NewWallet(walletAdapter)
	a := app.NewApp(logger, server, key, authService, userService, transService, walletService, validate)

	a.Server.Handler = a.NewRouter()
	logger.Info("Starting server", slog.String("address", a.Server.Addr))
	err = a.Server.ListenAndServe()
	if err != nil {
		logger.Error("Starting server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
