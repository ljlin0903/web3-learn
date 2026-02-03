# Web3 Quant Bot - Dockerfile
# Web3 量化交易机器人 - Docker 镜像

# Stage 1: Build Go binaries
# 阶段1：编译 Go 二进制文件
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Install dependencies / 安装依赖
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files / 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# Copy source code / 复制源代码
COPY *.go ./

# Build binaries / 编译二进制文件
RUN CGO_ENABLED=1 GOOS=linux go build -o /monitor_reconnect monitor_reconnect.go

# Stage 2: Runtime image
# 阶段2：运行时镜像
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS / 安装 HTTPS 证书
RUN apk --no-cache add ca-certificates tzdata

# Set timezone / 设置时区
ENV TZ=Asia/Shanghai

# Copy binary from builder / 从构建阶段复制二进制文件
COPY --from=go-builder /monitor_reconnect .

# Copy .env file / 复制配置文件
COPY .env .

# Run the bot / 运行机器人
CMD ["./monitor_reconnect"]
