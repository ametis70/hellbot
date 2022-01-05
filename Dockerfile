FROM golang:1.17.5-alpine3.15
LABEL org.opencontainers.image.source="https://github.com/ametis70/hellbot"

RUN apk add build-base

WORKDIR /app
COPY . .
VOLUME /app/db

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["hellbot"]
