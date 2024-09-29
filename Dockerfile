FROM --platform=$BUILDPLATFORM techknowlogick/xgo:go-1.23.1 AS build

ADD . /go/src
WORKDIR /go/src
ARG TARGETOS TARGETARCH
RUN --mount=target=. --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg xgo -x --targets=$TARGETOS/$TARGETARCH -ldflags '-s -w -extldflags "-static"' -out moneropay cmd/moneropay
COPY db /out/db
RUN mv /build/moneropay-* /out/moneropay

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=build /out .
ENTRYPOINT ["./moneropay", "-bind=:5000"]
