# syntax=docker/dockerfile:1

# ── Build Stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agent-server ./cmd/agent-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/qq ./cmd/qq

# ── Runtime Stage ─────────────────────────────────────────────────────────────
FROM node:22-alpine

WORKDIR /app

# Copy compiled binaries from builder
COPY --from=builder /app/agent-server ./agent-server
COPY --from=builder /app/qq ./qq

# Install bash (required by the bash tool) and verify node/npx
RUN apk add --no-cache bash && node --version && npx --version && bash --version

# Override node image's entrypoint so docker-compose `command` runs binaries directly
ENTRYPOINT []
