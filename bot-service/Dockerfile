FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod init botservice || true
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/bot ./cmd/main.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /usr/local/bin/bot /usr/local/bin/bot
ENTRYPOINT ["/usr/local/bin/bot"]
