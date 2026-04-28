package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	authpb "github.com/SlayerSv/payments/gen/auth"
	transpb "github.com/SlayerSv/payments/gen/trans"
	walletpb "github.com/SlayerSv/payments/gen/wallet"
	"github.com/SlayerSv/payments/internal/shared/errs"
	transmodels "github.com/SlayerSv/payments/internal/trans/models"
	walletmodels "github.com/SlayerSv/payments/internal/wallet/models"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AccountID struct {
	ID string `json:"id"`
}

type History struct {
	Transactions []transmodels.TransactionDTO `json:"transactions"`
}

// GetAccount gets user's wallet/savings account
// @Summary      Gets user's wallet/savings account information
// @Description  Gets user`s wallet/savings account information.
// @Tags         accounts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path string true "Account ID"
// @Success      200  {object}  models.AccountResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/accounts/{account_id} [get]
func (app *App) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accID, err := app.ExtractPathValue(r, "account_id")
	resp, err := app.Clients.Wallet.Get(ctx, &walletpb.GetRequest{Id: accID})
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
func (app *App) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := app.Clients.Wallet.GetAll(ctx, &walletpb.GetAllRequest{OwnerId: ""})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting accounts: %w", errs.Internal, err))
		return
	}
	accs := walletmodels.AccountsResponse{Accounts: []walletmodels.AccountResponse{}}
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
	resp, err := app.Clients.Wallet.Create(ctx, &walletpb.CreateRequest{OwnerId: ""})
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
// @Param        id  path string true "Account ID"
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
	_, err = app.Clients.Wallet.Delete(ctx, &walletpb.DeleteRequest{Id: req.ID})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error deleting account: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(204)
}

func (app *App) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	history, err := app.Clients.Trans.GetTransactionHistory(ctx, &emptypb.Empty{})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting account history: %w", errs.Internal, err))
		return
	}
	transHistory := History{Transactions: []transmodels.TransactionDTO{}}
	if len(history.Transactions) == 0 {
		app.Encode(w, r, transHistory)
		return
	}
	ids := map[string]string{}
	for _, trans := range history.Transactions {
		ids[trans.Id] = ""
	}
	idreq := &authpb.GetEmailsRequest{}
	for id := range ids {
		idreq.Ids = append(idreq.Ids, id)
	}
	idemail, err := app.Clients.User.GetEmails(ctx, idreq)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting emails: %w", errs.Internal, err))
		return
	}
	for _, trans := range history.Transactions {
		transHistory.Transactions = append(transHistory.Transactions, pbToTrans(trans, idemail.Emails))
	}
	app.Encode(w, r, transHistory)
}

func pbToAcc(acc *walletpb.Account) walletmodels.AccountResponse {
	return walletmodels.AccountResponse{
		ID:        acc.Id,
		OwnerID:   acc.OwnerId,
		Balance:   acc.Balance,
		CreatedAt: acc.CreateAt.AsTime(),
	}
}

func pbToTrans(trans *transpb.Transaction, idemail map[string]string) transmodels.TransactionDTO {
	return transmodels.TransactionDTO{
		ID:            trans.Id,
		OpType:        trans.OpType.String(),
		SenderID:      trans.SenderId,
		SenderEmail:   idemail[trans.SenderId],
		SenderType:    trans.SenderType.String(),
		ReceiverID:    trans.ReceiverId,
		ReceiverEmail: idemail[trans.ReceiverId],
		ReceiverType:  trans.ReceiverType.String(),
		Amount:        trans.Amount,
		CompletedAt:   trans.CreatedAt.AsTime(),
	}
}
