package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	app "github.com/SlayerSv/payments/internal/front/http"
)

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Transaction struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	CreatedAt   string  `json:"created_at"`
	AccountID   string  `json:"account_id"`
	ToAccount   string  `json:"to_account_id"`
	FromAccount string  `json:"from_account_id"`
}

func main() {
	app := &app.App{
		BackendURL: strings.TrimRight(env("BACKEND_URL", "http://localhost:8081"), "/"),
		CookieName: env("COOKIE_NAME", "frontend_token"),
		Client:     &http.Client{Timeout: 15 * time.Second},
	}

	router := app.NewRouter()
	addr := env("ADDR", "localhost:3000")
	log.Printf("frontend listening on %s, backend %s", addr, app.BackendURL)
	log.Fatal(http.ListenAndServe(addr, router))
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
