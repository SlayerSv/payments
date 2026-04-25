package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func (app *App) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			app.Log.Errorf("error reading request body: %v", err)
		}
		defer r.Body.Close()
		app.Log.Infof("Incoming request:\n%s %s\n%s",
			r.Method, r.URL, string(body))
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func (app *App) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				app.ErrorJSON(w, r, errs.Internal)
				app.Log.Errorf("error: %v, stack trace: %s", err, string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *App) Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		tokenStrTrim := strings.TrimPrefix(tokenStr, "Bearer ")
		if strings.TrimSpace(tokenStrTrim) == "" {
			app.ErrorJSON(w, r, fmt.Errorf("%w: missing token(%s)", errs.Unauthorized, tokenStr))
			return
		}
		ctx := context.WithValue(r.Context(), interceptors.JWTKey, tokenStrTrim)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
