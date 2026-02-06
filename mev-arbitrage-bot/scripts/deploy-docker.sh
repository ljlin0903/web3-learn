#!/bin/bash

# ============================================
# Docker 部署脚本 - 自动化部署到服务器
# ============================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
SERVER_IP="47.93.253.224"
SERVER_USER="root"
SERVER_PASSWORD="Root159647"
DEPLOY_DIR="/opt/mev-arbitrage-bot"
PROJECT_NAME="mev-arbitrage-bot"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  MEV Arbitrage Bot - Docker 部署${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查 sshpass
if ! command -v sshpass &> /dev/null; then
    echo -e "${YELLOW}⚠️  未检测到 sshpass，正在安装...${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install hudochenkov/sshpass/sshpass
    else
        sudo apt-get install -y sshpass
    fi
fi

# 步骤 1: 打包项目
echo -e "${GREEN}[1/6]${NC} 📦 打包项目文件..."
tar --exclude='.git' \
    --exclude='bin' \
    --exclude='vendor' \
    --exclude='*.log' \
    --exclude='.DS_Store' \
    -czf /tmp/${PROJECT_NAME}.tar.gz .

echo -e "${GREEN}✓${NC} 项目打包完成"
echo ""

# 步骤 2: 上传到服务器
echo -e "${GREEN}[2/6]${NC} 📤 上传到服务器..."
sshpass -p "${SERVER_PASSWORD}" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} \
    "mkdir -p ${DEPLOY_DIR}"

sshpass -p "${SERVER_PASSWORD}" scp -o StrictHostKeyChecking=no \
    /tmp/${PROJECT_NAME}.tar.gz ${SERVER_USER}@${SERVER_IP}:${DEPLOY_DIR}/

echo -e "${GREEN}✓${NC} 上传完成"
echo ""

# 步骤 3: 解压项目
echo -e "${GREEN}[3/6]${NC} 📂 解压项目文件..."
sshpass -p "${SERVER_PASSWORD}" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mev-arbitrage-bot
tar -xzf mev-arbitrage-bot.tar.gz
rm -f mev-arbitrage-bot.tar.gz
echo "✓ 解压完成"
EOF

echo ""

# 步骤 4: 检查 Docker 环境
echo -e "${GREEN}[4/6]${NC} 🐳 检查 Docker 环境..."
sshpass -p "${SERVER_PASSWORD}" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'EOF'
# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker 未安装，正在安装..."
    curl -fsSL https://get.docker.com | bash
    systemctl start docker
    systemctl enable docker
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "⚠️  Docker Compose 未安装，正在安装..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

echo "✓ Docker 环境就绪"
docker --version
docker-compose --version
EOF

echo ""

# 步骤 5: 构建并启动容器
echo -e "${GREEN}[5/6]${NC} 🚀 构建并启动容器..."
sshpass -p "${SERVER_PASSWORD}" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mev-arbitrage-bot

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "⚠️  未找到 .env 文件，请先配置环境变量"
    echo "   示例: cp .env.example .env && vim .env"
    exit 1
fi

# 停止旧容器
echo "停止旧容器..."
docker-compose -f docker/docker-compose.yml down 2>/dev/null || true

# 构建镜像
echo "构建 Docker 镜像..."
docker-compose -f docker/docker-compose.yml build --no-cache

# 启动容器
echo "启动容器..."
docker-compose -f docker/docker-compose.yml up -d

echo "✓ 容器已启动"
EOF

echo ""

# 步骤 6: 验证部署
echo -e "${GREEN}[6/6]${NC} ✅ 验证部署状态..."
sshpass -p "${SERVER_PASSWORD}" ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mev-arbitrage-bot

# 等待容器启动
sleep 3

# 检查容器状态
echo ""
echo "========================================"
echo "容器状态:"
docker-compose -f docker/docker-compose.yml ps

echo ""
echo "========================================"
echo "最近日志:"
docker-compose -f docker/docker-compose.yml logs --tail=20

echo ""
echo "========================================"
EOF

# 清理临时文件
rm -f /tmp/${PROJECT_NAME}.tar.gz

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ 部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}管理命令:${NC}"
echo -e "  查看日志: ${GREEN}ssh root@${SERVER_IP} 'cd ${DEPLOY_DIR} && docker-compose -f docker/docker-compose.yml logs -f'${NC}"
echo -e "  停止服务: ${GREEN}ssh root@${SERVER_IP} 'cd ${DEPLOY_DIR} && docker-compose -f docker/docker-compose.yml down'${NC}"
echo -e "  重启服务: ${GREEN}ssh root@${SERVER_IP} 'cd ${DEPLOY_DIR} && docker-compose -f docker/docker-compose.yml restart'${NC}"
echo -e "  查看状态: ${GREEN}ssh root@${SERVER_IP} 'cd ${DEPLOY_DIR} && docker-compose -f docker/docker-compose.yml ps'${NC}"
echo ""
