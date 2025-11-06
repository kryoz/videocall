FROM golang:1.25-alpine AS builder
ENV CGO_ENABLED=0
ENV GOCACHE=/root/.cache/go-build

WORKDIR /internal
RUN set -ex

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=${GOCACHE} \
    go build -ldflags="-s" -v -a -o server main.go

#   go build -gcflags "all=-N -l" -o server main.go

# debugger
#RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=builder /internal/server /app/server
# debugger
#COPY --from=builder /go/bin/dlv /

CMD /app/server
