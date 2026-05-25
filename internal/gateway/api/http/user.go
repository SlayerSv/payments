package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
)

// GetUser gets user's wallet
// @Summary      Gets user's wallet information
// @Description  Gets user`s wallet information (name, email).
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.UserDTO
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me [get]
func (app *App) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := app.userService.Get(ctx)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting user: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, user)
}

// UpdateUser updates User if a user
// @Summary      Updates user
// @Description  Updates user (name, password).
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input body models.UpdateUserRequest true "New name and/or password"
// @Success      200  {object}  models.UserDTO
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me [patch]
func (app *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uur := &models.UpdateUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&uur)
	user, err := app.userService.Update(ctx, uur.Name, uur.Password)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error updating: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, user)
}
