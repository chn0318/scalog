# ========= Stage 1: Build =========
FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -o /out/scalog .

RUN GOBIN=/out CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go install github.com/mattn/goreman@v0.3.12


# ========= Stage 2: Runtime =========
FROM debian:bookworm-slim AS runner

WORKDIR /root

COPY Procfile ./Procfile
COPY .scalog.yaml ./.scalog.yaml


COPY --from=builder /out/scalog /usr/local/bin/scalog
COPY --from=builder /out/goreman /usr/local/bin/goreman

ENTRYPOINT ["scalog"]

CMD ["--help"]


