# Build stage
FROM golang:1.23-alpine as build

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .  

RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN go build -o run ./cmd/server/main.go

# Development stage
FROM build as dev

RUN go install github.com/cosmtrek/air@v1.52.0

COPY . .

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM golang:1.23-alpine as prod

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY --from=build /app/run /app/run 
COPY --from=build /app/docs /app/docs
COPY --from=build /app/migrations /app/migrations
CMD ["/app/run"]
