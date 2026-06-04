package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type VictoriaWriter struct {
	client      *http.Client
	url         string
	logChan     chan map[string]any
	serviceName string
}

func NewVictoriaLogger(serviceName string) (*slog.Logger, func()) {
	endpoint := os.Getenv("VICTORIA_LOGS_URL")

	// API Виктории для приема JSON Lines. Указываем, что 'service' — это поле стрима
	apiURL := fmt.Sprintf("%s/insert/jsonline?_stream_fields=service", endpoint)

	writer := &VictoriaWriter{
		client:      &http.Client{Timeout: 5 * time.Second},
		url:         apiURL,
		logChan:     make(chan map[string]any, 1000),
		serviceName: serviceName,
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Запускаем фоновый сборщик логов
	go writer.worker(ctx)

	// Настраиваем вывод: пишем и в консоль (красивый JSON), и отправляем в Викторию
	slogHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Виктория требует, чтобы сообщение лежало в '_msg', а время в '_time'
			if a.Key == slog.MessageKey {
				a.Key = "_msg"
			}
			if a.Key == slog.TimeKey {
				a.Key = "_time"
			}
			return a
		},
	})

	// Оборачиваем логгер, чтобы он дублировал запись в наш канал
	logger := slog.New(&chanHandler{
		Handler: slogHandler,
		writer:  writer,
	})

	// Функция закрытия (нужно вызвать при выходе из приложения)
	cleanup := func() {
		cancel()
		writer.flush()
	}

	return logger, cleanup
}

// Хэндлер, который перехватывает логи и шлет их в канал
type chanHandler struct {
	slog.Handler
	writer *VictoriaWriter
}

func (h *chanHandler) Handle(ctx context.Context, r slog.Record) error {
	// Отправляем в стандартный хэндлер (вывод в консоль)
	err := h.Handler.Handle(ctx, r)
	if err != nil {
		return err
	}

	// Формируем карту для отправки в Викторию
	logEntry := map[string]any{
		"_msg":    r.Message,
		"_time":   r.Time.Format(time.RFC3339Nano),
		"level":   r.Level.String(),
		"service": h.writer.serviceName,
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		logEntry["trace_id"] = spanCtx.TraceID().String()
	}

	// Вытаскиваем trace_id из контекста, если он там есть (для связки логов и трейсов!)
	r.Attrs(func(a slog.Attr) bool {
		logEntry[a.Key] = a.Value.Any()
		return true
	})

	if spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}

	select {
	case h.writer.logChan <- logEntry:
	default:
		// Если канал переполнен, просто дропаем, чтобы не вешать приложение
	}

	return nil
}

// Воркер собирает логи пачками и шлет по HTTP
func (w *VictoriaWriter) worker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var batch []map[string]any

	for {
		select {
		case logEntry := <-w.logChan:
			batch = append(batch, logEntry)
			if len(batch) >= 50 {
				w.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				w.sendBatch(batch)
				batch = nil
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *VictoriaWriter) sendBatch(batch []map[string]any) {
	var buf bytes.Buffer
	for _, entry := range batch {
		b, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		buf.Write(b)
		buf.WriteByte('\n') // Формат JSON Lines требует переноса строки
	}

	req, err := http.NewRequest("POST", w.url, &buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (w *VictoriaWriter) flush() {
	close(w.logChan)
	var batch []map[string]any
	for logEntry := range w.logChan {
		batch = append(batch, logEntry)
	}
	if len(batch) > 0 {
		w.sendBatch(batch)
	}
}
