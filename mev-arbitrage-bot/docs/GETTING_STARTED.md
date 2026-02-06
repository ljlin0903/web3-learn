# 🚀 新手入门指南

> **写给完全不懂编程的你**: 这份文档会手把手教你如何理解和运行这个项目

---

## 📑 文档导航

**本项目提供 4 份文档，建议按顺序阅读**:

| 文档 | 适合人群 | 内容 |
|------|----------|------|
| **本文档** | 完全新手 | 从零开始的入门指南 |
| [TECHNICAL_GUIDE.md](TECHNICAL_GUIDE.md) | 非技术人员 | 通俗易懂的技术原理 |
| [CONTRACTS_GUIDE.md](CONTRACTS_GUIDE.md) | 想了解合约 | 智能合约详解 |
| [DEPLOYMENT.md](DEPLOYMENT.md) | 要部署运行 | Docker 部署指南 |

---

## 1️⃣ 第一步: 理解项目是什么

### 这个项目做什么？

**简单说**: 自动在不同交易所之间倒买倒卖赚差价

**举例**:
```
┌─────────────────────────────────────────┐
│ 发现价格差异                             │
├─────────────────────────────────────────┤
│ 交易所 A: 1 个比特币 = 60,000 美元       │
│ 交易所 B: 1 个比特币 = 61,000 美元       │
│                                          │
│ 套利操作                                 │
│ 1. 在 A 买入 1 个比特币 (花 60,000)      │
│ 2. 在 B 卖出 1 个比特币 (收 61,000)      │
│ 3. 赚到差价: 1,000 美元                 │
└─────────────────────────────────────────┘
```

**但是**:
- 我们在区块链上操作
- 速度要极快（毫秒级）
- 竞争激烈（很多机器人）

### 需要什么知识？

**不需要会编程！但需要理解这些概念**:

1. **区块链** = 一个公开的账本
2. **智能合约** = 自动执行的程序
3. **DEX** = 去中心化交易所
4. **套利** = 赚差价

详细解释请看 [TECHNICAL_GUIDE.md 第2章](TECHNICAL_GUIDE.md#2-核心概念解释)

---

## 2️⃣ 第二步: 准备环境

### 你需要准备的东西

**必需品**:
```
✅ 一台电脑 (Mac/Windows/Linux 都可以)
✅ 网络连接
✅ 以太坊钱包 (MetaMask)
✅ 一些测试币 (可免费获取)
```

**不需要**:
```
❌ 不需要安装 Go 语言 (在 Docker 里)
❌ 不需要安装 Node.js
❌ 不需要安装 Solidity
❌ 不需要真钱 (先用测试网)
```

### 获取测试币

**步骤**:
```
1. 安装 MetaMask 钱包
   https://metamask.io/

2. 切换到 Sepolia 测试网

3. 复制你的钱包地址

4. 访问水龙头获取测试 ETH
   https://sepoliafaucet.com/
   https://www.alchemy.com/faucets/ethereum-sepolia

5. 等待1-2分钟到账
```

### 注册 RPC 服务

**为什么需要？**
- 连接区块链需要通过 RPC 节点
- 自己运行节点太麻烦
- 使用第三方服务（免费）

**步骤**:
```
1. 访问 Alchemy
   https://www.alchemy.com/

2. 注册账号（免费）

3. 创建新项目
   - 选择 "Ethereum"
   - 选择 "Sepolia" 测试网

4. 获取 API Key
   - 点击 "View Key"
   - 复制 HTTPS URL
   - 复制 WebSocket URL

5. 保存这两个地址，后面要用
```

---

## 3️⃣ 第三步: 配置项目

### 创建配置文件

**位置**: `/Users/ljlin/web3/mev-arbitrage-bot/.env`

**内容**:
```bash
# ==================== 网络配置 ====================
# 使用测试网 (不花真钱)
NETWORK=sepolia

# RPC 地址 (从 Alchemy 获取)
RPC_HTTPS_URL=https://eth-sepolia.g.alchemy.com/v2/你的API密钥
RPC_WSS_URL=wss://eth-sepolia.g.alchemy.com/v2/你的API密钥

# ==================== 钱包配置 ====================
# 你的钱包私钥 (从 MetaMask 导出)
# ⚠️ 警告: 永远不要泄露私钥！
PRIVATE_KEY=0x你的私钥

# 你的钱包地址
PUBLIC_ADDRESS=0x你的地址

# ==================== 策略参数 ====================
# 最小利润要求 (100 = 1%)
MIN_PROFIT_BPS=100

# 单笔最大金额 (ETH)
MAX_TRADE_AMOUNT_ETH=1

# ==================== 运行模式 ====================
# 模拟模式 (不真实交易)
DRY_RUN=true

# 不使用 Flashbots (测试网不需要)
ENABLE_FLASHBOTS=false

# 日志级别
LOG_LEVEL=info
```

### 如何获取私钥？

**MetaMask 导出私钥**:
```
1. 打开 MetaMask
2. 点击右上角三个点
3. 选择 "账户详情"
4. 点击 "导出私钥"
5. 输入密码
6. 复制私钥 (0x开头的一串字符)
```

⚠️ **安全提醒**:
- 私钥 = 你的钱包密码
- 永远不要分享给别人
- 不要发到群里
- 不要截图发朋友圈
- 测试网的私钥可以公开，主网的绝对不行

---

## 4️⃣ 第四步: 理解代码结构

### 项目目录

```
mev-arbitrage-bot/
├── cmd/bot/          【大脑】主程序
├── pkg/              【器官】各种功能
│   ├── config/      配置管理
│   ├── blockchain/  连接区块链
│   ├── dex/         交易所交互
│   ├── strategy/    寻找机会
│   ├── executor/    执行交易
│   ├── flashbots/   防护机制
│   └── utils/       工具函数
├── contracts/        【手】智能合约
└── 文档/             你正在读的文档
```

### 代码是如何工作的？

**简化流程**:
```
1. 启动程序
   ↓
2. 连接区块链
   ↓
3. 监控价格变化
   ↓
4. 发现价格差异
   ↓
5. 计算是否盈利
   ↓
6. 执行交易
   ↓
7. 等待确认
   ↓
8. 记录结果
   ↓
9. 回到步骤 3
```

详细流程请看 [TECHNICAL_GUIDE.md 第4章](TECHNICAL_GUIDE.md#4-完整工作流程)

---

## 5️⃣ 第五步: 运行项目

### 方式 A: Docker 运行 (推荐)

**优点**: 不需要安装任何开发环境

**步骤**:
```bash
# 1. 确保已配置 .env 文件

# 2. 运行一键部署脚本
cd /Users/ljlin/web3/mev-arbitrage-bot
./scripts/deploy-docker.sh

# 3. 查看日志
ssh root@47.93.253.224 'cd /opt/mev-arbitrage-bot && docker-compose logs -f'
```

详细说明请看 [DEPLOYMENT.md](DEPLOYMENT.md)

### 方式 B: 本地运行 (开发者)

**前提**: 需要安装 Go 1.21+

**步骤**:
```bash
# 1. 编译程序
cd /Users/ljlin/web3/mev-arbitrage-bot
go build -o bin/mev-bot cmd/bot/main.go

# 2. 运行
./bin/mev-bot

# 预期输出:
========================================
  🤖 MEV Arbitrage Bot v1.0
========================================
✅ Config loaded
✅ Connected to Sepolia
✅ Balance: 1.5 ETH
✅ All systems operational!
🧪 Running in DRY RUN mode
Waiting for arbitrage opportunities...
```

---

## 6️⃣ 第六步: 查看日志

### 理解日志输出

**正常日志**:
```
INFO[0001] ✅ Config loaded
INFO[0002] ✅ Connected to blockchain
INFO[0003] Monitoring 5 pools
INFO[0010] Searching for arbitrage opportunities...
INFO[0015] ⚠️  No opportunities found (too low profit)
```

**发现机会**:
```
INFO[0120] 🎯 Arbitrage opportunity found!
INFO[0120] Path: ETH → USDC → DAI → ETH
INFO[0120] Profit: 0.05 ETH (5%)
INFO[0120] Net Profit: 0.04 ETH (4% after gas)
INFO[0120] 🧪 DRY RUN MODE - Transaction not sent
```

**执行交易** (关闭 DRY_RUN 后):
```
INFO[0200] 🎯 Executing arbitrage...
INFO[0201] 📡 Sending via Flashbots
INFO[0202] Simulation successful
INFO[0203] Bundle sent successfully
INFO[0215] ✅ Transaction confirmed!
INFO[0215] Profit realized: 0.04 ETH
```

**错误日志**:
```
ERROR[0100] Failed to connect to RPC
ERROR[0200] Gas price too high, skipping
ERROR[0300] Transaction reverted
```

### 日志级别

| 级别 | 含义 | 何时出现 |
|------|------|----------|
| DEBUG | 调试信息 | 开发时 |
| INFO | 正常信息 | 一般运行 |
| WARN | 警告 | 有问题但能继续 |
| ERROR | 错误 | 出错了 |

---

## 7️⃣ 第七步: 从模拟到真实

### 模拟模式 (DRY_RUN=true)

**特点**:
- ✅ 不发送真实交易
- ✅ 不消耗 Gas
- ✅ 看得到机会
- ✅ 适合学习

**用途**:
```
1. 理解系统如何工作
2. 观察能发现哪些机会
3. 验证配置是否正确
4. 测试不同参数
```

### 切换到真实模式

**步骤**:
```bash
# 1. 编辑 .env 文件
vim .env

# 2. 修改这一行
DRY_RUN=false  # 改成 false

# 3. 确认钱包有足够余额
# 至少需要: 0.1 ETH (用于 Gas)

# 4. 重启程序
```

**⚠️ 重要提醒**:
```
测试网 (Sepolia):
✅ 币是免费的
✅ 可以随便测试
✅ 不会亏真钱

主网 (Mainnet):
⚠️  真金白银！
⚠️  Gas 费很贵
⚠️  竞争激烈
⚠️  新手请勿直接用！
```

---

## 8️⃣ 第八步: 优化参数

### 可以调整的参数

**在 .env 文件中**:

```bash
# 利润要求
MIN_PROFIT_BPS=100    # 降低 → 更多机会，但利润低
                      # 提高 → 更少机会，但利润高

# 交易金额
MAX_TRADE_AMOUNT_ETH=1   # 提高 → 利润更多，但风险大
                         # 降低 → 风险小，但利润少

# Gas 策略
GAS_PRICE_MULTIPLIER=1.2  # 提高 → 更快被打包，但更贵
                          # 降低 → 更便宜，但可能不中

MAX_GAS_PRICE_GWEI=100    # Gas 超过这个值就不交易
```

### 优化建议

**保守策略** (新手):
```bash
MIN_PROFIT_BPS=200       # 要求 2% 以上
MAX_TRADE_AMOUNT_ETH=0.5 # 限制金额
GAS_PRICE_MULTIPLIER=1.5 # 保证能被打包
```

**激进策略** (老手):
```bash
MIN_PROFIT_BPS=30        # 0.3% 就执行
MAX_TRADE_AMOUNT_ETH=10  # 更大金额
GAS_PRICE_MULTIPLIER=1.1 # 降低成本
```

**测试策略** (学习):
```bash
MIN_PROFIT_BPS=10        # 看更多机会
DRY_RUN=true             # 不真实交易
LOG_LEVEL=debug          # 看详细日志
```

---

## 9️⃣ 第九步: 常见问题

### Q: 为什么一直找不到机会？

**可能原因**:
```
1. MIN_PROFIT_BPS 设置太高
   → 解决: 降低到 50 试试

2. 测试网流动性不足
   → 解决: 正常现象，测试网交易少

3. 监控的池子太少
   → 解决: 需要添加更多池子 (代码改动)

4. 竞争激烈
   → 解决: 需要更快的执行速度
```

### Q: 为什么交易失败？

**常见原因**:
```
1. 价格变动太快
   → 在你发送交易的时候价格已经变了

2. Gas 太低
   → 增加 GAS_PRICE_MULTIPLIER

3. 流动性不足
   → 减小 MAX_TRADE_AMOUNT_ETH

4. 被抢跑
   → 启用 Flashbots (主网)
```

### Q: 如何看懂代码？

**推荐阅读顺序**:
```
1. TECHNICAL_GUIDE.md → 理解原理
2. pkg/config/ → 看配置怎么加载
3. pkg/blockchain/ → 看怎么连区块链
4. pkg/strategy/ → 看怎么找机会
5. cmd/bot/main.go → 看整体流程
```

### Q: 能改成监控 XXX 代币吗？

**回答**: 可以，需要改代码

**需要修改的地方**:
```go
// 在 pkg/dex/pool_monitor.go 中

// 添加要监控的代币对
pairs := []TokenPair{
    {WETH, USDC},
    {WETH, DAI},
    {WETH, XXX},  // ← 添加你的代币
}
```

---

## 🔟 第十步: 下一步学习

### 深入理解

**推荐学习路径**:

```
第 1 周: 理解概念
├─ 阅读 TECHNICAL_GUIDE.md
├─ 了解区块链基础
└─ 了解 DeFi 基础

第 2 周: 实践操作
├─ 在测试网运行
├─ 观察日志输出
└─ 尝试调整参数

第 3 周: 深入代码
├─ 阅读关键模块
├─ 理解数据流
└─ 尝试小改动

第 4 周: 高级主题
├─ 学习智能合约
├─ 了解 Flashbots
└─ 研究优化策略
```

### 学习资源

**区块链基础**:
- Ethereum.org: https://ethereum.org/zh/developers/
- 以太坊白皮书: https://ethereum.org/zh/whitepaper/

**Go 语言** (如果想改代码):
- Go 语言之旅: https://tour.golang.org/
- Go by Example: https://gobyexample.com/

**DeFi 和套利**:
- Uniswap 文档: https://docs.uniswap.org/
- Aave 文档: https://docs.aave.com/

**MEV 相关**:
- Flashbots 文档: https://docs.flashbots.net/
- MEV 研究: https://www.mev.wiki/

---

## 📞 获取帮助

### 遇到问题怎么办？

**步骤 1: 检查日志**
```bash
# 查看最近的错误
grep ERROR logs/*.log

# 查看完整日志
tail -f logs/app.log
```

**步骤 2: 查看文档**
- 技术问题 → [TECHNICAL_GUIDE.md](TECHNICAL_GUIDE.md)
- 合约问题 → [CONTRACTS_GUIDE.md](CONTRACTS_GUIDE.md)
- 部署问题 → [DEPLOYMENT.md](DEPLOYMENT.md)

**步骤 3: 检查配置**
```bash
# 验证配置
cat .env

# 检查语法错误
go build cmd/bot/main.go
```

**步骤 4: 重新开始**
```bash
# 停止程序
Ctrl+C

# 清理日志
rm -rf logs/*.log

# 重新启动
./bin/mev-bot
```

---

## ✅ 检查清单

**开始前确认**:

```
□ 已安装 MetaMask 钱包
□ 已获取测试网 ETH
□ 已注册 Alchemy 账号
□ 已复制 RPC 地址
□ 已导出钱包私钥
□ 已创建 .env 文件
□ 已设置 DRY_RUN=true
```

**运行前确认**:

```
□ .env 文件配置正确
□ 钱包有足够余额 (>0.1 ETH)
□ 网络连接正常
□ 确认使用测试网
```

**上主网前确认** (重要！):

```
□ 已在测试网充分测试
□ 理解了所有风险
□ 准备了专用钱包
□ 资金在可承受范围内
□ 已启用 Flashbots
□ 已降低 MAX_TRADE_AMOUNT
```

---

## 🎯 总结

### 你现在应该知道的

✅ **概念理解**:
- 什么是套利
- 什么是闪电贷
- 什么是 Flashbots

✅ **操作能力**:
- 配置项目
- 运行程序
- 查看日志
- 调整参数

✅ **安全意识**:
- 保护私钥
- 先用测试网
- 小金额开始
- 监控异常

### 下一步行动

**立即可做**:
1. 在测试网运行
2. 观察 1-2 天
3. 理解日志输出
4. 尝试调参数

**深入学习** (选做):
1. 阅读代码
2. 学习 Go 语言
3. 学习 Solidity
4. 研究 MEV

**进阶操作** (谨慎):
1. 部署智能合约
2. 切换到主网
3. 优化策略
4. 提高收益

---

**文档版本**: v1.0  
**更新时间**: 2024-02-06  
**适合人群**: 完全新手  
**预计阅读时间**: 30 分钟

**祝你学习顺利！有问题随时提问 🚀**
