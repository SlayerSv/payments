package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	authmodels "github.com/SlayerSv/payments/internal/auth/models"
	transmodels "github.com/SlayerSv/payments/internal/trans/models"
	walletmodels "github.com/SlayerSv/payments/internal/wallet/models"
)

type App struct {
	BackendURL string
	Client     *http.Client
	CookieName string
}

type PageData struct {
	Title        string
	Error        string
	Authed       bool
	Accounts     []walletmodels.AccountResponse
	User         authmodels.UserDTO
	Transactions []transmodels.TransactionDTO
	Account      walletmodels.AccountResponse
}

func (a *App) api(ctx context.Context, method, path, token string, in any, out any) ([]byte, int, error) {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.BackendURL+path, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(b))
		if msg == "" {
			msg = resp.Status
		}
		return b, resp.StatusCode, fmt.Errorf("%s", msg)
	}

	if out != nil && len(b) > 0 {
		if err := json.Unmarshal(b, out); err != nil {
			return b, resp.StatusCode, err
		}
	}

	return b, resp.StatusCode, nil
}

func (a *App) token(r *http.Request) string {
	c, err := r.Cookie(a.CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func (a *App) setToken(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   60 * 15,
	})
}

func (a *App) render(w http.ResponseWriter, page string, data PageData) {
	// Формируем пути к файлам шаблонов
	layoutPath := "internal/front/templates/layout.html"
	pagePath := fmt.Sprintf("internal/front/templates/%s.html", page)

	// 1. Создаем объект шаблона и регистрируем функцию formatMoney
	tmpl := template.New("layout").Funcs(template.FuncMap{
		"formatMoney": func(amount int64) string {
			return fmt.Sprintf("%.2f", float64(amount)/100.0)
		},
		"strVal": func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		},
	})

	t, err := tmpl.ParseFiles(layoutPath, pagePath)
	if err != nil {
		log.Printf("Ошибка парсинга шаблона %s: %v", pagePath, err)
		http.Error(w, "Внутренняя ошибка сервера (проблема с шаблонами)", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Рендерим шаблон "layout", в который встроится нужный "content"
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Ошибка выполнения шаблона %s: %v", pagePath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := a.token(r)
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if _, _, err := a.api(r.Context(), http.MethodGet, "/me", token, nil, nil); err != nil {
			a.setToken(w, "")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
