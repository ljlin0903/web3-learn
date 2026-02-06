# ============================================
# Multi-stage Dockerfile for MEV Arbitrage Bot
# ============================================

# -------------------- Stage 1: Builder --------------------
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装编译依赖
RUN apk add --no-cache git gcc musl-dev

# 配置 Go 代理 (加速依赖下载)
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖 (利用 Docker 缓存层)
RUN go mod download

# 复制源代码
COPY . .

# 编译二进制文件
RUN go build -ldflags="-s -w" -o mev-bot cmd/bot/main.go

# -------------------- Stage 2: Runtime --------------------
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从 builder 复制二进制文件
COPY --from=builder /build/mev-bot .

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 健康检查 (可选)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep mev-bot || exit 1

# 暴露端口 (如果需要 HTTP API)
# EXPOSE 8080

# 启动命令
CMD ["./mev-bot"]
