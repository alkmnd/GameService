FROM golang:1.20

WORKDIR /game-service
COPY ./ ./
RUN go mod download
RUN go build -o game-service ./cmd/main.go

ENTRYPOINT ["/game-service/game-service"]