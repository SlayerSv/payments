package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
	"google.golang.org/protobuf/types/known/emptypb"
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
	resp, err := app.Clients.Auth.Login(ctx, &auth.LoginRequest{Email: lr.Email, Password: lr.Password})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error login: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, models.LoginResponse{Token: resp.GetToken()})
}

// Register creates a new user
// @Summary      Register User
// @Description  Creates a new user wallet and sends a one time password to the user's email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email body models.RegisterRequest true "Email of the user"
// @Success      201  {object}  uuid.UUID
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /register [post]
func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	var rr models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&rr)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding email: %w", errs.BadRequest, err))
		return
	}
	ctx := r.Context()
	_, err = app.Clients.Auth.Register(ctx, &auth.RegisterRequest{Email: rr.Email})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error register: %w", errs.Internal, err))
		return
	}
	w.WriteHeader(201)
}

type OTPRequest struct {
	Email string `json:"email"`
}

// Restore creates and sends one-time passwords
// @Summary      Creates a one-time password for a user with a certain email, saves it to a database and sends to the user
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Param        email body OTPRequest true "Email of a user"
// @Success      204
// @Router       /restore [post]
func (app *App) Restore(w http.ResponseWriter, r *http.Request) {
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

// GetUser gets user's wallet
// @Summary      Gets user's wallet information
// @Description  Gets user`s wallet information (name, email).
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me [get]
func (app *App) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := app.Clients.User.Get(ctx, &emptypb.Empty{})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error getting user: %w", errs.Internal, err))
		return
	}
	user := models.UserDTO{
		ID:        resp.Id,
		Email:     resp.Email,
		Name:      resp.Name,
		CreatedAt: resp.CreatedAt.String(),
		UpdatedAt: resp.UpdatedAt.String(),
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
// @Success      200  {object}  models.User
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /me [patch]
func (app *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &models.UpdateUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	resp, err := app.Clients.User.Update(ctx, &auth.UpdateRequest{NewName: body.Name, NewPassword: body.Password})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error updating: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, resp)
}
