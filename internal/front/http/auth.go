package app

import (
	"encoding/json"
	"net/http"

	"github.com/SlayerSv/payments/internal/auth/models"
)

func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, "login", PageData{Title: "Вход"})
}
func (a *App) registerPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, "register", PageData{Title: "Регистрация"})
}
func (a *App) restorePage(w http.ResponseWriter, r *http.Request) {
	a.render(w, "restore", PageData{Title: "Восстановление"})
}

func (a *App) loginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	b, _, err := a.api(r.Context(), http.MethodPost, "/login", "", models.LoginRequest{
		Email:    email,
		Password: password,
	}, nil)
	if err != nil {
		a.render(w, "login", PageData{Title: "Вход", Error: err.Error()})
		return
	}
	resp := models.LoginResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		a.render(w, "login", PageData{Title: "Вход", Error: err.Error()})
		return
	}
	if resp.Token != "" {
		a.setToken(w, resp.Token)
		http.Redirect(w, r, "/me", http.StatusFound)
		return
	}
	a.render(w, "login", PageData{Title: "Вход", Error: "Не удалось найти токен в ответе API"})
}

func (a *App) registerPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	_, _, err := a.api(r.Context(), http.MethodPost, "/register", "", models.RegisterRequest{
		Email: email,
	}, nil)
	if err != nil {
		a.render(w, "register", PageData{Title: "Регистрация", Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *App) restorePost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	_, _, err := a.api(r.Context(), http.MethodPost, "/restore", "", map[string]string{
		"email": email,
	}, nil)
	if err != nil {
		a.render(w, "restore", PageData{Title: "Восстановление", Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}
