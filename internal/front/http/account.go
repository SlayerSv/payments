package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	authmodels "github.com/SlayerSv/payments/internal/auth/models"
	transmodels "github.com/SlayerSv/payments/internal/trans/models"
	walletmodels "github.com/SlayerSv/payments/internal/wallet/models"
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

	var me authmodels.UserDTO
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me", token, nil, &me); err != nil {
		a.render(w, "me", PageData{Title: "Личный кабинет", Error: err.Error()})
		return
	}

	var accounts walletmodels.AccountsResponse
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/accounts", token, nil, &accounts)
	a.render(w, "me", PageData{
		Title:    "Личный кабинет",
		Authed:   true,
		User:     me,
		Accounts: accounts.Accounts,
	})
}

func (a *App) meUpdate(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	req := authmodels.UpdateUserRequest{}
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

func (a *App) accountsPage(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	var accounts walletmodels.AccountsResponse
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me/accounts", token, nil, &accounts); err != nil {
		a.render(w, "accounts", PageData{Title: "Счета", Authed: true, Error: err.Error()})
		return
	}

	a.render(w, "accounts", PageData{Title: "Счета", Authed: true, Accounts: accounts.Accounts})
}

func (a *App) createAccount(w http.ResponseWriter, r *http.Request) {
	token := a.token(r)

	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/accounts", token, map[string]any{}, nil); err != nil {
		a.render(w, "accounts", PageData{Title: "Счета", Authed: true, Error: err.Error()})
		return
	}
	http.Redirect(w, r, "/me/accounts", http.StatusFound)
}

func (a *App) accountPage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("account_id")
	token := a.token(r)

	var account walletmodels.AccountResponse
	if _, _, err := a.api(r.Context(), http.MethodGet, "/me/accounts/"+id, token, nil, &account); err != nil {
		a.render(w, "account", PageData{Title: "Счет", Authed: true, Error: err.Error(), Account: account})
		return
	}

	var tx transmodels.TransactionHistory
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/accounts/"+id+"/transactions", token, nil, &tx)

	a.render(w, "account", PageData{
		Title:        "Счет",
		Authed:       true,
		Account:      account,
		Transactions: tx.Transactions,
	})
}

func (a *App) deposit(w http.ResponseWriter, r *http.Request)  { a.accountAction(w, r, "/deposit") }
func (a *App) withdraw(w http.ResponseWriter, r *http.Request) { a.accountAction(w, r, "/withdraw") }

func (a *App) transfer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("account_id")
	token := a.token(r)

	amount, err := parseAmount(r.FormValue("amount"))
	if err != nil {
		a.accountError(w, r, id, err)
		return
	}
	payload := transmodels.TransferRequest{
		Amount:            amount,
		ReceiverAccountID: strings.TrimSpace(r.FormValue("to_account_id")),
	}
	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/accounts/"+id+"/transfer", token, payload, nil); err != nil {
		a.accountError(w, r, id, err)
		return
	}
	http.Redirect(w, r, "/me/accounts/"+id, http.StatusFound)
}

func (a *App) accountAction(w http.ResponseWriter, r *http.Request, suffix string) {
	id := r.PathValue("account_id")
	token := a.token(r)

	amount, err := parseAmount(r.FormValue("amount"))
	if err != nil {
		a.accountError(w, r, id, err)
		return
	}

	if _, _, err := a.api(r.Context(), http.MethodPost, "/me/accounts/"+id+suffix, token, transmodels.DepositRequest{
		Amount: amount,
	}, nil); err != nil {
		a.accountError(w, r, id, err)
		return
	}
	http.Redirect(w, r, "/me/accounts/"+id, http.StatusFound)
}

func (a *App) accountError(w http.ResponseWriter, r *http.Request, id string, err error) {
	token := a.token(r)

	var account walletmodels.AccountResponse
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/accounts/"+id, token, nil, &account)

	var tx []transmodels.TransactionDTO
	_, _, _ = a.api(r.Context(), http.MethodGet, "/me/accounts/"+id+"/transactions", token, nil, &tx)

	a.render(w, "account", PageData{
		Title:        "Счет",
		Authed:       true,
		Error:        err.Error(),
		Account:      account,
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
