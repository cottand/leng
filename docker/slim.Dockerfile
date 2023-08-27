FROM alpine:3.17 as certs
RUN apk --update add ca-certificates

FROM golang:1.21.0-alpine3.17 AS builder

RUN apk add git bash gcc musl-dev upx git
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod tidy
#RUN go test -v ./...
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go build -ldflags "-w -s" -v -o main
RUN upx -9 -o main.minify main && mv main.minify main

FROM  alpine:3.17
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/main /usr/bin/main

ENTRYPOINT ["/usr/bin/main"]