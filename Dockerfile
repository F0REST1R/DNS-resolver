# 1. Этап сборки (builder)
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 2. Копируем только файлы зависимостей (для кэширования)
COPY go.mod go.sum ./
RUN go mod download

# 3. Копируем весь код и собираем приложение
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /dns-service ./cmd 

# 4. Этап запуска (минимальный образ)
FROM alpine:3.21.3

WORKDIR /

# 5. Копируем бинарник и сертификаты
COPY --from=builder /dns-service /dns-service
RUN apk add --no-cache ca-certificates

# 6. Указываем точку входа
EXPOSE 8080
ENTRYPOINT ["/dns-service"]
