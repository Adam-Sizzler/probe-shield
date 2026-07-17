# Stage 1: Build the frontend assets
FROM node:22-alpine AS frontend-builder
WORKDIR /src/frontend

# Copy package files and local vendor contracts
COPY frontend/package*.json ./
COPY frontend/vendor ./vendor
RUN --mount=type=cache,target=/root/.npm npm ci --legacy-peer-deps

# Copy the rest of frontend files and compile
COPY frontend/ ./
RUN npm run start:build

# Stage 2: Build the Go backend
FROM golang:alpine AS backend-builder
WORKDIR /src/backend

# Copy go mod files first to leverage Docker layer caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the source code and build the statically linked binary
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/probe-shield .

# Stage 3: Create the minimal production image
FROM alpine:3.21
WORKDIR /app

# Create a non-root user for security
RUN addgroup -S probe && adduser -S -G probe probe

# Copy the compiled binary from the builder stage
COPY --from=backend-builder /out/probe-shield /app/probe-shield

# Copy static frontend assets and configuration files from builders/host
COPY --from=frontend-builder /src/frontend/dist /app/frontend/dist
COPY configs /app/configs

EXPOSE 4000

# Drop privileges to the non-root user
USER probe

ENTRYPOINT ["/app/probe-shield"]
