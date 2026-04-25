package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/errs"
)

type OTPRequest struct {
	Email string `json:"email"`
}

// SendOTP creates and sends one-time passwords
// @Summary      Creates a one-time password for a user with a certain email, saves it to a database and sends to the user
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Param        email body OTPRequest true "Email of a user"
// @Success      204
// @Router       /restore [post]
func (app *App) SendOTP(w http.ResponseWriter, r *http.Request) {
	t := OTPRequest{}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding email: %w", errs.BadRequest, err))
		return
	}
	t.Email = strings.TrimSpace(t.Email)
	_, err = app.Clients.Auth.Restore(r.Context(), &auth.RestoreRequest{Email: t.Email})
	if err != nil {
		app.ErrorJSON(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
