FROM golang:alpine as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
ENV CGO_ENABLED=0
RUN go build \
	cmd/moneropayd/moneropayd.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder /build/moneropayd .
CMD ["./moneropayd", "-bind=:5000"]
