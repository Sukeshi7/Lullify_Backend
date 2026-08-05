# ── Build ──
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server .

# ── Run server ──
FROM alpine:3.19
COPY --from=builder /bin/server /server
RUN chmod +x /server
EXPOSE 8080
ENTRYPOINT ["/server"]