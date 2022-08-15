ARG GOLANG_VERSION="1.18-alpine"
ARG ALPINE_VERSION="3.16"

FROM golang:$GOLANG_VERSION AS builder

RUN apk update

WORKDIR /app

COPY . .
RUN go mod download

RUN mkdir -p bin
RUN CGO_ENABLED=0 go build -trimpath -o bin/ -v ./cmd/server/...
RUN chmod -R +x ./bin

FROM alpine:$ALPINE_VERSION
COPY --from=builder /app/bin /app
WORKDIR /app
ENTRYPOINT ["./server"]