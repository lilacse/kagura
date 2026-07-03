FROM golang:1.25-alpine AS builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/kagura .

FROM debian:bookworm-slim

RUN groupadd -r app && useradd -g app -m app
RUN mkdir -p /data && chown app:app /data

COPY --from=builder /out/kagura /usr/local/bin/kagura

ENV KAGURA_DBPATH=/data/kagura.db

VOLUME /data

USER app

ENTRYPOINT /usr/local/bin/kagura
