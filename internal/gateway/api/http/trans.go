package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
)

// Deposit Deposits funds to a specified wallet
// @Summary      Deposits funds to a specified wallet
// @Description  Deposits funds to a specified wallet
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "wallet ID"
// @Param        amount body models.DepositRequest true "Amount to deposit"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id}/deposit [post]
func (app *App) Deposit(w http.ResponseWriter, r *http.Request) {
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	req := models.DepositRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding deposit request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating deposit request: %w", errs.BadRequest, err))
		return
	}
	newBalance, err := app.transService.Deposit(r.Context(), walletID, req.Amount)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error depositing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: newBalance})
}

// Withdraw Withdraws funds from a specified wallet
// @Summary      Withdraws funds from a specified wallet
// @Description  Withdraws funds from a specified wallet
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "wallet ID"
// @Param        amount body models.WithdrawRequest true "Amount to withdraw"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id}/withdraw [post]
func (app *App) Withdraw(w http.ResponseWriter, r *http.Request) {
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	req := models.WithdrawRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding withdraw request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating withdraw request: %w", errs.BadRequest, err))
		return
	}
	newBalance, err := app.transService.Withdraw(r.Context(), walletID, req.Amount)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error withdrawing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: newBalance})
}

// Transfer Transfers funds to a specified wallet
// @Summary      Transfers funds to a specified wallet
// @Description  Transfers funds to a specified wallet
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "wallet ID"
// @Param        details body models.TransferRequest true "Transfer details"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id}/transfer [post]
func (app *App) Transfer(w http.ResponseWriter, r *http.Request) {
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	req := models.TransferRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding transfer request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating transfer request: %w", errs.BadRequest, err))
		return
	}
	newBalance, err := app.transService.Transfer(r.Context(), walletID, req.ReceiverWalletID, req.Amount)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error Transfering: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: newBalance})
}

// GetTransactionHistory Gets history of transactions of an wallet
// @Summary      Gets history of transactions of an wallet
// @Description  Gets history of transactions of an wallet
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "wallet ID"
// @Success      200  {object} models.TransactionHistory
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id}/transactions [get]
func (app *App) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	history, err := app.transService.GetTransactionHistory(ctx, walletID)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting wallet history: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, history)
}
