package app

import (
	"net/http"

	_ "github.com/SlayerSv/payments/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *App) NewRouter() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	router.HandleFunc("POST /login", app.Login)
	router.HandleFunc("POST /register", app.Register)
	router.HandleFunc("POST /restore", app.SendOTP)

	router.HandleFunc("PATCH /me", app.Auth(app.UpdateUser))
	router.HandleFunc("GET /me", app.Auth(app.GetUser))

	router.HandleFunc("POST /me/accounts", app.Auth(app.CreateAccount))
	router.HandleFunc("GET /me/accounts/{account_id}", app.Auth(app.GetAccount))
	router.HandleFunc("GET /me/accounts", app.Auth(app.GetAllAccounts))
	router.HandleFunc("DELETE /me/accounts", app.Auth(app.DeleteAccount))

	// router.HandleFunc("POST /me/history", app.GetAccHistory)

	// router.HandleFunc("POST /transactions", app.Transaction)

	return router
}
