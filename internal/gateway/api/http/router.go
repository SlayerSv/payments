package app

import (
	"net/http"

	_ "github.com/SlayerSv/payments/docs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (app *App) NewRouter() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	router.Handle("GET /metrics", promhttp.Handler())

	router.HandleFunc("POST /login", app.Login)
	router.HandleFunc("POST /register", app.Register)
	router.HandleFunc("POST /restore", app.Restore)

	router.HandleFunc("PATCH /me", app.Auth(app.UpdateUser))
	router.HandleFunc("GET /me", app.Auth(app.GetUser))

	router.HandleFunc("POST /me/wallets", app.Auth(app.CreateWallet))
	router.HandleFunc("GET /me/wallets/{wallet_id}", app.Auth(app.GetWallet))
	router.HandleFunc("GET /me/wallets", app.Auth(app.GetAllWallets))
	router.HandleFunc("DELETE /me/wallets/{wallet_id}", app.Auth(app.DeleteWallet))

	router.HandleFunc("/me/wallets/{wallet_id}/transactions", app.Auth(app.GetTransactionHistory))

	router.HandleFunc("POST /me/wallets/{wallet_id}/deposit", app.Auth(app.Deposit))
	router.HandleFunc("POST /me/wallets/{wallet_id}/withdraw", app.Auth(app.Withdraw))
	router.HandleFunc("POST /me/wallets/{wallet_id}/transfer", app.Auth(app.Transfer))
	TraceRouter := otelhttp.NewHandler(router, "HTTP_Incoming_Request")
	return HTTPMetricsMiddleware(TraceRouter)
}
