package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	app "github.com/SlayerSv/payments/internal/front/http"
	"github.com/SlayerSv/payments/internal/shared/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := &app.App{
		BackendURL: "http://" + os.Getenv("BACKEND_ADDR"),
		CookieName: "frontend_token",
		Client:     &http.Client{Timeout: 10 * time.Second},
	}

	logger, cleanup := logger.NewVictoriaLogger("frontend")
	defer cleanup()
	slog.SetDefault(logger)

	router := app.NewRouter()
	addr := ":3000"
	slog.Info("Starting server", slog.String("address", addr), slog.String("url", app.BackendURL))
	err := http.ListenAndServe(addr, router)
	if err != nil {
		slog.Error("Starting server", slog.String("error", err.Error()))
	}
}
