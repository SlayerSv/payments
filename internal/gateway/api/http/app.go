package app

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/models"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email string) error
	Restore(ctx context.Context, email string) error
}

type UserService interface {
	Get(ctx context.Context) (models.UserDTO, error)
	Update(ctx context.Context, name, password *string) (models.UserDTO, error)
}

type TransService interface {
	Deposit(ctx context.Context, walletID string, amount int64) (int64, error)
	Withdraw(ctx context.Context, walletID string, amount int64) (int64, error)
	Transfer(ctx context.Context, donorWalletID, receiverWalletID string, amount int64) (int64, error)
	GetTransactionHistory(ctx context.Context, walletID string) (models.TransactionHistory, error)
}

type WalletService interface {
	Get(ctx context.Context, ID string) (models.WalletDTO, error)
	GetAll(ctx context.Context) ([]models.WalletDTO, error)
	Create(ctx context.Context) (string, error)
	Delete(ctx context.Context, ID string) error
}

type App struct {
	Log           *slog.Logger
	Server        *http.Server
	jwtkey        crypto.PublicKey
	authService   AuthService
	userService   UserService
	transService  TransService
	walletService WalletService
	Validator     *validator.Validate
}

func NewApp(logger *slog.Logger,
	server *http.Server,
	jwtkey crypto.PublicKey,
	authService AuthService,
	userService UserService,
	transService TransService,
	walletService WalletService,
	validator *validator.Validate) *App {
	return &App{
		Log:           logger,
		Server:        server,
		jwtkey:        jwtkey,
		authService:   authService,
		userService:   userService,
		transService:  transService,
		walletService: walletService,
		Validator:     validator,
	}
}

func (app *App) ErrorJSON(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	app.Log.ErrorContext(r.Context(), "Processing incoming request", slog.String("error", err.Error()), slog.String("method", r.Method), slog.String("url", r.URL.String()))
	var code int
	if errors.Is(err, errs.NotFound) {
		err = errs.NotFound
		code = http.StatusNotFound
	} else if errors.Is(err, errs.BadRequest) {
		err = errs.BadRequest
		code = http.StatusBadRequest
	} else if errors.Is(err, errs.InvalidCredentials) || errors.Is(err, errs.Unauthorized) {
		err = errs.Unauthorized
		code = http.StatusUnauthorized
	} else if errors.Is(err, errs.Forbidden) {
		err = errs.Forbidden
		code = http.StatusForbidden
	} else if errors.Is(err, errs.AlreadyExists) {
		err = errs.AlreadyExists
		code = http.StatusConflict
	} else {
		err = errs.Internal
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errs.Response{
		Error: err.Error(),
	})
}

func (app *App) Encode(w http.ResponseWriter, r *http.Request, obj any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(obj)
	if err != nil {
		app.ErrorJSON(w, r, err)
		return
	}
}

func (app *App) ExtractPathValue(r *http.Request, pathValue string) (string, error) {
	stringID := r.PathValue(pathValue)
	if stringID == "" {
		return "", fmt.Errorf("%w: empty path value(%s, %s)",
			errs.BadRequest, pathValue, stringID)
	}
	return stringID, nil
}

func (app *App) GetClaims(r *http.Request) (userID uuid.UUID, err error) {
	claims, ok := r.Context().Value(ClaimsKey).(jwt.RegisteredClaims)
	if !ok {
		err = fmt.Errorf("error getting claims")
		return
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return
	}
	userID, err = uuid.Parse(sub)
	return
}
