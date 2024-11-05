# Build stage
FROM golang:1.23 AS builder
WORKDIR /app

# Air for live reloading
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build -o application ./cmd/auction

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/application .

EXPOSE 3100

CMD ["./application"]

