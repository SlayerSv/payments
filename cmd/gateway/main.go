package main

import (
	"log"
	"net/http"
	"os"

	app "github.com/SlayerSv/payments/internal/gateway/api/http"
	"github.com/SlayerSv/payments/internal/gateway/clients"
	"github.com/SlayerSv/payments/internal/shared/bao"
	"github.com/SlayerSv/payments/internal/shared/jwttoken"
	"github.com/SlayerSv/payments/internal/shared/logger"
	"github.com/SlayerSv/payments/internal/shared/validator"
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
	logger, err := logger.NewConsoleLogger()
	if err != nil {
		log.Fatalf("Error getting logger: %v", err)
	}
	server := &http.Server{
		Addr:     ":8081",
		ErrorLog: logger.Error,
	}
	client, err := bao.NewBaoClient()
	if err != nil {
		log.Fatalf("Не удалось подлючиться к опенбао: %v\n", err)
	}
	key, err := jwttoken.GetPublicKey(client, "jwt_key")
	if err != nil {
		log.Fatalf("Error getting public key %v", err)
	}
	clients, err := clients.InitClients(os.Getenv("AUTH_ADDR"), os.Getenv("USER_ADDR"), os.Getenv("WALLET_ADDR"), os.Getenv("TRANS_ADDR"), "gateway")
	if err != nil {
		log.Fatalf("Error creating clients %v", err)
	}
	validate := validator.NewValidator()
	a := app.NewApp(logger, server, key, clients, validate)

	a.Server.Handler = a.NewRouter()
	a.Log.Infof("Starting server on %s", a.Server.Addr)
	err = a.Server.ListenAndServe()
	if err != nil {
		a.Log.Errorf("Error starting server: %v", err)
		os.Exit(1)
	}
}
