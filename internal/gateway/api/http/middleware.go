package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/shared/grpc/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type contextKey string

const ClaimsKey contextKey = "claims"

var (
	idRegex = regexp.MustCompile(`/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}|/\d+`)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Общее количество входящих HTTP запросов на API Gateway",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_http_request_duration_seconds",
			Help:    "Время обработки HTTP запросов на API Gateway в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// Обертка над writer для отслеживания статус кодов
type responseWriterDelegator struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterDelegator) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// sanitizePath превращает "/wallet/123" или "/wallet/de305d54-..." в "/wallet/{id}"
func sanitizePath(path string) string {
	return idRegex.ReplaceAllString(path, "/{id}")
}

// HTTPMetricsMiddleware оборачивает http.Handler и собирает метрики
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Игнорируем сам роут /metrics, чтобы не спамить в графики запросами самого Прометея/Виктории
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Оборачиваем оригинальный ResponseWriter (по умолчанию статус 200 OK)
		delegator := &responseWriterDelegator{ResponseWriter: w, statusCode: http.StatusOK}

		// Передаем управление следующему обработчику (gRPC-Gateway)
		next.ServeHTTP(delegator, r)

		// Считаем время
		duration := time.Since(start).Seconds()

		path := sanitizePath(r.URL.Path)
		method := r.Method
		statusStr := strconv.Itoa(delegator.statusCode)

		// Записываем метрики
		httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
		httpDuration.WithLabelValues(method, path).Observe(duration)
	})
}

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
