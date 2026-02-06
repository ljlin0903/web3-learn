# Docker 配置文件

> 所有 Docker 相关配置文件

---

## 📁 文件说明

### Dockerfile
**多阶段构建配置**
- **Stage 1 (Builder)**: 编译 Go 程序
- **Stage 2 (Runtime)**: 运行环境（Alpine Linux）
- **镜像大小**: ~50MB

**特点**:
- ✅ 多阶段构建，镜像小
- ✅ 使用 Alpine Linux
- ✅ 配置 GOPROXY 加速
- ✅ 非 root 用户运行
- ✅ 健康检查

### docker-compose.yml
**容器编排配置**
- **服务名**: mev-bot
- **重启策略**: unless-stopped
- **资源限制**: 512M 内存，1 CPU
- **日志管理**: 自动轮转

**特点**:
- ✅ 环境变量管理
- ✅ 日志持久化
- ✅ 资源限制
- ✅ 自动重启

---

## 🚀 快速使用

### 本地运行

```bash
# 在项目根目录执行
cd /Users/ljlin/web3/mev-arbitrage-bot

# 构建并启动
docker-compose -f docker/docker-compose.yml up -d

# 查看日志
docker-compose -f docker/docker-compose.yml logs -f

# 停止
docker-compose -f docker/docker-compose.yml down
```

### 服务器部署

```bash
# 使用自动部署脚本
./scripts/deploy-docker.sh
```

---

## 📝 配置说明

### 环境变量
所有环境变量通过 `../.env` 文件配置。

必需的环境变量:
- `RPC_HTTPS_URL` - RPC 节点 URL
- `RPC_WSS_URL` - WebSocket URL
- `PRIVATE_KEY` - 钱包私钥
- `PUBLIC_ADDRESS` - 钱包地址

### 资源配置

当前配置:
```yaml
limits:
  cpus: '1.0'      # 最大 1 核
  memory: 512M     # 最大 512M 内存
reservations:
  cpus: '0.5'      # 保留 0.5 核
  memory: 256M     # 保留 256M 内存
```

可根据实际情况调整。

### 日志配置

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"   # 单文件最大 10M
    max-file: "3"      # 最多保留 3 个文件
```

---

## 🔧 常用命令

### 构建镜像
```bash
docker-compose -f docker/docker-compose.yml build
```

### 启动容器
```bash
docker-compose -f docker/docker-compose.yml up -d
```

### 查看状态
```bash
docker-compose -f docker/docker-compose.yml ps
```

### 查看日志
```bash
# 实时日志
docker-compose -f docker/docker-compose.yml logs -f

# 最近 50 行
docker-compose -f docker/docker-compose.yml logs --tail=50

# 错误日志
docker-compose -f docker/docker-compose.yml logs | grep ERROR
```

### 重启服务
```bash
docker-compose -f docker/docker-compose.yml restart
```

### 停止服务
```bash
docker-compose -f docker/docker-compose.yml down
```

### 完全清理
```bash
# 停止并删除容器、网络、卷
docker-compose -f docker/docker-compose.yml down -v

# 删除镜像
docker rmi mev-arbitrage-bot:latest
```

---

## 🐛 故障排查

### 问题 1: 构建失败
```bash
# 清理缓存重新构建
docker-compose -f docker/docker-compose.yml build --no-cache
```

### 问题 2: 容器启动失败
```bash
# 查看详细日志
docker-compose -f docker/docker-compose.yml logs

# 检查配置
docker-compose -f docker/docker-compose.yml config
```

### 问题 3: 资源不足
编辑 `docker-compose.yml`:
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'      # 增加 CPU
      memory: 1024M    # 增加内存
```

---

## 📊 镜像信息

### 构建信息
- **基础镜像**: golang:1.21-alpine (Builder)
- **运行镜像**: alpine:3.18
- **最终大小**: ~50MB
- **构建时间**: ~2 分钟

### 安全特性
- ✅ 非 root 用户运行
- ✅ 最小化镜像
- ✅ 无不必要的工具
- ✅ 定期更新依赖

---

## 📚 相关文档

- [部署指南](../docs/DEPLOYMENT.md) - 详细的 Docker 部署说明
- [服务器运行指南](../docs/RUN_ON_SERVER.md) - 服务器部署步骤
- [项目 README](../README.md) - 项目总览

---

**注意**: 
- 所有命令都需要在项目根目录执行
- 使用 `-f docker/docker-compose.yml` 指定配置文件
- 确保 `.env` 文件配置正确
