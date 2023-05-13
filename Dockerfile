# Base
FROM golang:1.20-alpine AS base

RUN apk update && apk add --no-cache build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN make build

# Final
FROM alpine:3.17

WORKDIR /app

COPY --from=base /app/bin/api /app/api

CMD ["/app/api"]