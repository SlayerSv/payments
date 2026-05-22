package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	authpb "github.com/SlayerSv/payments/gen/auth"
	pb "github.com/SlayerSv/payments/gen/trans"
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
	accID, err := app.ExtractPathValue(r, "wallet_id")
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
	resp, err := app.Clients.Trans.Deposit(r.Context(), &pb.DepositRequest{
		WalletId: accID,
		Amount:   req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error depositing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
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
	accID, err := app.ExtractPathValue(r, "wallet_id")
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
	resp, err := app.Clients.Trans.Withdraw(r.Context(), &pb.WithdrawRequest{
		WalletId: accID,
		Amount:   req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error withdrawing: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
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
	accID, err := app.ExtractPathValue(r, "wallet_id")
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
	resp, err := app.Clients.Trans.Transfer(r.Context(), &pb.TransferRequest{
		DonorWalletId:    accID,
		ReceiverWalletId: req.ReceiverWalletID,
		Amount:           req.Amount,
	})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error Transfering: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.NewBalanceResponse{NewBalance: resp.NewBalance})
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
	accID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	history, err := app.Clients.Trans.GetTransactionHistory(ctx, &pb.GetTransactionHistoryRequest{WalletId: accID})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting wallet history: %w", errs.Internal, err))
		return
	}
	transHistory := models.TransactionHistory{Transactions: []models.TransactionDTO{}}
	if len(history.Transactions) == 0 {
		app.Encode(w, r, transHistory)
		return
	}
	ids := map[string]string{}
	for _, trans := range history.Transactions {
		if trans.DonorWalletId != nil {
			ids[*trans.DonorWalletId] = ""
		}
		if trans.ReceiverWalletId != nil {
			ids[*trans.ReceiverWalletId] = ""
		}
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

func pbToTrans(trans *pb.Transaction, idemail map[string]string) models.TransactionDTO {
	tr := models.TransactionDTO{}
	tr.ID = trans.Id
	tr.OpType = trans.OpType.String()
	tr.Amount = trans.Amount
	tr.CompletedAt = trans.CreatedAt.String()
	if trans.OpType == pb.OperationType_DEPOSIT || trans.OpType == pb.OperationType_TRANSFER {
		tr.ReceiverWalletID = trans.ReceiverWalletId
	}
	if trans.OpType == pb.OperationType_WITHDRAW || trans.OpType == pb.OperationType_TRANSFER {
		tr.DonorWalletID = trans.DonorWalletId
	}
	return tr
}
