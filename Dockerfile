FROM node:18 as fe-builder
WORKDIR /app
COPY . .
RUN cd /app/web && npm install -g pnpm && pnpm install && npm run build

FROM golang:1.20 AS builder
WORKDIR /app
COPY . .
COPY --from=fe-builder /app/static/dist ./static/dist
RUN make build

FROM debian:12.1-slim
RUN apt-get update \
    && apt-get install -y curl \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/keepshare .
CMD ["./keepshare", "start"]
