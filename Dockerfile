FROM golang:1.17.5-alpine3.15

RUN apk add build-base

WORKDIR /app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["hellbot"]
