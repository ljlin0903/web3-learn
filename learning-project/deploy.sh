#!/bin/bash
# Web3 Quant Bot Deployment Script
# Web3 量化交易机器人部署脚本

set -e

echo "========================================="
echo "🚀 Web3 Quant Bot 部署脚本"
echo "========================================="

# Variables / 变量
SERVER_USER="root"  # 修改为你的服务器用户名
SERVER_IP="47.93.253.224"        # 阿里云服务器公网 IP
SERVER_PASS="Root159647"  # 服务器密码
DEPLOY_PATH="/opt/web3-quant"

# SSH command with password / 带密码的 SSH 命令
SSH_CMD="sshpass -p '$SERVER_PASS' ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP"
RSYNC_CMD="sshpass -p '$SERVER_PASS' rsync -avz --progress -e 'ssh -o StrictHostKeyChecking=no'"

# Check if server IP is set / 检查服务器 IP 是否设置
if [ -z "$SERVER_IP" ]; then
    echo "❌ 错误：请在脚本中设置 SERVER_IP 变量"
    exit 1
fi

echo "📋 部署配置："
echo "   服务器: $SERVER_USER@$SERVER_IP"
echo "   路径: $DEPLOY_PATH"
echo ""

# Step 1: Install Docker on server (if not installed)
# 步骤1：在服务器上安装 Docker（如果未安装）
echo "📦 步骤 1/5: 检查 Docker 安装..."
$SSH_CMD << 'ENDSSH'
if ! command -v docker &> /dev/null; then
    echo "   正在安装 Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl start docker
    systemctl enable docker
    echo "   ✓ Docker 安装完成"
else
    echo "   ✓ Docker 已安装"
fi

if ! command -v docker-compose &> /dev/null; then
    echo "   正在安装 Docker Compose..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    echo "   ✓ Docker Compose 安装完成"
else
    echo "   ✓ Docker Compose 已安装"
fi
ENDSSH

# Step 2: Create deployment directory
# 步骤2：创建部署目录
echo ""
echo "📁 步骤 2/5: 创建部署目录..."
$SSH_CMD "mkdir -p $DEPLOY_PATH"
echo "   ✓ 目录创建完成"

# Step 3: Upload files to server
# 步骤3：上传文件到服务器
echo ""
echo "📤 步骤 3/5: 上传文件到服务器..."
$RSYNC_CMD \
    --exclude 'contracts/' \
    --exclude '.git/' \
    --exclude 'test_program' \
    --exclude '*.md' \
    ./ $SERVER_USER@$SERVER_IP:$DEPLOY_PATH/
echo "   ✓ 文件上传完成"

# Step 4: Build and start Docker container
# 步骤4：构建并启动 Docker 容器
echo ""
echo "🔨 步骤 4/5: 构建 Docker 镜像..."
$SSH_CMD << ENDSSH
cd $DEPLOY_PATH
docker-compose down 2>/dev/null || true
docker-compose build
echo "   ✓ 镜像构建完成"
ENDSSH

# Step 5: Start services
# 步骤5：启动服务
echo ""
echo "🚀 步骤 5/5: 启动服务..."
$SSH_CMD << ENDSSH
cd $DEPLOY_PATH
docker-compose up -d
echo "   ✓ 服务启动完成"
ENDSSH

echo ""
echo "========================================="
echo "✅ 部署成功！"
echo "========================================="
echo ""
echo "📊 查看日志："
echo "   $SSH_CMD 'cd $DEPLOY_PATH && docker-compose logs -f'"
echo ""
echo "🔍 查看状态："
echo "   $SSH_CMD 'cd $DEPLOY_PATH && docker-compose ps'"
echo ""
echo "🛑 停止服务："
echo "   $SSH_CMD 'cd $DEPLOY_PATH && docker-compose down'"
echo ""
