FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies required for CGO and secp256k1
RUN apk add --no-cache build-base linux-headers git

# Copy the Go module files
COPY application/backend/go.mod application/backend/go.sum ./
RUN go mod download

# Copy the source code
COPY application/backend/ ./

# Build the binaries with CGO enabled but with optimizations
RUN go build -ldflags="-s -w" -o payment-processor ./cmd/payment-processor/main.go
RUN go build -ldflags="-s -w" -o blockchain-processor ./cmd/blockchain-processor/main.go

# Build admin app dependencies (and optionally bundle) in a separate stage
FROM node:22-slim AS admin-builder

WORKDIR /code/

COPY application/frontend/admin-app/package.json application/frontend/admin-app/package-lock.json /code/
RUN npm ci
COPY application/frontend/admin-app/ /code/
RUN npx adminforth bundle

# Final runtime stage - use postgres alpine directly with supervisor and nodejs
FROM postgres:17.4-alpine3.21

# Install additional packages
RUN apk add --no-cache supervisor curl tzdata libc6-compat nodejs npm

# Copy the binaries from the builder stage
COPY --from=builder /app/payment-processor /app/payment-processor
COPY --from=builder /app/blockchain-processor /app/blockchain-processor
COPY --from=builder /app/static /static

# Copy the admin app (already installed via npm ci in admin-builder)
COPY --from=admin-builder /code/ /code/

# Copy the migrations
COPY application/backend/migrations /app/migrations

# Copy runtime scripts and configs
COPY application/deploy/run.sh /run.sh
COPY application/deploy/init-db.sh /init-db.sh
COPY application/deploy/supervisord.conf /etc/supervisor/supervisord.conf
COPY application/deploy/start.sh /start.sh
COPY application/deploy/admin-start.sh /admin-start.sh
RUN chmod +x /run.sh /init-db.sh /start.sh /admin-start.sh

# Copy ijson from the external image
COPY --from=lega911/ijson /ijson /ijson
RUN chmod +x /ijson

# Add volume for PostgreSQL data persistence
VOLUME ["/var/lib/postgresql/data"]

# Install migrate CLI
RUN apk add --no-cache go git && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3 && \
    mv /root/go/bin/migrate /usr/local/bin/migrate && \
    apk del go git && \
    rm -rf /root/go

ENV WORKER_COUNT=1

# Fiatless API
EXPOSE 8080

# Admin (frontend) app
EXPOSE 3500

CMD ["/bin/sh","/run.sh"]