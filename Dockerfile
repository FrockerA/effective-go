# Этап 1: Сборка (Builder)
FROM golang:alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник. CGO_ENABLED=0 нужен для корректной работы в alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

# Этап 2: Финальный легковесный образ
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из первого этапа
COPY --from=builder /app/main .
# ВАЖНО: Копируем папку с миграциями, так как наш код ищет их по пути file://migrations
COPY --from=builder /app/migrations ./migrations

# Открываем порт, на котором крутится приложение
EXPOSE 8080

# Запускаем бинарник
CMD ["./main"]