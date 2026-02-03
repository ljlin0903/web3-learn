# Web3 量化交易机器人 - 系统流程图

## 1. 整体系统架构

```mermaid
graph TB
    subgraph "数据源层 Data Source"
        A[Ethereum Mainnet/Sepolia]
        B[Alchemy RPC Node]
        A -->|WebSocket/HTTPS| B
    end
    
    subgraph "数据采集层 Data Collection"
        C[Block Monitor<br/>monitor_reconnect.go]
        D[Pool Monitor<br/>pool_monitor.go]
        E[Price Calculator<br/>amm_calculator.go]
        
        B -->|newHeads Event| C
        B -->|logs Event Swap| D
        B -->|getAmountsOut| E
    end
    
    subgraph "策略层 Strategy Layer"
        F[Arbitrage Finder<br/>arbitrage_finder.go]
        G[Flash Loan Monitor<br/>flashloan_monitor.go]
        
        C -->|Block Data| F
        D -->|Pool Reserves| F
        E -->|Price Data| F
        F -->|Opportunities| G
    end
    
    subgraph "执行层 Execution Layer"
        H[Flash Arbitrage Contract<br/>FlashArbitrage.sol]
        I[Flash Loan Contract<br/>FlashLoanArbitrage.sol]
        
        G -->|Need Capital| H
        G -->|No Capital| I
    end
    
    subgraph "链上协议 On-Chain Protocols"
        J[Uniswap V2/V3]
        K[SushiSwap]
        L[Aave V3]
        
        H -->|Swap| J
        H -->|Swap| K
        I -->|Flash Loan| L
        I -->|Swap| J
        I -->|Swap| K
    end
```

## 2. 闪电贷套利详细流程

```mermaid
sequenceDiagram
    participant User as 用户/机器人
    participant Monitor as flashloan_monitor.go
    participant Contract as FlashLoanArbitrage
    participant Aave as Aave Pool
    participant Uni as Uniswap
    participant Sushi as SushiSwap
    
    User->>Monitor: 启动监控
    
    loop 每5秒扫描
        Monitor->>Monitor: 扫描套利机会
        Monitor->>Contract: simulateArbitrage()<br/>(链下模拟)
        Contract-->>Monitor: 预期利润 19.91%
        
        alt 利润 > 阈值
            Monitor->>Contract: executeFlashLoanArbitrage()<br/>(发起交易)
            
            Note over Contract,Aave: 单笔交易内的原子执行
            Contract->>Aave: flashLoanSimple()<br/>借入 100 ETH
            Aave->>Contract: 转账 100 ETH
            
            Contract->>Contract: executeOperation()<br/>(回调函数)
            
            Contract->>Uni: swap(100 ETH → 200 USDC)
            Uni-->>Contract: 200 USDC
            
            Contract->>Sushi: swap(200 USDC → 300 DAI)
            Sushi-->>Contract: 300 DAI
            
            Contract->>Uni: swap(300 DAI → 120 ETH)
            Uni-->>Contract: 120 ETH
            
            Contract->>Contract: 检查利润<br/>120 - 100.09 = 19.91 ETH
            
            alt 利润足够
                Contract->>Aave: 归还 100.09 ETH<br/>(本金 + 0.09% 手续费)
                Aave-->>Contract: 确认归还
                Contract->>User: 保留利润 19.91 ETH
                Monitor-->>User: ✅ 套利成功
            else 利润不足
                Contract->>Contract: revert<br/>(整个交易回滚)
                Monitor-->>User: ❌ 交易回滚<br/>(不消耗 Gas)
            end
        else 无利润机会
            Monitor->>Monitor: 继续扫描
        end
    end
```

## 3. 代码文件依赖关系

```mermaid
graph LR
    subgraph "入口程序 Entry Points"
        M1[main.go<br/>基础测试]
        M2[monitor_reconnect.go<br/>区块监听]
        M3[flashloan_monitor.go<br/>套利执行]
    end
    
    subgraph "核心库 Core Libraries"
        L1[blockchain/client.go<br/>RPC连接]
        L2[dex/pool.go<br/>池子监控]
        L3[dex/amm.go<br/>价格计算]
        L4[strategy/arbitrage.go<br/>套利搜索]
    end
    
    subgraph "智能合约 Smart Contracts"
        C1[FlashArbitrage.sol<br/>普通套利]
        C2[FlashLoanArbitrage.sol<br/>闪电贷套利]
    end
    
    subgraph "测试套件 Test Suites"
        T1[FlashArbitrage.t.sol]
        T2[FlashLoanArbitrage.t.sol]
    end
    
    M1 --> L1
    M2 --> L1
    M2 --> L2
    M3 --> L1
    M3 --> L2
    M3 --> L3
    M3 --> L4
    M3 --> C2
    
    L2 --> L1
    L3 --> L1
    L4 --> L2
    L4 --> L3
    
    C1 --> T1
    C2 --> T2
```

## 4. Docker 部署流程

```mermaid
graph TB
    subgraph "本地开发 Local Development"
        D1[编写 Go 代码]
        D2[编写 Solidity 合约]
        D3[运行测试]
        
        D1 --> D3
        D2 --> D3
    end
    
    subgraph "Docker 构建 Build"
        B1[Dockerfile<br/>多阶段构建]
        B2[Stage 1: Go Builder<br/>golang:1.21-alpine]
        B3[Stage 2: Runtime<br/>alpine:latest]
        
        D3 --> B1
        B1 --> B2
        B2 -->|编译二进制| B3
    end
    
    subgraph "镜像仓库 Registry"
        R1[Docker Hub]
        R2[阿里云镜像仓库]
        
        B3 --> R1
        B3 --> R2
    end
    
    subgraph "服务器部署 Server Deployment"
        S1[deploy_expect.sh<br/>部署脚本]
        S2[rsync 上传文件]
        S3[docker build 构建]
        S4[docker run 启动]
        
        R2 --> S1
        S1 --> S2
        S2 --> S3
        S3 --> S4
    end
    
    subgraph "运行监控 Monitoring"
        M1[docker logs 查看日志]
        M2[docker ps 检查状态]
        M3[Prometheus 性能监控]
        
        S4 --> M1
        S4 --> M2
        S4 --> M3
    end
```

## 5. 合约部署与测试流程

```mermaid
graph TB
    subgraph "开发阶段 Development"
        C1[编写 Solidity 合约]
        C2[编写 Foundry 测试]
        C3[本地测试<br/>forge test]
        
        C1 --> C2
        C2 --> C3
    end
    
    subgraph "测试网部署 Testnet Deployment"
        T1[部署到 Sepolia]
        T2[验证合约<br/>forge verify]
        T3[调用测试<br/>cast call]
        
        C3 -->|测试通过| T1
        T1 --> T2
        T2 --> T3
    end
    
    subgraph "审计阶段 Audit"
        A1[代码审计]
        A2[安全测试]
        A3[Gas 优化]
        
        T3 --> A1
        A1 --> A2
        A2 --> A3
    end
    
    subgraph "主网部署 Mainnet Deployment"
        M1[部署到 Mainnet]
        M2[验证合约]
        M3[转入启动资金]
        M4[授权合约操作]
        M5[启动套利机器人]
        
        A3 -->|审计通过| M1
        M1 --> M2
        M2 --> M3
        M3 --> M4
        M4 --> M5
    end
```

## 6. Go 程序执行流程

```mermaid
stateDiagram-v2
    [*] --> 初始化
    初始化 --> 加载配置: 读取.env文件
    加载配置 --> 连接RPC: 建立WebSocket连接
    连接RPC --> 订阅事件: newHeads + logs
    订阅事件 --> 监听中
    
    监听中 --> 接收区块: 新区块到达
    接收区块 --> 解析数据: 提取Swap事件
    解析数据 --> 更新池子: 更新储备量
    更新池子 --> 计算套利: 搜索路径
    
    计算套利 --> 发现机会: 利润 > 阈值
    计算套利 --> 监听中: 无机会
    
    发现机会 --> 链下模拟: simulateArbitrage()
    链下模拟 --> 确认盈利: 扣除Gas后仍盈利
    链下模拟 --> 监听中: 模拟亏损
    
    确认盈利 --> 发送交易: executeFlashLoanArbitrage()
    发送交易 --> 等待确认: 监控交易状态
    
    等待确认 --> 交易成功: 记录利润
    等待确认 --> 交易失败: 记录错误
    
    交易成功 --> 监听中
    交易失败 --> 监听中
    
    监听中 --> 连接断开: 网络异常
    连接断开 --> 重连机制: 指数退避
    重连机制 --> 连接RPC: 重新连接
    
    监听中 --> [*]: 用户停止
```

## 7. 数据流转详图

```mermaid
flowchart LR
    subgraph "数据源"
        A1[Ethereum 区块]
        A2[Uniswap 池子]
        A3[SushiSwap 池子]
    end
    
    subgraph "数据采集"
        B1[WebSocket 监听]
        B2[Event Log 解析]
        B3[Pool Reserves 查询]
    end
    
    subgraph "数据处理"
        C1[价格计算<br/>AMM公式]
        C2[路径搜索<br/>图算法]
        C3[利润计算<br/>扣除费用]
    end
    
    subgraph "决策引擎"
        D1[盈利性判断]
        D2[Gas价格评估]
        D3[滑点计算]
    end
    
    subgraph "执行引擎"
        E1[构造交易]
        E2[签名交易]
        E3[广播交易]
    end
    
    A1 --> B1
    A2 --> B2
    A3 --> B2
    
    B1 --> B3
    B2 --> C1
    B3 --> C1
    
    C1 --> C2
    C2 --> C3
    
    C3 --> D1
    D1 --> D2
    D2 --> D3
    
    D3 -->|盈利| E1
    D3 -->|不盈利| B1
    
    E1 --> E2
    E2 --> E3
    E3 --> A1
```

## 8. 错误处理流程

```mermaid
graph TD
    A[程序运行] --> B{遇到错误?}
    
    B -->|网络错误| C[重连机制]
    C --> C1[指数退避]
    C1 --> C2[最多重试5次]
    C2 -->|成功| A
    C2 -->|失败| Z[记录日志并告警]
    
    B -->|RPC错误| D[切换节点]
    D --> D1[尝试备用RPC]
    D1 -->|成功| A
    D1 -->|失败| Z
    
    B -->|交易失败| E[分析原因]
    E --> E1{失败类型}
    E1 -->|Gas不足| E2[提高Gas价格]
    E1 -->|滑点过大| E3[增加滑点容忍]
    E1 -->|利润不足| E4[回滚,不消耗Gas]
    E2 --> A
    E3 --> A
    E4 --> A
    
    B -->|数据异常| F[数据验证]
    F --> F1[丢弃异常数据]
    F1 --> A
    
    B -->|合约错误| G[链上检查]
    G --> G1{合约状态}
    G1 -->|暂停| G2[等待恢复]
    G1 -->|正常| G3[重新部署]
    G2 --> A
    G3 --> A
    
    B -->|无错误| H[继续监控]
    H --> A
```

---

## 📋 文件清单说明

### **Go 源文件作用**

| 文件 | 行数 | 核心功能 | 依赖 |
|-----|------|---------|------|
| `main.go` | 80 | 基础测试：连接RPC，查余额 | go-ethereum, godotenv |
| `monitor.go` | 100 | 简单区块监听器 | go-ethereum |
| `monitor_reconnect.go` | 150 | 生产级监听器（重连+心跳） | go-ethereum |
| `pool_monitor.go` | 180 | Uniswap Swap事件监听 | go-ethereum, abi |
| `amm_calculator.go` | 200 | AMM价格计算（Uniswap公式） | math/big |
| `arbitrage_finder.go` | 250 | 三角套利路径搜索 | math/big |
| `flashloan_monitor.go` | 300 | 闪电贷套利监控器 | go-ethereum, abi, bind |

### **Solidity 合约作用**

| 文件 | 行数 | 核心功能 | 外部依赖 |
|-----|------|---------|---------|
| `FlashArbitrage.sol` | 180 | 普通套利（需自有资金） | Uniswap Router |
| `FlashLoanArbitrage.sol` | 308 | 闪电贷套利（无需资金） | Aave V3 Pool |
| `FlashArbitrage.t.sol` | 250 | 套利合约测试 | Foundry Test |
| `FlashLoanArbitrage.t.sol` | 400 | 闪电贷合约测试 | Foundry Test |

### **部署脚本作用**

| 文件 | 行数 | 用途 | 依赖工具 |
|-----|------|------|---------|
| `deploy.sh` | 120 | SSH自动部署（sshpass） | sshpass, rsync |
| `deploy_expect.sh` | 85 | SSH自动部署（expect） | expect, rsync |
| `Dockerfile` | 40 | Docker镜像构建 | Docker |
| `docker-compose.yml` | 20 | 服务编排 | Docker Compose |

---

**版本**: v1.0  
**生成时间**: 2026-02-03  
**格式**: Mermaid Diagram
