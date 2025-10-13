FROM golang:1.25 AS build

WORKDIR /build
COPY . .

# Скачиваем зависимости (кешируется в слое Docker)
RUN go mod download

# Копируем файлы, которые нужны в рантайме
RUN mkdir out && \
    mkdir out/database && \
    mkdir -p out/service/inspection/templates && \
    mv service/inspection/templates/ out/service/inspection/ && \
    mv database/migrations/ out/database/migrations/ && \
    mv .config/ out/

# Билдим гошечку в бинарник out/app
RUN go build -o out/app

FROM ubuntu:24.04

EXPOSE 80

WORKDIR /app

COPY --from=build /build/out ./

ENTRYPOINT ["./app"]
