# ── Stage 1: Build frontend ──────────────────────────────────────────
FROM node:22-alpine AS frontend-builder

WORKDIR /build

COPY echo-news-fronted/package.json echo-news-fronted/package-lock.json ./
RUN npm ci

COPY echo-news-fronted/ ./
RUN npm run build-only

# ── Stage 2: Build backend ──────────────────────────────────────────
FROM golang:1.26-alpine AS backend-builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -o echo-news .

# ── Stage 3: Final runtime image ────────────────────────────────────
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /build/echo-news ./echo-news
COPY --from=frontend-builder /build/dist ./static

ENV PORT=8080
ENV DATABASE_URL=/app/data/echo-news.db
ENV STATIC_DIR=/app/static

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./echo-news"]
