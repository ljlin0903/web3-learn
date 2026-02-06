# 🚀 快速开始指南

## 📋 5分钟运行项目

### Step 1: 进入项目目录
```bash
cd /Users/ljlin/web3/mev-arbitrage-bot
```

### Step 2: 配置环境变量

已从学习项目复制 `.env` 文件，直接可用。

如需修改，编辑 `.env`:
```bash
vim .env
```

必填项:
- `RPC_HTTPS_URL` - Alchemy HTTPS URL
- `RPC_WSS_URL` - Alchemy WebSocket URL  
- `PRIVATE_KEY` - 你的钱包私钥
- `PUBLIC_ADDRESS` - 你的钱包地址

### Step 3: 运行程序

**方式 1: 直接运行**
```bash
go run cmd/bot/main.go
```

**方式 2: 编译后运行** (推荐)
```bash
# 已编译好，直接运行
./bin/mev-bot
```

### Step 4: 验证输出

应该看到:
```
╔══════════════════════════════════════════════════════════╗
║         🤖  MEV ARBITRAGE BOT  🤖                        ║
╚══════════════════════════════════════════════════════════╝

INFO[2024-02-06] Loading configuration...
INFO[2024-02-06] HTTPS RPC client connected
INFO[2024-02-06] ✅ All systems operational!
```

### Step 5: 停止程序
按 `Ctrl+C` 优雅退出

---

## 🔍 项目结构速览

```
mev-arbitrage-bot/
├── bin/mev-bot              # 编译好的可执行文件 ✅
├── cmd/bot/main.go          # 主程序入口 ✅
├── pkg/
│   ├── blockchain/          # 区块链客户端 ✅
│   ├── config/              # 配置管理 ✅
│   └── utils/               # 工具函数 ✅
├── .env                     # 环境配置 ✅
└── README.md                # 完整文档 ✅
```

---

## ✅ 当前功能

**已实现**:
- [x] 连接以太坊网络 (Sepolia测试网)
- [x] 查询账户余额
- [x] 获取实时 Gas 价格
- [x] 获取区块信息
- [x] 配置管理系统
- [x] 日志系统
- [x] 优雅关闭

**待实现**:
- [ ] DEX 池子监控
- [ ] 套利路径搜索
- [ ] Flashbots 集成
- [ ] 交易执行
- [ ] 智能合约

---

## 🎯 下一步

### 继续开发
查看 `PROJECT_STATUS.md` 了解:
- 详细进度
- 待开发模块
- 优先级排序
- 预计工作量

### 学习参考
`/Users/ljlin/web3/learning-project/` 目录包含完整的学习版本

---

## 🐛 常见问题

### Q: 编译失败
**A**: 运行 `go mod tidy` 下载依赖

### Q: 连接超时
**A**: 检查 `.env` 中的 RPC URL 是否正确

### Q: 余额为 0
**A**: 正常，测试网可在 [sepoliafaucet.com](https://sepoliafaucet.com/) 领取测试币

---

## 📞 获取帮助

- 查看 `README.md` - 完整文档
- 查看 `PROJECT_STATUS.md` - 开发进度
- 查看代码注释 - 实现细节

---

**更新时间**: 2024-02-06  
**状态**: 基础框架完成，可正常运行 ✅
