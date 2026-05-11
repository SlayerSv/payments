# --- ЭТАП 1: Сборка (Builder) ---
FROM golang:1.26.3-alpine3.22 AS builder

WORKDIR /app

# Скачиваем зависимости (кешируется докером, если go.mod не менялся)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download


# Копируем исходный код
COPY . .

# Получаем имя сервиса при сборке
ARG SERVICE_NAME

# Собираем бинарник. CGO_ENABLED=0 нужен для запуска в пустом alpine
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o /app/service ./cmd/${SERVICE_NAME}/main.go

# --- ЭТАП 2: Запуск (Runner) ---
FROM alpine:3.22
WORKDIR /app

# Добавляем сертификаты для HTTPS и таймзоны
RUN apk --no-cache add ca-certificates tzdata

# Копируем готовый бинарник из первого этапа
COPY --from=builder /app/service .

CMD ["./service"]