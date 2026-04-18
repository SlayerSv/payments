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
	router.HandleFunc("PATCH /users", app.Auth(app.UpdateUser))
	router.HandleFunc("POST /users/restore", app.SendOTP)

	return router
}
