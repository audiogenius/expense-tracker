FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/api ./cmd/api

FROM gcr.io/distroless/static-debian11
COPY --from=builder /usr/local/bin/api /usr/local/bin/api
ENTRYPOINT ["/usr/local/bin/api"]
