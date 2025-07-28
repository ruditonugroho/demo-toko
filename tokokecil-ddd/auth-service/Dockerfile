# Dockerfile - Auth Service (Go + MongoDB)
FROM golang:1.23-alpine as build

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o auth-service cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/auth-service /app/auth-service
COPY .env .env

EXPOSE 8081
CMD ["./auth-service"]
