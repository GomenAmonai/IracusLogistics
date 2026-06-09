# syntax=docker/dockerfile:1

# --- build stage -----------------------------------------------------------
FROM golang:1.26 AS build
WORKDIR /src

# Слой зависимостей кэшируется отдельно: пересобирается только при изменении go.mod/go.sum.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO выключен → статический бинарь под distroless static. -trimpath/-ldflags убирают пути и
# отладочную информацию (меньше образ, не светим локальные пути сборки).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# --- runtime stage ---------------------------------------------------------
# distroless static: нет shell и пакетного менеджера (меньше поверхность атаки), запуск от nonroot.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
# Миграции нужны бинарю migrate (file://migrations резолвится относительно WORKDIR /app).
COPY --from=build /src/migrations /app/migrations

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]
