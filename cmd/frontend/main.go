package main

import (
	"log"
	"net/http"
	"os"
	"time"

	app "github.com/SlayerSv/payments/internal/front/http"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := &app.App{
		BackendURL: "http://" + os.Getenv("BACKEND_ADDR"),
		CookieName: "frontend_token",
		Client:     &http.Client{Timeout: 10 * time.Second},
	}

	router := app.NewRouter()
	addr := ":3000"
	log.Printf("frontend listening on %s, backend %s", addr, app.BackendURL)
	log.Fatal(http.ListenAndServe(addr, router))
}
