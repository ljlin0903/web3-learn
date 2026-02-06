# 🚀 服务器部署指南 - Docker 方式

## 📋 目录
1. [环境要求](#环境要求)
2. [快速部署](#快速部署)
3. [手动部署](#手动部署)
4. [管理命令](#管理命令)
5. [故障排查](#故障排查)

---

## 🔧 环境要求

### 服务器环境
**已配置服务器信息**:
- **IP**: `47.93.253.224`
- **用户**: `root`
- **密码**: `Root159647`
- **操作系统**: Linux (推荐 Ubuntu 20.04+)

### Docker 环境 (会自动安装)
- Docker 20.10+
- Docker Compose 2.0+

### 网络要求
- 能访问以太坊 RPC 节点 (Alchemy/Infura)
- 能访问 Docker Hub / 阿里云镜像仓库

---

## 🚀 快速部署 (推荐)

### 方式一: 一键自动部署

**前提条件**:
1. 本地安装 `sshpass` (脚本会自动检测并安装)
2. 已配置 `.env` 文件

**执行部署**:
```bash
cd /Users/ljlin/web3/mev-arbitrage-bot

# 运行部署脚本
./scripts/deploy-docker.sh
```

**脚本会自动完成**:
1. ✅ 打包项目文件
2. ✅ 上传到服务器
3. ✅ 检查/安装 Docker 环境
4. ✅ 构建 Docker 镜像
5. ✅ 启动容器
6. ✅ 验证运行状态

**预期输出**:
```
========================================
  MEV Arbitrage Bot - Docker 部署
========================================

[1/6] 📦 打包项目文件...
✓ 项目打包完成

[2/6] 📤 上传到服务器...
✓ 上传完成

[3/6] 📂 解压项目文件...
✓ 解压完成

[4/6] 🐳 检查 Docker 环境...
✓ Docker 环境就绪

[5/6] 🚀 构建并启动容器...
✓ 容器已启动

[6/6] ✅ 验证部署状态...
容器状态: Up

========================================
✅ 部署完成！
========================================
```

---

## 🔨 手动部署

### 步骤 1: 连接服务器

```bash
ssh root@47.93.253.224
# 密码: Root159647
```

### 步骤 2: 安装 Docker (如果未安装)

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 启动 Docker
systemctl start docker
systemctl enable docker

# 验证安装
docker --version

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证安装
docker-compose --version
```

### 步骤 3: 上传项目

**方式 A: 使用 scp**
```bash
# 在本地执行
cd /Users/ljlin/web3/mev-arbitrage-bot

# 打包项目
tar -czf mev-bot.tar.gz \
    --exclude='.git' \
    --exclude='bin' \
    --exclude='vendor' \
    .

# 上传到服务器
scp mev-bot.tar.gz root@47.93.253.224:/opt/

# 在服务器上解压
ssh root@47.93.253.224
cd /opt
tar -xzf mev-bot.tar.gz -C mev-arbitrage-bot
cd mev-arbitrage-bot
```

**方式 B: 使用 Git**
```bash
# 在服务器上执行
cd /opt
git clone https://github.com/ljlin/mev-arbitrage-bot.git
cd mev-arbitrage-bot
```

### 步骤 4: 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置
vim .env
```

**必填配置项**:
```bash
NETWORK=sepolia
RPC_HTTPS_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
RPC_WSS_URL=wss://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
PRIVATE_KEY=0x...
PUBLIC_ADDRESS=0x...
```

### 步骤 5: 构建并启动

```bash
# 构建镜像
docker-compose build

# 启动容器
docker-compose up -d

# 查看日志
docker-compose logs -f
```

---

## 📊 管理命令

### 基础操作

```bash
# 进入项目目录
cd /opt/mev-arbitrage-bot

# 查看容器状态
docker-compose ps

# 查看实时日志
docker-compose logs -f

# 查看最近 100 行日志
docker-compose logs --tail=100

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 启动服务
docker-compose up -d

# 重新构建并启动
docker-compose up -d --build
```

### 进入容器

```bash
# 进入运行中的容器
docker-compose exec mev-bot sh

# 或使用 docker 命令
docker exec -it mev-arbitrage-bot sh
```

### 资源监控

```bash
# 查看容器资源使用
docker stats mev-arbitrage-bot

# 查看容器详细信息
docker inspect mev-arbitrage-bot
```

### 日志管理

```bash
# 清理日志
docker-compose logs --no-color > logs_backup.txt
docker-compose down
docker-compose up -d

# 实时跟踪日志
docker-compose logs -f --tail=50
```

---

## 🔧 Docker 镜像说明

### 镜像特点

1. **多阶段构建**
   - Builder 阶段: 编译 Go 程序
   - Runtime 阶段: 运行程序
   - 最终镜像大小: ~50MB

2. **安全优化**
   - 使用非 root 用户运行
   - 最小依赖
   - 只包含必要文件

3. **性能优化**
   - 使用 Go 代理加速依赖下载
   - 利用 Docker 缓存层
   - 静态编译二进制文件

### 镜像构建

```bash
# 手动构建 (可选)
docker build -t mev-bot:latest .

# 指定构建参数
docker build \
    --build-arg GOPROXY=https://goproxy.cn \
    -t mev-bot:latest .

# 查看镜像
docker images | grep mev-bot
```

---

## 🐛 故障排查

### 问题 1: 容器无法启动

**检查步骤**:
```bash
# 查看容器状态
docker-compose ps

# 查看详细日志
docker-compose logs

# 检查 .env 文件
cat .env

# 验证配置
docker-compose config
```

**常见原因**:
- `.env` 文件未配置
- RPC URL 无效
- 端口冲突
- 内存不足

### 问题 2: 连接 RPC 失败

**检查网络**:
```bash
# 测试 RPC 连接
curl -X POST https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# 检查 DNS
nslookup eth-sepolia.g.alchemy.com

# 检查防火墙
iptables -L
```

### 问题 3: 内存不足

**增加限制**:
```bash
# 编辑 docker-compose.yml
vim docker-compose.yml

# 修改资源限制
deploy:
  resources:
    limits:
      memory: 1G  # 增加到 1GB
```

### 问题 4: 编译失败

**常见原因**:
- Go 版本不匹配
- 依赖下载失败
- 网络问题

**解决方案**:
```bash
# 使用国内镜像
docker build \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    -t mev-bot:latest .

# 清理缓存重新构建
docker-compose build --no-cache
```

### 问题 5: 查看详细错误

```bash
# 查看容器退出原因
docker-compose ps -a

# 查看完整日志
docker-compose logs --no-color > full_logs.txt

# 进入容器调试
docker-compose run --rm mev-bot sh
```

---

## 📈 性能监控

### 资源使用监控

```bash
# 实时监控
docker stats mev-arbitrage-bot

# 查看资源使用历史
docker inspect mev-arbitrage-bot | grep -A 20 "Resources"
```

### 日志分析

```bash
# 统计错误日志
docker-compose logs | grep ERROR | wc -l

# 查找特定关键词
docker-compose logs | grep "arbitrage"

# 导出日志
docker-compose logs --no-color > logs_$(date +%Y%m%d).txt
```

---

## 🔄 更新部署

### 方式 1: 重新运行部署脚本

```bash
cd /Users/ljlin/web3/mev-arbitrage-bot
./scripts/deploy-docker.sh
```

### 方式 2: 手动更新

```bash
# 连接服务器
ssh root@47.93.253.224

# 进入项目目录
cd /opt/mev-arbitrage-bot

# 拉取最新代码 (如果使用 Git)
git pull

# 重新构建并启动
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# 查看日志
docker-compose logs -f
```

---

## 🎯 最佳实践

### 1. 定期备份

```bash
# 备份 .env 文件
cp .env .env.backup.$(date +%Y%m%d)

# 备份日志
docker-compose logs --no-color > backup/logs_$(date +%Y%m%d).txt
```

### 2. 监控告警

```bash
# 设置 crontab 监控 (每5分钟检查一次)
crontab -e

# 添加:
*/5 * * * * docker ps | grep mev-arbitrage-bot || echo "Bot stopped!" | mail -s "Alert" your@email.com
```

### 3. 日志轮转

Docker Compose 已配置日志轮转:
- 单个日志文件最大 10MB
- 保留最近 3 个文件

### 4. 安全建议

- ✅ 定期更新 Docker
- ✅ 使用强密码
- ✅ 启用防火墙
- ✅ 定期备份私钥
- ✅ 监控异常活动

---

## 📞 获取帮助

### 快速命令参考

```bash
# 一键查看状态
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose ps && docker-compose logs --tail=20'

# 一键重启
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose restart'

# 一键停止
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose down'
```

### 文档链接

- [README.md](README.md) - 项目介绍
- [PROJECT_STATUS.md](PROJECT_STATUS.md) - 开发进度
- [QUICKSTART.md](QUICKSTART.md) - 快速开始

---

**版本**: v1.0.0  
**更新时间**: 2024-02-06  
**部署环境**: Docker + Docker Compose
