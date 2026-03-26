FROM golang:latest AS build

WORKDIR /app

COPY go.mod ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api_gateway ./cmd/api_gateway


FROM alpine:latest

WORKDIR /app

COPY --from=build /app/api_gateway .

ENTRYPOINT ["./api_gateway"]
