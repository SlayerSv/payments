package metrics

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InitMetricsServer запускает HTTP сервер на указанном порту для сбора метрик Prometheus
func InitMetricsServer(port string) {
	// Стандартный хэндлер прометея
	http.Handle("/metrics", promhttp.Handler())

	// Запускаем в горутине, чтобы не блокировать основной поток приложения
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			slog.Error("Start metrics server", slog.String("error", err.Error()))
			return
		}
	}()
}
