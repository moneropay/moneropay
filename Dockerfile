FROM golang:1.18-alpine3.15 as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w"

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder /build/moneropay .
COPY --from=builder /build/db db
ENTRYPOINT ["./moneropay", "-bind=:5000"]
