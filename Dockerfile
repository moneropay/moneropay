FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.17 as build

RUN apk add --no-cache \
    gcc \
    musl-dev

WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=1 go build -o /out/moneropay -ldflags '-s -w -extldflags "-static"'
COPY db /out/db

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=build /out .
ENTRYPOINT ["./moneropay", "-bind=:5000"]
