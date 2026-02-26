FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/mc-link-bot ./cmd/mc-link-bot

FROM alpine:3.21
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=build /out/mc-link-bot /usr/local/bin/mc-link-bot
ENTRYPOINT ["/usr/local/bin/mc-link-bot"]
