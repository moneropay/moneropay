FROM golang:alpine as builder
WORKDIR /build
COPY . .
ENV CGO_ENABLED=0
RUN go build

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder /build/moneropay .
COPY --from=builder /build/db db
CMD ["./moneropay", "-bind=:5000"]
