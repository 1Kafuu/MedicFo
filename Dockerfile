# syntax=docker/dockerfile:1

# Tahap build: compiler Go hanya dipakai untuk menghasilkan binary.
FROM golang:1.26-alpine AS build
WORKDIR /src

# Salin go.mod/go.sum lebih dulu agar layer download dependency bisa di-cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binary statis agar bisa berjalan di image runtime yang minimal.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

# Tahap runtime: hanya binary, tanpa compiler dan source code.
FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/app ./app

# Logger menulis ke logs/app.log, jadi folder ini harus bisa ditulis user app.
RUN mkdir -p logs && chown app:app logs

USER app

EXPOSE 3000

ENTRYPOINT ["./app"]
