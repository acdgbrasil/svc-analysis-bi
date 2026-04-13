FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/svc-analysis-bi ./cmd/server/

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/acdgbrasil/svc-analysis-bi"
LABEL org.opencontainers.image.description="Descriptive analytics service for ACDG Brasil"
LABEL org.opencontainers.image.licenses="UNLICENSED"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/svc-analysis-bi /bin/svc-analysis-bi

EXPOSE 8080

ENTRYPOINT ["/bin/svc-analysis-bi"]
