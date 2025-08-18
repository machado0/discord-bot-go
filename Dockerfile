FROM golang:1.24.2-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY cmd/discord-bot/.env .env  
RUN go build -o /app/main ./cmd/discord-bot
CMD ["/app/main"]