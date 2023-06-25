# syntax=docker/dockerfile:1

FROM golang:1.19-alpine AS builder

WORKDIR /build

COPY go.mod /build
COPY go.sum /build
COPY miniwiki.go /build

RUN go build -ldflags="-s -w" -o miniwiki

FROM alpine

WORKDIR /app

COPY --from=builder /build/miniwiki /app/miniwiki

RUN mkdir /pages

ENV NAME=wiki

CMD ./miniwiki -name ${NAME} -dir /pages -pass ${PASS}
