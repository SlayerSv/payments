package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
)

// Login logs in a user
// @Summary      User Login
// @Description  Authenticates a user and signs a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email&password body models.LoginRequest true "Login Credentials"
// @Success      200  {object}  models.LoginResponse
// @Failure      400  {object}  errs.Response "Bad Request"
// @Failure      401  {object}  errs.Response "Unauthorized"
// @Router       /login [post]
func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	var lr models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding login credentials: %w", errs.BadRequest, err))
		return
	}
	ctx := r.Context()
	token, err := app.authService.Login(ctx, lr.Email, lr.Password)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error login: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, models.LoginResponse{Token: token})
}

// Register creates a new user
// @Summary      Register User
// @Description  Creates a new user wallet and sends a one time password to the user's email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email body models.EmailDTO true "Email of the user"
// @Success      201  {object}  uuid.UUID
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /register [post]
func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	var email models.EmailDTO
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding email: %w", errs.BadRequest, err))
		return
	}
	ctx := r.Context()
	err = app.authService.Register(ctx, email.Email)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error register: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
}

// Restore creates and sends one-time passwords
// @Summary      Creates a one-time password for a user with a certain email, saves it to a database and sends to the user
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Param        email body models.EmailDTO true "Email of a user"
// @Success      204
// @Router       /restore [post]
func (app *App) Restore(w http.ResponseWriter, r *http.Request) {
	email := models.EmailDTO{}
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding email: %w", errs.BadRequest, err))
		return
	}
	email.Email = strings.TrimSpace(email.Email)
	err = app.authService.Restore(r.Context(), email.Email)
	if err != nil {
		app.ErrorJSON(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
