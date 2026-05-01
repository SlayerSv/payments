package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/trans/models"
)

// Deposit Deposits funds to a specified account
// @Summary      Deposits funds to a specified account
// @Description  Deposits funds to a specified account
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        account_id  path string true "Account ID"
// @Param        amount body models.DepositRequest true "Amount to deposit"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts/{account_id}/deposit [post]
func (app *App) Deposit(w http.ResponseWriter, r *http.Request) {
	accID, err := app.ExtractPathValue(r, "account_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting account id from path: %w", errs.BadRequest, err))
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
	resp, err := app.Clients.Trans.Deposit(r.Context(), &pb.DepositRequest{
		AccountId:   accID,
		AccountType: pb.AccountType(pb.AccountType_value[req.AccountType]),
		Amount:      req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error depositing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
}

// Withdraw Withdraws funds from a specified account
// @Summary      Withdraws funds from a specified account
// @Description  Withdraws funds from a specified account
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        account_id  path string true "Account ID"
// @Param        amount body models.WithdrawRequest true "Amount to withdraw"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts/{account_id}/withdraw [post]
func (app *App) Withdraw(w http.ResponseWriter, r *http.Request) {
	accID, err := app.ExtractPathValue(r, "account_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting account id from path: %w", errs.BadRequest, err))
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
	resp, err := app.Clients.Trans.Withdraw(r.Context(), &pb.WithdrawRequest{
		AccountId:   accID,
		AccountType: pb.AccountType(pb.AccountType_value[req.AccountType]),
		Amount:      req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error withdrawing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
}

// Transfer Transfers funds to a specified account
// @Summary      Transfers funds to a specified account
// @Description  Transfers funds to a specified account
// @Tags         transactions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        account_id  path string true "Account ID"
// @Param        details body models.TransferRequest true "Transfer details"
// @Success      201 {object} models.NewBalanceResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts/{account_id}/transfer [post]
func (app *App) Transfer(w http.ResponseWriter, r *http.Request) {
	accID, err := app.ExtractPathValue(r, "account_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting account id from path: %w", errs.BadRequest, err))
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
	resp, err := app.Clients.Trans.Transfer(r.Context(), &pb.TransferRequest{
		DonorAccountId:      accID,
		DonorAccountType:    pb.AccountType(pb.AccountType_value[req.DonorAccountType]),
		ReceiverAccountId:   req.ReceiverAccountID,
		ReceiverAccountType: pb.AccountType(pb.AccountType_value[req.ReceiverAccountType]),
		Amount:              req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error Transfering: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
}
