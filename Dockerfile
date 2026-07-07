FROM golang:1.23-alpine AS backend-builder
WORKDIR /src/backend
COPY backend/go.mod ./
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/probe-shield .

FROM alpine:3.21
WORKDIR /app
RUN addgroup -S probe && adduser -S -G probe probe
COPY --from=backend-builder /out/probe-shield /app/probe-shield
COPY frontend/dist /app/frontend/dist
COPY configs /app/configs
ENV PROBE_SHIELD_CONFIG_FILE=/app/configs/probe-shield.json
ENV PROBE_SHIELD_STATIC_DIR=/app/frontend/dist
EXPOSE 3000
USER probe
ENTRYPOINT ["/app/probe-shield"]
