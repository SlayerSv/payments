package app

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/SlayerSv/payments/internal/gateway/clients"
	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/logger"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type App struct {
	Log       logger.Logger
	Server    *http.Server
	jwtkey    crypto.PublicKey
	Clients   *clients.Clients
	Validator *validator.Validate
}

func NewApp(logger logger.Logger, server *http.Server, jwtkey crypto.PublicKey, clients *clients.Clients, validator *validator.Validate) *App {
	return &App{Log: logger, Server: server, jwtkey: jwtkey, Clients: clients, Validator: validator}
}

func (app *App) ErrorJSON(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	app.Log.Errorln(r.Method, r.URL, err.Error())
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
