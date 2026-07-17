# Real-protocol HTTP stub for the provider's spec_url (scenarios 5 HTTP_ERROR / 6 TIMEOUT_ERROR).
# A genuine long-running HTTP server (never an in-code mock), reached by the SUT over the
# real protocol via the compose service name "provider-stub". Standalone module — no
# dependency on the SUT's go.mod.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY component-tests/stub/go.mod ./
COPY component-tests/stub/main.go ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags '-s -w' -o /out/provider-stub .

FROM alpine:3.21
COPY --from=build /out/provider-stub /provider-stub
EXPOSE 8080
ENTRYPOINT ["/provider-stub"]
