# ==============================================================================
# Production Runtime for Poltekkes CAT Backend API
# ==============================================================================
FROM alpine:3.21

WORKDIR /app

# Install runtime dependencies (tzdata for Asia/Jakarta timezone, ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates tzdata curl || true
ENV TZ=Asia/Jakarta

# Copy pre-compiled binary
COPY poltekkes-cat-api /app/poltekkes-cat-api
RUN chmod +x /app/poltekkes-cat-api

# Expose backend port
EXPOSE 5000

# Default environment configuration
ENV PORT=5000 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASS=postgres \
    DB_NAME=cat \
    DB_SSLMODE=disable

# Healthcheck
HEALTHCHECK --interval=15s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:5000/api/v1/health || exit 1

ENTRYPOINT ["/app/poltekkes-cat-api"]

