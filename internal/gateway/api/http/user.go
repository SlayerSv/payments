package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SlayerSv/payments/gen/auth"
	"github.com/SlayerSv/payments/internal/shared/errs"
)

type LoginRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password"`
}

type RegisterRequest struct {
	Email string `json:"email"`
}

type UpdateUser struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
}

// Login logs in a user
// @Summary      User Login
// @Description  Authenticates a user and signs a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email&password body LoginRequest true "Login Credentials"
// @Success      200  {object}  string
// @Failure      400  {object}  errs.Response "Bad Request"
// @Failure      401  {object}  errs.Response "Unauthorized"
// @Router       /login [post]
func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding login credentials: %w", errs.BadRequest, err))
		return
	}
	ctx := r.Context()
	resp, err := app.Clients.Auth.Login(ctx, &auth.LoginRequest{Email: lr.Email, Password: lr.PasswordHash})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error login: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, resp.GetToken())
}

// Register creates a new user
// @Summary      Register User
// @Description  Creates a new user account and sends a one time password to the user's email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email body RegisterRequest true "Email of the user"
// @Success      201  {object}  uuid.UUID
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /register [post]
func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	var rr RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&rr)
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error decoding email: %w", errs.BadRequest, err))
		return
	}
	ctx := r.Context()
	fmt.Println(rr)
	resp, err := app.Clients.Auth.Register(ctx, &auth.RegisterRequest{Email: rr.Email})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error register: %w", errs.Internal, err))
		return
	}
	fmt.Println("ok")
	w.WriteHeader(201)
	app.Encode(w, r, resp.GetStatus())
}

// UpdateUser updates User if a user
// @Summary      Updates user
// @Description  Updates user (name, password).
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input body UpdateUser true "New name and/or password"
// @Success      200  {object}  models.User
// @Failure      400  {object}  errs.Response "Bad Request"
// @Router       /users [patch]
func (app *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &UpdateUser{}
	err := json.NewDecoder(r.Body).Decode(&body)
	resp, err := app.Clients.User.UpdateUser(ctx, &auth.UpdateUserRequest{NewName: body.Name, NewPassword: body.Password})
	if err != nil {
		app.ErrorJSON(w, r, fmt.Errorf("%w: error updating: %w", errs.Internal, err))
		return
	}
	app.Encode(w, r, resp)
}
