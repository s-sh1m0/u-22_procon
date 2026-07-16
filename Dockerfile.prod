# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.26 AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 3: Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y git ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend-builder /app/server .
# go/packages がリポジトリをクローン後にビルドするため Go SDK が必要
COPY --from=backend-builder /usr/local/go /usr/local/go
COPY --from=frontend-builder /app/dist ./frontend
ENV PATH="/usr/local/go/bin:${PATH}"
ENV STATIC_DIR=/app/frontend
EXPOSE 8080
CMD ["./server"]
