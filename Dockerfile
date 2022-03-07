FROM golang:1.15-alpine AS builder

ADD . /app
WORKDIR /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"'

FROM alpine:3.15

LABEL org.opencontainers.image.source https://github.com/yuya-takeyama/github-dependabot-prometheus

RUN apk --update add ca-certificates

COPY --from=builder /app/github-dependabot-prometheus /github-dependabot-prometheus

ENTRYPOINT ["/github-dependabot-prometheus"]
