FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/ocr ./cmd/ocr

FROM gcr.io/distroless/static-debian11
COPY --from=builder /usr/local/bin/ocr /usr/local/bin/ocr
ENTRYPOINT ["/usr/local/bin/ocr"]
