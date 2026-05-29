package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InitMetricsServer запускает HTTP сервер на указанном порту для сбора метрик Prometheus
func InitMetricsServer(port string) {
	// Стандартный хэндлер прометея
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting metrics server on port %s", port)
	// Запускаем в горутине, чтобы не блокировать основной поток приложения
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}
