# syntax=docker/dockerfile:1

FROM golang:1.26.3 AS build

WORKDIR /build
COPY go.mod go.sum ./

# Скачиваем зависимости (кешируется между BuildKit-сборками)
RUN --mount=type=cache,id=go-mod-cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Копируем файлы, которые нужны в рантайме
RUN mkdir out && \
    mkdir out/database && \
    mkdir -p out/service/inspection/templates && \
    mv service/inspection/templates/ out/service/inspection/ && \
    mv database/migrations/ out/database/migrations/ && \
    mv .config/ out/

# Билдим гошечку в бинарник out/app
RUN --mount=type=cache,id=go-mod-cache,target=/go/pkg/mod \
    --mount=type=cache,id=go-build-cache,target=/root/.cache/go-build \
    go build -o out/app

FROM ubuntu:24.04

EXPOSE 80

WORKDIR /app

COPY --from=build /build/out ./

ENTRYPOINT ["./app"]
