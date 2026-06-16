package app

import (
	"net/http"
)

func (app *App) NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("GET /login", app.loginPage)
	mux.HandleFunc("POST /login", app.loginPost)
	mux.HandleFunc("GET /register", app.registerPage)
	mux.HandleFunc("POST /register", app.registerPost)
	mux.HandleFunc("GET /restore", app.restorePage)
	mux.HandleFunc("POST /restore", app.restorePost)
	mux.HandleFunc("GET /logout", app.logout)

	mux.HandleFunc("GET /me", app.requireAuth(app.mePage))
	mux.HandleFunc("POST /me", app.requireAuth(app.meUpdate))
	mux.HandleFunc("POST /me/wallets", app.requireAuth(app.createWallet))
	mux.HandleFunc("GET /me/wallets", app.requireAuth(app.walletsPage))
	mux.HandleFunc("GET /me/wallets/{wallet_id}", app.requireAuth(app.walletPage))
	mux.HandleFunc("GET /me/wallets/{wallet_id}/history", app.requireAuth(app.walletHistoryPage))
	mux.HandleFunc("POST /me/wallets/{wallet_id}/deposit", app.requireAuth(app.deposit))
	mux.HandleFunc("POST /me/wallets/{wallet_id}/withdraw", app.requireAuth(app.withdraw))
	mux.HandleFunc("POST /me/wallets/{wallet_id}/transfer", app.requireAuth(app.transfer))
	return mux
}
