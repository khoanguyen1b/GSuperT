FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata bash

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOBIN=/usr/local/bin

# Hot reload
RUN go install github.com/air-verse/air@v1.61.7

# Migrations (postgres)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]