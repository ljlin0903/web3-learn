# 🔧 测试环境配置说明

> 本配置已从 learning-project 复制，可直接使用

---

## ✅ 当前配置状态

### 网络配置
- **网络**: Sepolia 测试网
- **链 ID**: 11155111
- **RPC 提供商**: Alchemy

### 钱包配置
- **地址**: `0x57e6159fa93942F0a0a7A075EF8D0eb27cC60560`
- **私钥**: 已配置（从 learning-project）

### 运行模式
- **DRY_RUN**: `true` (模拟模式)
- **ENABLE_FLASHBOTS**: `false` (测试网不支持)

---

## 🚀 快速测试

### 1. 验证配置

```bash
cd /Users/ljlin/web3/mev-arbitrage-bot

# 查看关键配置
grep -E "^RPC_HTTPS_URL|^PUBLIC_ADDRESS|^NETWORK=" .env
```

### 2. 编译测试

```bash
# 编译
go build -o bin/mev-bot cmd/bot/main.go

# 查看二进制文件
ls -lh bin/mev-bot
```

### 3. 运行测试 (DRY RUN 模式)

```bash
# 直接运行
./bin/mev-bot

# 或使用 Docker
docker-compose up -d
docker-compose logs -f
```

---

## 📝 配置详情

### RPC 端点
```bash
RPC_HTTPS_URL=https://eth-sepolia.g.alchemy.com/v2/CRS3TfxjM5fIqYAjm4Oox
RPC_WSS_URL=wss://eth-sepolia.g.alchemy.com/v2/CRS3TfxjM5fIqYAjm4Oox
```

### 钱包信息
```bash
PUBLIC_ADDRESS=0x57e6159fa93942F0a0a7A075EF8D0eb27cC60560
```
**注意**: 私钥已配置但不显示（安全考虑）

### Sepolia 测试网代币地址
```bash
WETH_ADDRESS=0x7b79995e5f793A07Bc00c21412e50Ecae098E7f9
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
DAI_ADDRESS=0xFF34B3d4Aee8ddCd6F9AFFFB6Fe49bD371b8a357
```

### 策略参数
```bash
MIN_PROFIT_BPS=50              # 最小利润 0.5%
MAX_TRADE_AMOUNT_ETH=1         # 最大交易 1 ETH
MIN_TRADE_AMOUNT_ETH=0.01      # 最小交易 0.01 ETH
DRY_RUN=true                   # 模拟模式
```

---

## ⚠️ 重要提示

### 1. DRY RUN 模式
当前配置为 `DRY_RUN=true`，意味着：
- ✅ 会搜索套利机会
- ✅ 会计算利润
- ✅ 会构建交易
- ❌ **不会发送真实交易**
- 📝 所有操作只记录日志

### 2. 测试币余额
确保钱包有足够的测试 ETH：
```bash
# 钱包地址
0x57e6159fa93942F0a0a7A075EF8D0eb27cC60560

# 获取测试 ETH
https://sepoliafaucet.com/
```

### 3. Git 安全
- ✅ `.env` 文件已添加到 `.gitignore`
- ✅ 不会被 Git 追踪
- ⚠️ 请勿手动提交 `.env` 文件

---

## 🔄 切换到真实模式

### 步骤 1: 修改配置
```bash
# 编辑 .env 文件
vim .env

# 修改以下参数
DRY_RUN=false              # 关闭模拟模式
ENABLE_FLASHBOTS=true      # 如果使用主网
NETWORK=mainnet            # 如果切换到主网
```

### 步骤 2: 更新 RPC (如果切换主网)
```bash
# 主网 RPC
RPC_HTTPS_URL=https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
RPC_WSS_URL=wss://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
```

### 步骤 3: 更新代币地址 (如果切换主网)
```bash
# 主网代币地址
WETH_ADDRESS=0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2
USDC_ADDRESS=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
DAI_ADDRESS=0x6B175474E89094C44Da98b954EedeAC495271d0F
```

### ⚠️ 切换前检查清单
- [ ] 充分测试 DRY RUN 模式（至少 24 小时）
- [ ] 理解所有代码逻辑
- [ ] 准备充足的 Gas 费（主网）
- [ ] 设置合理的利润阈值
- [ ] 准备监控和告警
- [ ] 使用硬件钱包（主网）
- [ ] 小金额测试

---

## 🐛 故障排查

### 问题 1: 编译失败
```bash
# 清理缓存
go clean -cache

# 重新下载依赖
go mod tidy

# 重新编译
go build -o bin/mev-bot cmd/bot/main.go
```

### 问题 2: RPC 连接失败
```bash
# 测试 RPC 连接
curl -X POST https://eth-sepolia.g.alchemy.com/v2/CRS3TfxjM5fIqYAjm4Oox \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

### 问题 3: 没有找到套利机会
**正常现象**，因为：
- Sepolia 测试网流动性低
- 竞争者少，价差小
- DRY RUN 模式只是验证逻辑

---

## 📊 预期日志输出

### 启动日志
```
🤖 MEV Arbitrage Bot Starting...
   Version: v1.0.0
   Network: sepolia (11155111)
   Mode: DRY RUN (Simulation)
   
✅ Configuration loaded successfully
✅ Connected to Ethereum node
✅ Pool monitor initialized
✅ Strategy engine ready
🔍 Searching for arbitrage opportunities...
```

### 运行日志
```
[INFO] Monitoring 3 pools
[INFO] No profitable opportunities found (checked 150 paths)
[INFO] Best path profit: -0.23% (below threshold 0.5%)
```

### 模拟交易日志
```
🧪 DRY RUN MODE - Transaction not sent
💰 Opportunity: WETH -> USDC -> DAI -> WETH
📊 Profit: 0.78% (7.8 bps)
💸 Amount: 0.5 ETH
⛽ Gas: ~300,000
```

---

## ✅ 配置验证清单

- [x] ✅ RPC 端点已配置 (来自 learning-project)
- [x] ✅ 钱包私钥已配置
- [x] ✅ 公钥地址已配置
- [x] ✅ 网络设置正确 (Sepolia)
- [x] ✅ DRY_RUN 模式已启用
- [x] ✅ .env 文件已添加到 .gitignore
- [x] ✅ 代码编译成功
- [ ] ⏳ 钱包有测试 ETH (需检查)
- [ ] ⏳ 实际运行测试

---

## 📚 相关文档

- [服务器运行指南](RUN_ON_SERVER.md)
- [新手入门指南](GETTING_STARTED.md)
- [技术原理详解](TECHNICAL_GUIDE.md)

---

**配置时间**: 2024-02-06  
**来源**: learning-project  
**状态**: ✅ 就绪，可直接使用
