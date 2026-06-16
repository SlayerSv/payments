package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SlayerSv/payments/internal/shared/models"
)

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if token := a.token(r); token != "" {
		if _, _, err := a.api(r.Context(), http.MethodGet, "/me", token, nil, nil); err == nil {
			http.Redirect(w, r, "/me", http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *App) mePage(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	var me models.UserDTO
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me", token, nil, &me); err != nil {
		a.render(w, "me", PageData{Title: "Личный кабинет", Error: err.Error()})
		return
	}

	var wallets models.WalletsDTO
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/wallets", token, nil, &wallets)
	a.render(w, "me", PageData{
		Title:   "Личный кабинет",
		Authed:  true,
		User:    me,
		Wallets: wallets.Wallets,
	})
}

func (a *App) meUpdate(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	req := models.UpdateUserRequest{}
	if v := strings.TrimSpace(r.FormValue("name")); v != "" {
		req.Name = &v
	}
	if v := strings.TrimSpace(r.FormValue("password")); v != "" {
		req.Password = &v
	}
	if req.Name == nil && req.Password == nil {
		http.Redirect(w, r, "/me", http.StatusFound)
		return
	}

	if _, _, err := a.api(r.Context(), http.MethodPatch, "/me", token, req, nil); err != nil {
		a.render(w, "me", PageData{Title: "Личный кабинет", Authed: true, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/me", http.StatusFound)
}

func (a *App) walletsPage(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	var wallets models.WalletsDTO
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me/wallets", token, nil, &wallets); err != nil {
		a.render(w, "wallets", PageData{Title: "Счета", Authed: true, Error: err.Error()})
		return
	}

	a.render(w, "wallets", PageData{Title: "Счета", Authed: true, Wallets: wallets.Wallets})
}

func (a *App) createWallet(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/wallets", token, map[string]any{}, nil); err != nil {
		a.render(w, "wallets", PageData{Title: "Счета", Authed: true, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/me/wallets", http.StatusFound)
}

func (a *App) walletPage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("wallet_id")
	token := a.token(r)

	var wallet models.WalletDTO
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me/wallets/"+id, token, nil, &wallet); err != nil {
		a.render(w, "wallet", PageData{Title: "Счет", Authed: true, Error: err.Error(), Wallet: wallet})
		return
	}

	var tx models.TransactionHistory
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/wallets/"+id+"/transactions", token, nil, &tx)
	recentTx := tx.Transactions
	if len(recentTx) > 5 {
		recentTx = recentTx[:5]
	}
	a.render(w, "wallet", PageData{
		Title:        "Счет",
		Authed:       true,
		Wallet:       wallet,
		Transactions: recentTx,
	})
}

func (a *App) walletHistoryPage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("wallet_id")
	token := a.token(r)

	// Запрашиваем инфу о кошельке
	var wallet models.WalletDTO
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me/wallets/"+id, token, nil, &wallet); err != nil {
		a.render(w, "transactions", PageData{Title: "История", Authed: true, Error: err.Error()})
		return
	}

	// Получаем ВСЕ транзакции
	var tx models.TransactionHistory
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/wallets/"+id+"/transactions", token, nil, &tx)

	allTx := tx.Transactions
	totalItems := len(allTx)
	pageSize := 10

	// Читаем параметр ?page=... из URL
	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	// Считаем общее количество страниц
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// Вырезаем нужный кусок (slice) для текущей страницы
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	pageTx := []models.TransactionDTO{}
	if start < totalItems {
		pageTx = allTx[start:end]
	}

	// Отдаем в новый шаблон
	a.render(w, "transactions", PageData{
		Title:        "Вся история",
		Authed:       true,
		Wallet:       wallet,
		Transactions: pageTx,
		CurrentPage:  page,
		TotalPages:   totalPages,
		HasNext:      page < totalPages,
		HasPrev:      page > 1,
		NextPage:     page + 1,
		PrevPage:     page - 1,
	})
}

func (a *App) deposit(w http.ResponseWriter, r *http.Request)  { a.walletAction(w, r, "/deposit") }
func (a *App) withdraw(w http.ResponseWriter, r *http.Request) { a.walletAction(w, r, "/withdraw") }

func (a *App) transfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("wallet_id")
	token := a.token(r)

	amount, err := parseAmount(r.FormValue("amount"))
	if err != nil {
		a.walletError(w, r, id, err)
		return
	}
	payload := models.TransferRequest{
		Amount:           amount,
		ReceiverWalletID: strings.TrimSpace(r.FormValue("to_wallet_id")),
	}
	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/wallets/"+id+"/transfer", token, payload, nil); err != nil {
		a.walletError(w, r, id, err)
		return
	}
	http.Redirect(w, r, "/me/wallets/"+id, http.StatusFound)
}

func (a *App) walletAction(w http.ResponseWriter, r *http.Request, suffix string) {
	id := r.PathValue("wallet_id")
	token := a.token(r)

	amount, err := parseAmount(r.FormValue("amount"))
	if err != nil {
		a.walletError(w, r, id, err)
		return
	}

	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/wallets/"+id+suffix, token, models.DepositRequest{
		Amount: amount,
	}, nil); err != nil {
		a.walletError(w, r, id, err)
		return
	}
	http.Redirect(w, r, "/me/wallets/"+id, http.StatusFound)
}

func (a *App) walletError(w http.ResponseWriter, r *http.Request, id string, err error) {
	token := a.token(r)

	var wallet models.WalletDTO
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/wallets/"+id, token, nil, &wallet)

	var tx []models.TransactionDTO
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/wallets/"+id+"/transactions", token, nil, &tx)

	a.render(w, "wallet", PageData{
		Title:        "Счет",
		Authed:       true,
		Error:        err.Error(),
		Wallet:       wallet,
		Transactions: tx,
	})
}

func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" {
		return 0, fmt.Errorf("amount is required")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount")
	}
	if v <= 0 {
		return 0, fmt.Errorf("amount must be > 0")
	}
	return int64(v * 100), nil
}
