FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main cmd/sub/main.go

FROM alpine:latest  

WORKDIR /app

COPY --from=builder /app/main .

COPY tern.conf tern.conf
COPY config/local.yaml config/local.yaml
COPY docs docs
COPY migrations migrations

EXPOSE 8080


CMD ["./main"]
