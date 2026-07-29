FROM golang:1.26 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hellbot ./cmd/hellbot

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/ametis70/hellbot"

COPY --from=builder /app/hellbot /hellbot
ENTRYPOINT ["/hellbot"]
