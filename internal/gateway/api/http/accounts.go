package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/wallet/models"
)

type AccountID struct {
	ID string `json:"id"`
}

// GetAccount gets user's wallet/savings account
// @Summary      Gets user's wallet/savings account information
// @Description  Gets user`s wallet/savings account information.
// @Tags         accounts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.AccountResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts/{account_id} [get]
func (app *App) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accID, err := app.ExtractPathValue(r, "account_id")
	resp, err := app.Clients.Wallet.GetAccount(ctx, &pb.GetAccountRequest{Id: accID})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting account: %w", errs.Internal, err))
		return
	}
	account := pbToAcc(resp.GetAccount())
	app.Encode(w, r, account)
}

// GetAccounts gets user's all wallet/savings accounts
// @Summary      Gets user's all wallet/savings accounts
// @Description  Gets user`s all wallet/savings accounts
// @Tags         accounts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.AccountsResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts [get]
func (app *App) GetAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := app.Clients.Wallet.GetAccounts(ctx, &pb.GetAccountsRequest{OwnerId: ""})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting accounts: %w", errs.Internal, err))
		return
	}
	accs := models.AccountsResponse{Accounts: []models.AccountResponse{}}
	for _, acc := range resp.Accounts {
		accs.Accounts = append(accs.Accounts, pbToAcc(acc))
	}
	app.Encode(w, r, accs)
}

// CreateAccount Creates wallet/savings account
// @Summary      Creates wallet/savings account
// @Description  Creates wallet/savings account
// @Tags         accounts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      201  {object}  AccountID
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts [post]
func (app *App) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := app.Clients.Wallet.CreateAccount(ctx, &pb.CreateAccountRequest{OwnerId: ""})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error creating account: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, AccountID{resp.GetAccountId()})
}

// DeleteAccount Deletes wallet/savings account
// @Summary      Deletes wallet/savings account
// @Description  Deletes wallet/savings account
// @Tags         accounts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts [delete]
func (app *App) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := AccountID{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding request: %w", errs.BadRequest, err))
		return
	}
	_, err = app.Clients.Wallet.DeleteAccount(ctx, &pb.DeleteAccountRequest{Id: req.ID})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error deleting account: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(204)
}

func pbToAcc(acc *pb.Account) models.AccountResponse {
	return models.AccountResponse{
		ID:        acc.Id,
		OwnerID:   acc.OwnerId,
		Balance:   acc.Balance,
		CreatedAt: acc.CreateAt.AsTime(),
	}
}
