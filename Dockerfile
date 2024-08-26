# Build stage
FROM golang:1.23-alpine as build

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .  

RUN go build -o run ./... 

# Development stage
FROM build as dev

RUN go install github.com/cosmtrek/air@v1.52.0

COPY . .

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM scratch as prod

WORKDIR /app

COPY --from=build /app/run /app/run 

CMD ["/app/run"]
