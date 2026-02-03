# Web3 量化交易机器人 - 项目结构文档

## 📁 项目目录结构

```
web3-quant/
├── cmd/                           # 可执行程序入口 (Executables Entry Points)
│   ├── monitor/                   # 区块监听器 (Block Monitor)
│   │   └── main.go               # 监听新区块并输出
│   └── arbitrage/                 # 套利执行器 (Arbitrage Executor)
│       └── main.go               # 扫描并执行套利机会
│
├── pkg/                           # 可复用的库代码 (Reusable Libraries)
│   ├── blockchain/                # 区块链交互层 (Blockchain Layer)
│   │   ├── client.go             # 以太坊客户端封装
│   │   └── monitor.go            # 区块监听核心逻辑
│   ├── dex/                       # DEX 交互层 (DEX Layer)
│   │   ├── uniswap.go            # Uniswap V2/V3 接口
│   │   ├── amm.go                # AMM 价格计算
│   │   └── pool.go               # 池子监控
│   └── strategy/                  # 策略层 (Strategy Layer)
│       ├── arbitrage.go          # 套利策略
│       └── flashloan.go          # 闪电贷策略
│
├── contracts/                     # Solidity 智能合约 (Smart Contracts)
│   ├── src/                       # 合约源码
│   │   ├── FlashArbitrage.sol    # 普通套利合约
│   │   └── FlashLoanArbitrage.sol # 闪电贷套利合约
│   ├── test/                      # 合约测试
│   │   ├── FlashArbitrage.t.sol
│   │   └── FlashLoanArbitrage.t.sol
│   ├── script/                    # 部署脚本
│   ├── out/                       # 编译输出
│   └── foundry.toml              # Foundry 配置
│
├── scripts/                       # 部署和运维脚本 (Deployment Scripts)
│   ├── deploy.sh                 # 服务器部署脚本
│   └── deploy_expect.sh          # 带密码认证的部署
│
├── docs/                          # 项目文档 (Documentation)
│   ├── architecture.md           # 架构设计文档
│   ├── api.md                    # API 文档
│   └── deployment.md             # 部署指南
│
├── .env                          # 环境变量（私钥等敏感信息）
├── .env.example                  # 环境变量模板
├── go.mod                        # Go 依赖管理
├── go.sum                        # Go 依赖锁定
├── Dockerfile                    # Docker 镜像构建
├── docker-compose.yml            # Docker Compose 配置
└── README.md                     # 项目说明文档
```

---

## 🗂️ 当前文件分类与作用

### **1. Go 源文件（需重构）**

| 当前位置 | 作用 | 建议迁移位置 |
|---------|------|------------|
| `main.go` | 基础测试：连接 RPC，查询余额 | 保持根目录（作为快速测试） |
| `monitor.go` | 简单区块监听器 | → `pkg/blockchain/monitor_simple.go` |
| `monitor_reconnect.go` | 带重连的区块监听器 | → `cmd/monitor/main.go` |
| `pool_monitor.go` | Uniswap 池子监听 | → `pkg/dex/pool.go` |
| `amm_calculator.go` | AMM 价格计算 | → `pkg/dex/amm.go` |
| `arbitrage_finder.go` | 三角套利路径搜索 | → `pkg/strategy/arbitrage.go` |
| `flashloan_monitor.go` | 闪电贷套利监控器 | → `cmd/arbitrage/main.go` |

### **2. Solidity 合约（已规范）**

| 文件 | 作用 | 状态 |
|-----|------|------|
| `contracts/src/FlashArbitrage.sol` | 普通三角套利合约（需自有资金） | ✅ 完成 |
| `contracts/src/FlashLoanArbitrage.sol` | 闪电贷套利合约（无需资金） | ✅ 完成 |
| `contracts/test/*.t.sol` | Foundry 测试套件 | ✅ 完成 |

### **3. 部署文件**

| 文件 | 作用 | 状态 |
|-----|------|------|
| `Dockerfile` | Docker 镜像构建（多阶段构建） | ✅ 完成 |
| `docker-compose.yml` | 服务编排配置 | ✅ 完成 |
| `deploy.sh` | SSH 部署脚本（需 sshpass） | ✅ 完成 |
| `deploy_expect.sh` | Expect 部署脚本（密码认证） | ✅ 完成 |
| `daemon.json` | Docker 镜像加速配置 | ✅ 完成 |

### **4. 配置文件**

| 文件 | 作用 | 状态 |
|-----|------|------|
| `.env` | 环境变量（私钥、RPC URL 等） | ✅ 完成 |
| `.env.example` | 环境变量模板 | ✅ 完成 |
| `go.mod` / `go.sum` | Go 依赖管理 | ✅ 完成 |
| `foundry.toml` | Foundry 配置 | ✅ 已存在 |

### **5. 文档文件**

| 文件 | 作用 | 状态 |
|-----|------|------|
| `TECH_PLAN.md` | 技术方案 | ✅ 完成 |
| `SESSION_LOG.md` | 开发日志 | ✅ 完成 |
| `PROJECT_STRUCTURE.md` | 本文档 | ✅ 新增 |

---

## 🏗️ 系统架构流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                    Web3 量化交易机器人系统架构                      │
└─────────────────────────────────────────────────────────────────┘

                            ┌──────────────┐
                            │  Sepolia     │
                            │  测试网      │
                            │  (主网)      │
                            └──────┬───────┘
                                   │
                                   │ WebSocket / HTTPS
                                   │
                    ┌──────────────▼──────────────┐
                    │   Alchemy RPC 节点          │
                    │   wss://eth-sepolia...      │
                    └──────────────┬──────────────┘
                                   │
                ┌──────────────────┼──────────────────┐
                │                  │                  │
                ▼                  ▼                  ▼
    ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
    │  区块监听器      │  │  池子监听器      │  │  价格查询器      │
    │  monitor.go     │  │  pool_monitor   │  │  amm_calculator │
    └────────┬────────┘  └────────┬────────┘  └────────┬────────┘
             │                    │                    │
             │ 新区块事件          │ Swap 事件           │ 价格数据
             │                    │                    │
             └────────────────────┼────────────────────┘
                                  │
                        ┌─────────▼─────────┐
                        │   策略引擎         │
                        │   arbitrage_finder│
                        │   (路径搜索)       │
                        └─────────┬─────────┘
                                  │
                    发现套利机会   │ (利润 > 阈值)
                                  │
                        ┌─────────▼─────────┐
                        │   决策层           │
                        │   flashloan_monitor│
                        │   (链下模拟)       │
                        └─────────┬─────────┘
                                  │
                          确认盈利 │ 
                                  │
                        ┌─────────▼─────────┐
                        │   执行层           │
                        │   (发送交易)       │
                        └─────────┬─────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                │                 │                 │
                ▼                 ▼                 ▼
    ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
    │ 普通套利合约      │  │ 闪电贷套利合约    │  │ 直接 DEX 交易     │
    │ FlashArbitrage   │  │ FlashLoanArbitrage│  │ (小额测试)        │
    └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
             │                     │                     │
             │                     │ 闪电贷               │
             ▼                     ▼                     ▼
    ┌──────────────────────────────────────────────────────┐
    │              Ethereum 主网 / Sepolia 测试网            │
    │         Uniswap V2/V3 · SushiSwap · Aave            │
    └──────────────────────────────────────────────────────┘
```

---

## 🔄 数据流详解

### **监听流程**
```
1. WebSocket 连接 Alchemy
   ↓
2. 订阅 newHeads 事件（新区块）
   ↓
3. 订阅 logs 事件（Swap 事件）
   ↓
4. 实时接收区块数据
   ↓
5. 解析 Swap 事件参数
   ↓
6. 更新池子储备量
```

### **套利发现流程**
```
1. 获取多个 DEX 池子数据
   ↓
2. 构建代币交易图
   ↓
3. 搜索三角套利路径
   ↓
4. 计算每条路径利润
   ↓
5. 考虑 Gas 费用和手续费
   ↓
6. 筛选出利润 > 阈值的机会
```

### **闪电贷套利流程（单笔交易原子执行）**
```
链下：
1. 扫描套利机会
2. 调用合约 simulateArbitrage 模拟
3. 确认盈利后发起交易

链上（单笔交易内）：
1. 从 Aave 借入资金（如 100 ETH）
   ↓
2. executeOperation 回调：
   ├─ Swap 1: ETH → USDC (Uniswap)
   ├─ Swap 2: USDC → DAI (SushiSwap)
   └─ Swap 3: DAI → ETH (Uniswap)
   ↓
3. 检查利润是否足够
   ↓
4. 归还借款 + 0.09% 手续费
   ↓
5. 保留利润

如果任何步骤失败 → 整个交易回滚（不消耗 Gas）
```

---

## 🚀 部署方式

### **方式 1: 本地开发测试**

#### **Go 程序**
```bash
# 1. 安装依赖
go mod download

# 2. 运行基础测试
go run main.go

# 3. 运行区块监听器
go run monitor_reconnect.go

# 4. 运行套利监控器
go run flashloan_monitor.go
```

#### **Solidity 合约**
```bash
# 进入合约目录
cd contracts

# 编译合约
forge build

# 运行测试
forge test -vv

# 测试单个合约
forge test --match-contract FlashLoanArbitrageTest -vv
```

---

### **方式 2: Docker 本地部署**

```bash
# 构建镜像
docker build -t web3-monitor .

# 运行容器
docker run -d --name web3-monitor \
  --env-file .env \
  --restart unless-stopped \
  web3-monitor

# 查看日志
docker logs -f web3-monitor

# 使用 Docker Compose
docker-compose up -d
docker-compose logs -f
```

---

### **方式 3: 服务器自动化部署**

```bash
# 方式 A: 使用 sshpass（需先安装）
brew install hudochenkov/sshpass/sshpass
./deploy.sh

# 方式 B: 使用 expect（macOS 自带）
./deploy_expect.sh
```

**部署脚本执行流程**：
```
1. 检查服务器 Docker 环境
   ↓
2. 创建部署目录 /opt/web3-quant
   ↓
3. rsync 上传项目文件
   ↓
4. 构建 Docker 镜像
   ↓
5. 启动容器并自动重启
```

---

### **方式 4: 主网部署（真实交易）**

⚠️ **警告：主网部署需要真实资金和充分测试**

#### **步骤 1: 部署合约**
```bash
cd contracts

# 部署到主网
forge create --rpc-url https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY \
  --private-key $PRIVATE_KEY \
  src/FlashLoanArbitrage.sol:FlashLoanArbitrage \
  --constructor-args 0x2f39d218133AFaB8F2B819B1066c7E434Ad94E9e \
  --verify

# 记录合约地址
export FLASHLOAN_CONTRACT=0x...
```

**Aave V3 Pool Addresses Provider**:
- **Mainnet**: `0x2f39d218133AFaB8F2B819B1066c7E434Ad94E9e`
- **Sepolia**: `0x012bAC54348C0E635dCAc9D5FB99f06F24136C9A`

#### **步骤 2: 配置监控器**
```bash
# 修改 .env
RPC_HTTPS_URL=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
RPC_WSS_URL=wss://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
FLASHLOAN_CONTRACT_ADDRESS=0x...
PRIVATE_KEY=your_private_key
```

#### **步骤 3: 启动监控**
```bash
# 编译
go build -o arbitrage-bot flashloan_monitor.go

# 后台运行
nohup ./arbitrage-bot > bot.log 2>&1 &

# 查看日志
tail -f bot.log
```

#### **步骤 4: 使用 Flashbots（防 MEV）**
```go
// 集成 Flashbots RPC
// 发送交易到 Flashbots relay 而非公开 mempool
// 防止被前置交易攻击
```

---

## 📊 各组件技术栈

| 组件 | 语言 | 框架/库 | 用途 |
|-----|------|--------|------|
| 区块监听 | Go | go-ethereum | WebSocket 连接，事件订阅 |
| 数据处理 | Go | math/big | 大数计算（wei/ether） |
| 策略引擎 | Go | 自研算法 | 套利路径搜索 |
| 智能合约 | Solidity | - | 链上执行逻辑 |
| 合约测试 | Solidity | Foundry | 单元测试、集成测试 |
| 容器化 | Docker | Alpine Linux | 轻量级运行环境 |
| 部署 | Shell | rsync, ssh | 自动化部署 |

---

## 🔐 安全建议

### **测试网环境（当前）**
- ✅ 私钥可以提交到 Git（测试用）
- ✅ 免费测试币无经济风险
- ✅ 用于学习和调试

### **主网环境（生产）**
- ❌ 绝不提交私钥到 Git
- ✅ 使用硬件钱包或 HSM
- ✅ 设置 Gas 上限
- ✅ 启用多签钱包
- ✅ 使用 Flashbots 防 MEV
- ✅ 监控异常交易

---

## 📈 性能指标

| 指标 | 测试网 | 主网预期 |
|-----|-------|---------|
| 区块延迟 | ~12秒 | ~12秒 |
| WebSocket 延迟 | ~100ms | ~50ms (优化后) |
| 套利机会发现 | 模拟数据 | 每小时 0-5 次 |
| 单次 Gas 消耗 | ~271,730 gas | ~250,000 gas |
| 闪电贷手续费 | 0.09% | 0.09% |
| 目标利润率 | > 1% | > 0.5% (扣除 Gas) |

---

## 🛠️ 开发工具链

```bash
# Go 开发
go version      # 1.21+
go fmt          # 格式化代码
go vet          # 静态检查
go test         # 运行测试

# Solidity 开发
forge --version # Foundry 版本
forge build     # 编译合约
forge test      # 测试合约
forge fmt       # 格式化合约

# Docker
docker --version
docker-compose --version
```

---

## 📝 下一步优化建议

### **代码结构优化**
1. ✅ 按照上述结构重构 Go 代码
2. ✅ 提取公共逻辑到 pkg/
3. ✅ 完善错误处理和日志

### **功能增强**
1. 添加更多 DEX 支持（SushiSwap、Curve）
2. 实现多池子套利（4-5 跳）
3. 集成 Flashbots
4. 添加 Telegram 通知

### **性能优化**
1. 并发处理多个池子
2. 缓存池子状态
3. 使用本地节点（减少延迟）

### **监控告警**
1. Prometheus + Grafana
2. 异常交易告警
3. Gas 价格监控

---

**更新时间**: 2026-02-03  
**版本**: v1.0  
**状态**: 测试网运行中
