package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/SlayerSv/payments/gen/trans"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/trans/models"
)

func (app *App) Deposit(w http.ResponseWriter, r *http.Request) {
	req := models.DepositRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding deposit request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating deposit request: %w", errs.BadRequest, err))
		return
	}
	_, err = app.Clients.Trans.Deposit(r.Context(), &pb.DepositRequest{
		AccountType: pb.AccountType(pb.AccountType_value[req.AccountType]),
		Amount:      req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error depositing: %w", errs.Internal, err))
		return
	}
}

func (app *App) Withdraw(w http.ResponseWriter, r *http.Request) {
	req := models.WithdrawRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding withdraw request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating withdraw request: %w", errs.BadRequest, err))
		return
	}
	_, err = app.Clients.Trans.Withdraw(r.Context(), &pb.WithdrawRequest{
		AccountType: pb.AccountType(pb.AccountType_value[req.AccountType]),
		Amount:      req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error withdrawing: %w", errs.Internal, err))
		return
	}
}

func (app *App) Transfer(w http.ResponseWriter, r *http.Request) {
	req := models.TransferRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding transfer request: %w", errs.BadRequest, err))
		return
	}
	err = app.Validator.Struct(req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error validating transfer request: %w", errs.BadRequest, err))
		return
	}
	_, err = app.Clients.Trans.Transfer(r.Context(), &pb.TransferRequest{
		SenderId:      req.SenderID,
		SenderType:    pb.AccountType(pb.AccountType_value[req.SenderType]),
		ReceiverEmail: req.ReceiverEmail,
		ReceiverType:  pb.AccountType(pb.AccountType_value[req.ReceiverType]),
		Amount:        req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error Transfering: %w", errs.Internal, err))
		return
	}
}
