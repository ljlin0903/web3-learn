#!/bin/bash
# Web3 Quant Bot Deployment Script (Expect version)
# Web3 量化交易机器人部署脚本 (Expect 版本)

set -e

echo "========================================="
echo "🚀 Web3 Quant Bot 部署脚本"
echo "========================================="

# Variables / 变量
SERVER_USER="root"
SERVER_IP="47.93.253.224"
SERVER_PASS="Root159647"
DEPLOY_PATH="/opt/web3-quant"

echo "📋 部署配置："
echo "   服务器: $SERVER_USER@$SERVER_IP"
echo "   路径: $DEPLOY_PATH"
echo ""

# Step 1: Check Docker installation
# 步骤1：检查 Docker 安装
echo "📦 步骤 1/4: 检查服务器环境..."
expect << EOF
spawn ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "docker --version && docker-compose --version || echo 'Need install'"
expect "password:"
send "$SERVER_PASS\r"
expect eof
EOF
echo "   ✓ 环境检查完成"

# Step 2: Create deployment directory
# 步骤2：创建部署目录
echo ""
echo "📁 步骤 2/4: 创建部署目录..."
expect << EOF
spawn ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "mkdir -p $DEPLOY_PATH"
expect "password:"
send "$SERVER_PASS\r"
expect eof
EOF
echo "   ✓ 目录创建完成"

# Step 3: Upload files to server
# 步骤3：上传文件到服务器
echo ""
echo "📤 步骤 3/4: 上传文件到服务器..."
expect << EOF
set timeout 300
spawn rsync -avz --progress \
    --exclude 'contracts/' \
    --exclude '.git/' \
    --exclude 'test_program' \
    --exclude '*.md' \
    -e "ssh -o StrictHostKeyChecking=no" \
    ./ $SERVER_USER@$SERVER_IP:$DEPLOY_PATH/
expect "password:"
send "$SERVER_PASS\r"
expect eof
EOF
echo "   ✓ 文件上传完成"

# Step 4: Build and start
# 步骤4：构建并启动
echo ""
echo "🔨 步骤 4/4: 构建并启动服务..."
expect << EOF
set timeout 600
spawn ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "cd $DEPLOY_PATH && docker-compose down 2>/dev/null || true && docker-compose build && docker-compose up -d"
expect "password:"
send "$SERVER_PASS\r"
expect eof
EOF

echo ""
echo "========================================="
echo "✅ 部署成功！"
echo "========================================="
echo ""
echo "📊 查看日志："
echo "   ssh root@$SERVER_IP"
echo "   cd $DEPLOY_PATH && docker-compose logs -f"
echo ""
