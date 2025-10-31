# ---- build stage ----
FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0
ENV GOCACHE=/root/.cache/go-build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=${GOCACHE} \
    go build -ldflags=${LDFLAGS} -v -a -o server main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server /app/server

CMD /app/server
