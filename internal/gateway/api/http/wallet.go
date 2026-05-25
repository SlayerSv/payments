package app

import (
	"fmt"
	"net/http"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
)

// GetWallet gets user's wallet
// @Summary      Gets user's wallet information
// @Description  Gets user`s wallet information.
// @Tags         wallets
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "Wallet ID"
// @Success      200  {object}  models.WalletDTO
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id} [get]
func (app *App) GetWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error extracting wallet id from path: %w", errs.BadRequest, err))
		return
	}
	wallet, err := app.walletService.Get(ctx, walletID)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting wallet: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, wallet)
}

// GetAllWallets gets all user's wallets
// @Summary      Gets all user's wallets
// @Description  Gets all user`s wallets
// @Tags         wallets
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.WalletsDTO
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets [get]
func (app *App) GetAllWallets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wallets, err := app.walletService.GetAll(ctx)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting wallets: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, wallets)
}

// CreateWallet Creates wallet
// @Summary      Creates wallet
// @Description  Creates wallet
// @Tags         wallets
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      201  {object}  models.WalletID
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets [post]
func (app *App) CreateWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ID, err := app.walletService.Create(ctx)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error creating wallet: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
	app.Encode(w, r, models.WalletID{ID: ID})
}

// DeleteWallet Deletes wallet
// @Summary      Deletes wallet
// @Description  Deletes wallet
// @Tags         wallets
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        wallet_id  path string true "Wallet ID"
// @Success      204
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me/wallets/{wallet_id} [delete]
func (app *App) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletID, err := app.ExtractPathValue(r, "wallet_id")
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: missing id: %w", errs.BadRequest, err))
		return
	}
	err = app.walletService.Delete(ctx, walletID)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error deleting wallet: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(204)
}
