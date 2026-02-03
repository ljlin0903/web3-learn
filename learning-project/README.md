# 🚀 快速开始指南 - 从零开始运行 Web3 套利机器人

> **适合人群**：完全没接触过 Web3 开发的新手  
> **阅读时间**：10 分钟  
> **动手实践**：30 分钟

---

## 📖 文档阅读顺序

如果你是**完全的新手**，建议按这个顺序阅读：

```
1️⃣ README.md (本文档)           ← 你现在在这里
   ↓ 了解项目是什么
   
2️⃣ TECH_PLAN.md                 ← 技术方案（5分钟快速浏览）
   ↓ 知道要做什么
   
3️⃣ SESSION_LOG.md               ← 开发日志（看"阶段总结"部分）
   ↓ 看看别人怎么做的
   
4️⃣ PROJECT_STRUCTURE.md         ← 项目结构（重点看"部署方式"）
   ↓ 动手跑起来
   
5️⃣ ARCHITECTURE_DIAGRAMS.md     ← 流程图（可选，深入理解用）
```

---

## 🎯 这个项目是什么？

**一句话总结**：监听以太坊区块链，自动发现并执行套利交易赚钱的机器人

**通俗解释**：
```
想象你在三个货币兑换点之间发现了汇率差：
- A 点：1 美元 = 7 人民币
- B 点：7 人民币 = 60 日元  
- C 点：60 日元 = 1.1 美元

你发现循环一圈能赚 0.1 美元！

这个项目就是自动发现这种机会并瞬间完成交易。
```

**核心特点**：
- ✅ **无需自己的钱**：使用闪电贷（借钱-赚钱-还钱在1秒内完成）
- ✅ **零风险**：不赚钱就自动取消，不会亏损
- ✅ **全自动**：24/7 运行，发现机会就执行

---

## 🔑 核心概念 5 分钟速成

### 1. 什么是区块链？
**简单理解**：一个公开的账本，所有交易记录都在上面

### 2. 什么是以太坊？
**简单理解**：一个可以运行程序的区块链（我们的机器人就运行在这上面）

### 3. 什么是 DEX（去中心化交易所）？
**简单理解**：没有老板的交易所，代码自动运行
- Uniswap：最大的 DEX
- SushiSwap：第二大的 DEX

### 4. 什么是套利？
**简单理解**：
```
Uniswap：1 ETH = 2000 USDC
SushiSwap：1 ETH = 2010 USDC

→ 在 Uniswap 买入 → 在 SushiSwap 卖出 → 赚 10 USDC
```

### 5. 什么是闪电贷？
**简单理解**：
```
普通贷款：
借钱 → 等几天 → 还钱

闪电贷：
借钱 → 赚钱 → 还钱（全部在 1 秒内完成！）
如果赚不到钱，交易自动取消，不会发生
```

---

## 🛠️ 环境准备（10 分钟）

### 第一步：安装 Go 语言
```bash
# macOS
brew install go

# 验证安装
go version  # 应该显示：go version go1.21.x
```

### 第二步：安装 Foundry（智能合约工具）
```bash
# 下载并安装
curl -L https://foundry.paradigm.xyz | bash
foundryup

# 验证安装
forge --version
```

### 第三步：克隆项目
```bash
cd ~/Desktop  # 或其他你喜欢的位置
git clone https://github.com/ljlin0903/web3-learn.git
cd web3-learn/learning-project
```

### 第四步：配置环境变量
```bash
# 复制配置模板
cp .env.example .env

# 编辑配置文件
vim .env  # 或用任何编辑器打开
```

**需要填写的内容**：
```bash
# 1. 获取 Alchemy API Key（免费）
# 访问 https://www.alchemy.com/
# 注册 → 创建 App → 选择 Sepolia 测试网 → 复制 API Key

RPC_HTTPS_URL=https://eth-sepolia.g.alchemy.com/v2/你的API_KEY
RPC_WSS_URL=wss://eth-sepolia.g.alchemy.com/v2/你的API_KEY

# 2. 创建钱包（用于测试）
# 访问 https://metamask.io/
# 安装插件 → 创建钱包 → 切换到 Sepolia 测试网

PRIVATE_KEY=你的私钥  # MetaMask 设置里可以导出
PUBLIC_ADDRESS=你的钱包地址

# 3. 领取测试币（免费）
# 访问 https://sepoliafaucet.com/
# 输入你的地址 → 领取 0.5 ETH（测试币，无价值）
```

---

## 🎮 运行第一个程序（5 分钟）

### 测试 1：检查连接
```bash
go run main.go
```

**预期输出**：
```
[步骤1] ✓ .env 文件加载成功
[步骤2] RPC 节点: https://eth-sepolia...
[步骤3] ✓ 成功连接到以太坊网络!
[步骤4] 正在查询账户余额...

=== 查询结果 ===
钱包地址: 0x57e6159fa...
账户余额: 0.500000 ETH

✅ 测试验证通过！环境配置正常。
```

### 测试 2：监听区块
```bash
go run cmd/monitor/main.go
```

**预期输出**：
```
🚀 监听中... (按 Ctrl+C 退出)

📦 新区块 #1 | Block Number: 10183442
   ⏰ 时间: 2026-02-03 23:12:00
   🔗 区块哈希: 0x6d90aa1bc3...
   ⛽ Gas 使用: 13313715
```

**看到这个说明成功了！** 按 `Ctrl+C` 停止。

### 测试 3：运行合约测试
```bash
cd contracts
forge test -vv
```

**预期输出**：
```
Running 10 tests...

[PASS] testFlashLoanArbitrageSuccess() (gas: 271730)
  Loan Amount: 100 TKA
  Profit: 19 TKA
  Profit %: 1991 bps

Suite result: ok. 10 passed; 0 failed
```

---

## 📚 代码文件导读

### 最重要的 3 个文件

#### 1. `main.go` - 快速测试
**作用**：验证环境配置是否正确  
**何时用**：每次修改配置后
```go
// 做了什么？
1. 读取 .env 文件
2. 连接以太坊节点
3. 查询钱包余额
```

#### 2. `cmd/monitor/main.go` - 区块监听器
**作用**：实时接收新区块信息  
**何时用**：了解区块链实时状态
```go
// 做了什么？
1. 建立 WebSocket 连接
2. 订阅新区块事件
3. 显示区块详情
4. 自动重连（如果断线）
```

#### 3. `contracts/src/FlashLoanArbitrage.sol` - 闪电贷合约
**作用**：在链上执行套利交易  
**何时用**：发现套利机会时自动调用
```solidity
// 做了什么？
1. 从 Aave 借入资金（如 100 ETH）
2. 在 Uniswap 交易：ETH → USDC
3. 在 SushiSwap 交易：USDC → DAI
4. 再回到 Uniswap：DAI → ETH
5. 归还借款 + 手续费
6. 保留利润
```

---

## 🎓 学习路径

### Level 1：环境熟悉（已完成 ✅）
- [x] 安装工具
- [x] 配置环境
- [x] 运行测试程序

### Level 2：理解原理（30 分钟）
```bash
# 1. 看技术方案
cat TECH_PLAN.md

# 2. 理解 AMM 价格计算
go run main.go  # 然后阅读 pkg/dex/amm.go 源码
```

**重点理解**：
- Uniswap 如何定价？（恒定乘积公式）
- 手续费怎么算？（0.3%）
- 为什么会有套利机会？（价格不同步）

### Level 3：修改代码（1 小时）
**任务 1**：修改监控器显示内容
```go
// 打开 cmd/monitor/main.go
// 找到这行：
fmt.Printf("📦 新区块 #%d\n", blockNumber)

// 改成：
fmt.Printf("🎉 发现新区块！编号：%d\n", blockNumber)

// 重新运行看效果
go run cmd/monitor/main.go
```

**任务 2**：调整套利参数
```go
// 打开 pkg/strategy/arbitrage.go
// 找到：
minProfitBps := 100  // 最低 1% 利润

// 改成：
minProfitBps := 50   // 最低 0.5% 利润
```

### Level 4：部署到服务器（可选）
```bash
# 如果你有自己的服务器
./scripts/deploy_expect.sh
```

---

## ❓ 常见问题 FAQ

### Q1: 运行 main.go 报错 "Error loading .env file"
**原因**：没有配置 .env 文件  
**解决**：
```bash
cp .env.example .env
vim .env  # 填写你的 API Key
```

### Q2: 连接超时 "failed to connect"
**原因**：网络问题或 API Key 错误  
**解决**：
1. 检查网络连接
2. 确认 API Key 是否正确
3. 尝试科学上网

### Q3: "command not found: go"
**原因**：Go 没装好  
**解决**：
```bash
# macOS
brew install go
source ~/.zshrc  # 重新加载配置

# 验证
go version
```

### Q4: forge test 失败
**原因**：Foundry 没装好  
**解决**：
```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
forge --version
```

### Q5: 测试网没钱了
**解决**：重新领取
```bash
# 访问水龙头
open https://sepoliafaucet.com/
# 输入你的地址领取
```

---

## 🎯 下一步做什么？

### 如果你想深入学习：
1. **阅读代码**：
   ```bash
   # 从简单到复杂
   cat main.go                          # 基础
   cat pkg/dex/amm.go                   # AMM 算法
   cat pkg/strategy/arbitrage.go        # 套利策略
   cat cmd/arbitrage/main.go            # 完整流程
   ```

2. **修改参数实验**：
   - 调整最低利润率
   - 修改日志输出格式
   - 添加自己的判断逻辑

3. **阅读测试用例**：
   ```bash
   cd contracts
   cat test/FlashLoanArbitrage.t.sol
   ```

### 如果你想实际赚钱（⚠️ 高风险）：
1. **切换到主网** （需要真钱）
2. **部署合约** （需要 Gas 费）
3. **竞争激烈** （MEV 机器人很多）
4. **建议先学习至少 3 个月**

---

## 📞 获取帮助

### 项目文档
- 技术方案：`TECH_PLAN.md`
- 开发日志：`SESSION_LOG.md`
- 项目结构：`PROJECT_STRUCTURE.md`
- 流程图：`ARCHITECTURE_DIAGRAMS.md`

### 学习资源
- **以太坊官方文档**：https://ethereum.org/zh/developers/
- **Solidity 教程**：https://solidity-by-example.org/
- **Uniswap 文档**：https://docs.uniswap.org/
- **Go 语言教程**：https://tour.golang.org/

### 遇到问题？
1. 先看文档的"常见问题"部分
2. 检查 GitHub Issues
3. 查看错误日志详细信息

---

## ⚠️ 重要提醒

### 关于测试网
- ✅ 测试网的币**没有价值**
- ✅ 可以随意实验，不会有损失
- ✅ 所有操作都是练习

### 关于主网
- ❌ 主网需要**真实的钱**
- ❌ 操作失误会**实际亏损**
- ❌ 竞争非常激烈
- ⚠️ **建议学习 3-6 个月后再考虑**

### 安全建议
- 🔐 **绝不**把私钥发给任何人
- 🔐 测试项目的私钥可以公开，但主网的不行
- 🔐 定期备份钱包
- 🔐 使用硬件钱包（如 Ledger）

---

## ✅ 检查清单

完成以下任务后，你就掌握了基础：

- [ ] 成功运行 `main.go` 查询余额
- [ ] 成功运行 `cmd/monitor/main.go` 看到新区块
- [ ] 成功运行 `forge test` 所有测试通过
- [ ] 理解什么是 AMM（阅读 `pkg/dex/amm.go`）
- [ ] 理解什么是套利（阅读 `pkg/strategy/arbitrage.go`）
- [ ] 理解什么是闪电贷（阅读 `contracts/src/FlashLoanArbitrage.sol`）
- [ ] 修改过至少一个参数并重新运行
- [ ] 阅读完 `TECH_PLAN.md`

**全部完成？恭喜！你已经入门了 🎉**

---

**版本**：v1.0  
**更新时间**：2026-02-03  
**难度**：⭐⭐☆☆☆（新手友好）
