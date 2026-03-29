# build stage 
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o docqa-api ./cmd/main.go

# runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl poppler-utils

WORKDIR /app
COPY --from=builder /app/docqa-api .

EXPOSE 8080
ENTRYPOINT ["./docqa-api"]
