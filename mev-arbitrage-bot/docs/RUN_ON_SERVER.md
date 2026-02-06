# 🚀 服务器运行指南

> **完整的服务器部署和运行流程**

---

## 📋 前置准备

### 1. 获取必要信息

**你需要准备的**:
- ✅ Alchemy API Key (注册: https://www.alchemy.com/)
- ✅ MetaMask 钱包地址和私钥
- ✅ 测试网 ETH (从水龙头获取)

###  2. 注册 Alchemy

```
1. 访问 https://www.alchemy.com/
2. 点击 "Sign Up" 注册账号
3. 创建新应用 (Create App)
   - Name: MEV Bot
   - Chain: Ethereum
   - Network: Sepolia (测试网)
4. 获取 API Key
   - 点击 "View Key"
   - 复制 HTTPS URL 和 WebSocket URL
```

**示例**:
```
HTTPS: https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
WSS:   wss://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
```

### 3. 准备钱包

**安装 MetaMask**:
```
1. 访问 https://metamask.io/
2. 安装浏览器插件
3. 创建新钱包（记住助记词！）
4. 切换到 Sepolia 测试网
```

**获取私钥**:
```
MetaMask → 右上角三个点 → 账户详情 → 导出私钥
⚠️  警告: 私钥=银行密码，绝对不能泄露！
```

**获取测试 ETH**:
```
访问水龙头:
- https://sepoliafaucet.com/
- https://www.alchemy.com/faucets/ethereum-sepolia

输入你的钱包地址，等待 1-2 分钟
```

---

## 🔧 服务器配置

### 方案 A: 一键自动部署 (推荐)

**步骤 1**: 在本地配置 `.env` 文件

```bash
cd /Users/ljlin/web3/mev-arbitrage-bot

# 编辑 .env 文件
vim .env
```

**必填配置** (示例):
```bash
# RPC 端点 - 替换成你的 API Key
RPC_HTTPS_URL=https://eth-sepolia.g.alchemy.com/v2/abc123def456
RPC_WSS_URL=wss://eth-sepolia.g.alchemy.com/v2/abc123def456

# 钱包配置 - 替换成你的私钥和地址
PRIVATE_KEY=36e7ea7469614617f8dd9e696ca2e7a44e1c84e753a2362e89d1cd0904cbc61a
PUBLIC_ADDRESS=0x57e6159fa93942F0a0a7A075EF8D0eb27cC60560

# 其他保持默认即可
DRY_RUN=true  # 模拟模式，不发送真实交易
LOG_LEVEL=info
```

**步骤 2**: 执行部署脚本

```bash
# 运行自动部署
./scripts/deploy-docker.sh
```

**自动完成**:
- ✅ 打包项目
- ✅ 上传到服务器 (47.93.253.224)
- ✅ 安装 Docker 环境
- ✅ 构建镜像
- ✅ 启动容器
- ✅ 验证运行

**预计耗时**: 5-10 分钟

---

### 方案 B: 手动部署

**步骤 1**: 连接服务器

```bash
ssh root@47.93.253.224
# 密码: Root159647
```

**步骤 2**: 准备项目目录

```bash
# 创建目录
mkdir -p /opt/mev-arbitrage-bot
cd /opt/mev-arbitrage-bot
```

**步骤 3**: 上传项目文件

**在本地执行**:
```bash
cd /Users/ljlin/web3/mev-arbitrage-bot

# 打包项目
tar -czf mev-bot.tar.gz \
    --exclude='.git' \
    --exclude='bin' \
    --exclude='vendor' \
    .

# 上传到服务器
scp mev-bot.tar.gz root@47.93.253.224:/opt/mev-arbitrage-bot/
```

**在服务器上执行**:
```bash
cd /opt/mev-arbitrage-bot
tar -xzf mev-bot.tar.gz
rm mev-bot.tar.gz
```

**步骤 4**: 配置 .env 文件

```bash
cp .env.example .env
vim .env
```

填入你的配置 (参考上面的必填配置)

**步骤 5**: 安装 Docker

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 启动 Docker
systemctl start docker
systemctl enable docker

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

**步骤 6**: 启动容器

```bash
cd /opt/mev-arbitrage-bot

# 构建镜像
docker-compose build

# 启动容器
docker-compose up -d

# 查看日志
docker-compose logs -f
```

---

## 📊 查看运行状态

### 查看日志

```bash
# 方式 1: 通过 SSH
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose logs -f'

# 方式 2: 登录服务器后
cd /opt/mev-arbitrage-bot
docker-compose logs -f

# 查看最近 50 行
docker-compose logs --tail=50

# 只看错误
docker-compose logs | grep ERROR
```

### 查看容器状态

```bash
cd /opt/mev-arbitrage-bot

# 查看容器状态
docker-compose ps

# 查看资源使用
docker stats mev-arbitrage-bot
```

### 预期日志输出

**启动成功**:
```
╔══════════════════════════════════════════════════════════╗
║         🤖  MEV ARBITRAGE BOT  🤖                        ║
╚══════════════════════════════════════════════════════════╝

INFO Loading configuration...
INFO ✅ Config loaded
INFO Initializing blockchain client...
INFO Connected to network with Chain ID: 11155111
INFO Latest block number: 5234567
INFO Account balance: 0.500000 ETH
INFO Current gas price: 25 Gwei
INFO ✅ All systems operational!
WARN 🧪 Running in DRY RUN mode
INFO 🚀 Arbitrage bot is running...
INFO    Monitoring DEX pools for arbitrage opportunities
```

**正常运行**:
```
INFO Starting arbitrage detection loop...
DEBUG Searching for arbitrage opportunities...
DEBUG Updated 3 pools
DEBUG No profitable arbitrage opportunities found
```

**发现机会**:
```
INFO 🎯 Found opportunity! Profit: 0.045000 ETH (4.50%)
INFO Path: WETH → USDC → DAI → WETH
WARN 🧪 DRY RUN MODE - Transaction not sent
INFO ========================================
INFO Arbitrage Opportunity Details
INFO ID: 7f8a9b2c
INFO Profit: 0.045000 ETH (4.50%)
INFO Net Profit: 0.035000 ETH (3.50% after gas)
```

---

## 🔄 管理命令

### 基础操作

```bash
cd /opt/mev-arbitrage-bot

# 启动
docker-compose up -d

# 停止
docker-compose down

# 重启
docker-compose restart

# 查看日志
docker-compose logs -f

# 进入容器
docker-compose exec mev-bot sh
```

### 更新代码

```bash
# 停止容器
docker-compose down

# 更新代码 (从本地上传新版本)
# 或使用 git pull (如果使用 Git)

# 重新构建并启动
docker-compose build --no-cache
docker-compose up -d
```

### 清理数据

```bash
# 停止并删除容器
docker-compose down

# 删除镜像
docker rmi mev-arbitrage-bot_mev-bot

# 清理未使用的镜像
docker system prune -a
```

---

## ⚙️ 配置调整

### 修改配置

```bash
cd /opt/mev-arbitrage-bot

# 编辑配置
vim .env

# 重启生效
docker-compose restart
```

### 常用配置调整

**提高利润要求**:
```bash
MIN_PROFIT_BPS=100  # 改成 1%
```

**调整日志级别**:
```bash
LOG_LEVEL=debug  # 查看详细日志
LOG_LEVEL=warn   # 只看警告和错误
```

**切换到真实模式** (⚠️ 谨慎！):
```bash
DRY_RUN=false  # 会发送真实交易
```

---

## 🐛 故障排查

### 问题 1: 连接 RPC 失败

**错误**:
```
ERROR Failed to initialize blockchain client: dial tcp: lookup failed
```

**解决方案**:
1. 检查 API Key 是否正确
2. 测试 RPC 连接:
```bash
curl -X POST https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```
3. 检查服务器网络

### 问题 2: 余额为 0

**错误**:
```
WARN ⚠️  Account has zero balance!
```

**解决方案**:
1. 访问水龙头获取测试 ETH
2. 检查钱包地址是否正确
3. 确认网络是 Sepolia

### 问题 3: 找不到池子

**错误**:
```
ERROR Failed to add monitored pools: no pools available
```

**原因**: 测试网可能没有 Uniswap V2 池子

**解决方案**:
1. 使用主网进行测试 (改 `NETWORK=mainnet`)
2. 或者等待智能合约部署后使用模拟池子

### 问题 4: 容器无法启动

**检查步骤**:
```bash
# 查看容器状态
docker-compose ps

# 查看详细日志
docker-compose logs

# 检查配置文件
docker-compose config
```

### 问题 5: Docker 编译失败

**解决方案**:
```bash
# 使用国内镜像
docker build \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    -t mev-bot:latest .

# 清理缓存重建
docker-compose build --no-cache
```

---

## 📈 监控和优化

### 性能监控

```bash
# 实时监控资源
docker stats mev-arbitrage-bot

# 查看容器详情
docker inspect mev-arbitrage-bot | grep -A 10 "Resources"
```

### 日志分析

```bash
# 统计运行时间
docker logs mev-arbitrage-bot 2>&1 | grep "operational" | head -1

# 统计发现的机会
docker logs mev-arbitrage-bot 2>&1 | grep "Found opportunity" | wc -l

# 导出日志
docker logs mev-arbitrage-bot > logs_$(date +%Y%m%d).txt
```

### 优化建议

**内存优化**:
```yaml
# docker-compose.yml
resources:
  limits:
    memory: 256M  # 降低限制
```

**网络优化**:
```bash
# 使用更快的 RPC 节点
# 或购买 Alchemy Growth 计划
```

---

## 🔒 安全提示

### 重要提醒

⚠️  **永远不要**:
- 泄露私钥
- 在公开场合展示 .env 文件
- 使用主网地址在测试网
- 在未充分测试前切换到真实模式

✅ **最佳实践**:
- 定期备份 .env 文件
- 使用独立的测试钱包
- 小金额开始测试
- 监控异常活动
- 定期更新依赖

### 备份配置

```bash
# 备份 .env
cp .env .env.backup.$(date +%Y%m%d)

# 备份到本地
scp root@47.93.253.224:/opt/mev-arbitrage-bot/.env ~/backups/
```

---

## 📞 获取帮助

### 快速诊断

```bash
# 一键检查状态
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && \
  echo "=== 容器状态 ===" && docker-compose ps && \
  echo "" && echo "=== 最近日志 ===" && docker-compose logs --tail=20'
```

### 常用命令速查

```bash
# 快速重启
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose restart'

# 快速查看日志
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose logs --tail=50'

# 快速停止
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose down'
```

---

## ✅ 检查清单

**部署前**:
- □ 已注册 Alchemy 账号
- □ 已获取 API Key
- □ 已创建 MetaMask 钱包
- □ 已导出私钥和地址
- □ 已获取测试网 ETH
- □ 已配置 .env 文件
- □ 已设置 DRY_RUN=true

**部署后**:
- □ 容器成功启动
- □ 日志显示 "All systems operational"
- □ 账户余额 > 0
- □ 能看到 "Starting arbitrage detection loop"
- □ 无 ERROR 日志

**切换到真实模式前** (⚠️ 重要):
- □ 在测试网运行 > 24 小时
- □ 观察到正常的机会发现
- □ 理解了所有日志输出
- □ 准备了专用钱包
- □ 充分理解了风险
- □ 设置了合理的参数

---

**版本**: v1.0  
**更新时间**: 2024-02-06  
**服务器**: 47.93.253.224  
**环境**: Docker + Docker Compose  

**🎉 祝你运行顺利！遇到问题随时查看文档或提问。**
