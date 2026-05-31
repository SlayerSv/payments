package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer настраивает отправку трейсов в VictoriaTraces
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Адрес VictoriaTraces (можно прокинуть через ENV)
	endpoint := os.Getenv("OTLP_ENDPOINT")

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, err
	}

	// Указываем имя сервиса (чтобы в Графане было видно, кто генерит спан)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	// Создаем провайдер трейсов
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Устанавливаем его как глобальный провайдер в приложении
	otel.SetTracerProvider(tp)
	// Настраиваем глобальный упаковщик/распаковщик заголовков
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	log.Printf("Трейсинг инициализирован для сервиса %s, шлем в %s", serviceName, endpoint)
	return tp, nil
}
