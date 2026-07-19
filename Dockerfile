FROM golang:1.26-alpine
LABEL org.opencontainers.image.source="https://github.com/ametis70/hellbot"

RUN apk add build-base

WORKDIR /app
COPY . .
VOLUME /app/db

RUN go mod download
RUN go build -o hellbot .

CMD ["./hellbot"]
