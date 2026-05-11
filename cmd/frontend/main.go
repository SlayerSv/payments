package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	app "github.com/SlayerSv/payments/internal/front/http"
)

func main() {
	app := &app.App{
		BackendURL: "http://" + os.Getenv("BACKEND_ADDR"),
		CookieName: env("COOKIE_NAME", "frontend_token"),
		Client:     &http.Client{Timeout: 15 * time.Second},
	}

	router := app.NewRouter()
	addr := ":3000"
	log.Printf("frontend listening on %s, backend %s", addr, app.BackendURL)
	log.Fatal(http.ListenAndServe(addr, router))
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
