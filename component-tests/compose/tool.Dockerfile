# Сборка статик-бинаря SUT (template-go-cli) + one-shot стейджинг в общий том.
# Контекст сборки — корень репо. CGO=0 → полностью статический бинарь.
# Контейнер-«tool» запускается ОДИН РАЗ: копирует бинарь в /out (том) и выходит 0.
# tester ждёт его успешного завершения и запускает бинарь из тома.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags '-s -w' -o /out/tool ./cmd/app

FROM alpine:3.21
COPY --from=build /out/tool /tool
COPY config.yaml /config.yaml
# One-shot стейджер (эфемерный, не SUT-рантайм): под root, т.к. named-том /out
# создаётся root-owned на первом монтировании. Копирует бинарь+конфиг в том и выходит.
ENTRYPOINT ["/bin/sh", "-c", "cp /tool /out/tool && cp /config.yaml /out/config.yaml && echo 'staged tool into shared volume /out'"]
