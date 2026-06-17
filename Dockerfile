# syntax=docker/dockerfile:1

FROM golang:1.26.3 AS build

WORKDIR /build
COPY go.mod go.sum ./

ENV GOPRIVATE=github.com/sunshineOfficial/* \
    GONOSUMDB=github.com/sunshineOfficial/*

# Скачиваем зависимости (кешируется между BuildKit-сборками). Для приватного
# github.com/sunshineOfficial/golib передайте BuildKit secret github_token.
RUN --mount=type=cache,id=go-mod-cache,target=/go/pkg/mod \
    --mount=type=secret,id=github_token,required=true \
    GITHUB_TOKEN="$(cat /run/secrets/github_token)" && \
    git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/" && \
    trap 'git config --global --unset-all url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf || true' EXIT && \
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
