FROM golang:alpine as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
ENV CGO_ENABLED=0
RUN go build

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder /build/moneropay .
CMD ["./moneropay", "-bind=:5000"]
