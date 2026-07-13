# provider-stub — real-protocol HTTP double serving the three provider-side
# shapes compat-check's component tests need (CS6 unreachable/500, CS7 delayed
# past settings.timeout, CS8 malformed OpenAPI). Real HTTP server, real
# protocol — never an in-code mock. Build context is the repo root (mirrors
# tool.Dockerfile) so this stub can later share code with the main module if
# needed; today it is a self-contained stdlib-only Go program.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY component-tests/compose/provider-stub/go.mod ./
COPY component-tests/compose/provider-stub/main.go ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags '-s -w' -o /out/provider-stub .

FROM alpine:3.21
COPY --from=build /out/provider-stub /provider-stub
EXPOSE 8080
ENTRYPOINT ["/provider-stub"]
