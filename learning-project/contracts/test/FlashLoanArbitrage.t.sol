// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/FlashLoanArbitrage.sol";

/// @title FlashLoanArbitrageTest - Test suite for flash loan arbitrage
/// @title FlashLoanArbitrageTest - 闪电贷套利测试套件
contract FlashLoanArbitrageTest is Test {
    FlashLoanArbitrage public arbitrage;
    
    // Mock contracts / 模拟合约
    MockPoolAddressesProvider public addressesProvider;
    MockPool public pool;
    MockERC20 public tokenA;
    MockERC20 public tokenB;
    MockERC20 public tokenC;
    MockRouter public router1;
    MockRouter public router2;
    MockRouter public router3;
    
    address public owner = address(this);
    
    function setUp() public {
        // Deploy mock tokens / 部署模拟代币
        tokenA = new MockERC20("Token A", "TKA", 18);
        tokenB = new MockERC20("Token B", "TKB", 18);
        tokenC = new MockERC20("Token C", "TKC", 18);
        
        // Deploy mock routers / 部署模拟路由器
        router1 = new MockRouter();
        router2 = new MockRouter();
        router3 = new MockRouter();
        
        // Deploy mock Aave components / 部署模拟 Aave 组件
        pool = new MockPool();
        addressesProvider = new MockPoolAddressesProvider(address(pool));
        
        // Deploy arbitrage contract / 部署套利合约
        arbitrage = new FlashLoanArbitrage(address(addressesProvider));
        
        // Setup liquidity in mock routers / 在模拟路由器中设置流动性
        // Router 1: TKA -> TKB (rate: 1 TKA = 2 TKB)
        // 路由器1: TKA -> TKB (汇率: 1 TKA = 2 TKB)
        router1.setRate(address(tokenA), address(tokenB), 2 * 1e18);
        
        // Router 2: TKB -> TKC (rate: 1 TKB = 1.5 TKC)
        // 路由器2: TKB -> TKC (汇率: 1 TKB = 1.5 TKC)
        router2.setRate(address(tokenB), address(tokenC), 15 * 1e17);
        
        // Router 3: TKC -> TKA (rate: 1 TKC = 0.4 TKA) - Creates arbitrage opportunity
        // 路由器3: TKC -> TKA (汇率: 1 TKC = 0.4 TKA) - 创造套利机会
        // Path: 1 TKA -> 2 TKB -> 3 TKC -> 1.2 TKA (20% profit before fees)
        // 路径: 1 TKA -> 2 TKB -> 3 TKC -> 1.2 TKA (手续费前利润20%)
        router3.setRate(address(tokenC), address(tokenA), 4 * 1e17);
        
        // Fund mock pool with tokens for flash loans / 为模拟池注入代币用于闪电贷
        tokenA.mint(address(pool), 1000000 * 1e18);
        
        // Fund routers with liquidity / 为路由器注入流动性
        tokenA.mint(address(router1), 1000000 * 1e18);
        tokenB.mint(address(router1), 1000000 * 1e18);
        tokenB.mint(address(router2), 1000000 * 1e18);
        tokenC.mint(address(router2), 1000000 * 1e18);
        tokenC.mint(address(router3), 1000000 * 1e18);
        tokenA.mint(address(router3), 1000000 * 1e18);
        
        // Configure pool to use arbitrage contract / 配置池使用套利合约
        pool.setReceiver(address(arbitrage));
    }
    
    /// @notice Test successful flash loan arbitrage
    /// @notice 测试成功的闪电贷套利
    function testFlashLoanArbitrageSuccess() public {
        uint256 loanAmount = 100 * 1e18; // Borrow 100 TKA / 借入 100 TKA
        
        // Setup parameters / 设置参数
        address[3] memory routers = [
            address(router1),
            address(router2),
            address(router3)
        ];
        address[3] memory tokens = [
            address(tokenA),
            address(tokenB),
            address(tokenC)
        ];
        uint256 minProfitBps = 100; // 1% minimum profit / 最低1%利润
        
        // Record initial balance / 记录初始余额
        uint256 initialBalance = tokenA.balanceOf(address(arbitrage));
        
        // Execute flash loan arbitrage / 执行闪电贷套利
        arbitrage.executeFlashLoanArbitrage(
            address(tokenA),
            loanAmount,
            routers,
            tokens,
            minProfitBps
        );
        
        // Check profit / 检查利润
        uint256 finalBalance = tokenA.balanceOf(address(arbitrage));
        uint256 profit = finalBalance - initialBalance;
        
        // Expected flow:
        // 预期流程:
        // 1. Borrow 100 TKA / 借入 100 TKA
        // 2. Swap to 200 TKB / 兑换为 200 TKB
        // 3. Swap to 300 TKC / 兑换为 300 TKC
        // 4. Swap to 120 TKA / 兑换为 120 TKA
        // 5. Repay 100.09 TKA (0.09% premium) / 归还 100.09 TKA (0.09%手续费)
        // 6. Profit = 120 - 100.09 = 19.91 TKA / 利润 = 120 - 100.09 = 19.91 TKA
        
        uint256 premium = (loanAmount * 9) / 10000; // 0.09% / 0.09%
        uint256 expectedProfit = 20 * 1e18 - premium; // ~19.91 TKA / 约19.91 TKA
        
        assertGt(profit, 0, "Should have profit");
        assertApproxEqRel(profit, expectedProfit, 0.01e18, "Profit should match expected");
        
        console.log("Loan Amount:", loanAmount / 1e18, "TKA");
        console.log("Premium:", premium / 1e15, "milli-TKA");
        console.log("Profit:", profit / 1e18, "TKA");
        console.log("Profit %:", (profit * 10000) / loanAmount, "bps");
    }
    
    /// @notice Test flash loan arbitrage with insufficient profit
    /// @notice 测试利润不足的闪电贷套利
    function testFlashLoanArbitrageInsufficientProfit() public {
        // Modify router 3 to reduce profit / 修改路由器3降低利润
        // New rate: 1 TKC = 0.34 TKA (barely profitable)
        // 新汇率: 1 TKC = 0.34 TKA (几乎无利润)
        router3.setRate(address(tokenC), address(tokenA), 34 * 1e16);
        
        uint256 loanAmount = 100 * 1e18;
        
        address[3] memory routers = [
            address(router1),
            address(router2),
            address(router3)
        ];
        address[3] memory tokens = [
            address(tokenA),
            address(tokenB),
            address(tokenC)
        ];
        uint256 minProfitBps = 200; // Require 2% profit / 要求2%利润
        
        // Should revert due to insufficient profit / 应因利润不足而回滚
        vm.expectRevert("Profit below minimum");
        arbitrage.executeFlashLoanArbitrage(
            address(tokenA),
            loanAmount,
            routers,
            tokens,
            minProfitBps
        );
    }
    
    /// @notice Test flash loan arbitrage with loss (no profit)
    /// @notice 测试亏损的闪电贷套利（无利润）
    function testFlashLoanArbitrageLoss() public {
        // Modify router 3 to create loss / 修改路由器3造成亏损
        // New rate: 1 TKC = 0.3 TKA (loss scenario)
        // 新汇率: 1 TKC = 0.3 TKA (亏损场景)
        router3.setRate(address(tokenC), address(tokenA), 3 * 1e17);
        
        uint256 loanAmount = 100 * 1e18;
        
        address[3] memory routers = [
            address(router1),
            address(router2),
            address(router3)
        ];
        address[3] memory tokens = [
            address(tokenA),
            address(tokenB),
            address(tokenC)
        ];
        uint256 minProfitBps = 0;
        
        // Should revert due to no profit after repayment / 应因归还后无利润而回滚
        vm.expectRevert("No profit after loan repayment");
        arbitrage.executeFlashLoanArbitrage(
            address(tokenA),
            loanAmount,
            routers,
            tokens,
            minProfitBps
        );
    }
    
    /// @notice Test profit simulation (view function)
    /// @notice 测试利润模拟（视图函数）
    function testSimulateArbitrage() public view {
        uint256 loanAmount = 100 * 1e18;
        
        address[3] memory routers = [
            address(router1),
            address(router2),
            address(router3)
        ];
        address[3] memory tokens = [
            address(tokenA),
            address(tokenB),
            address(tokenC)
        ];
        
        (uint256 finalAmount, uint256 profit, uint256 premium, bool isProfitable) = 
            arbitrage.simulateArbitrage(routers, tokens, loanAmount, 9);
        
        assertTrue(isProfitable, "Should be profitable");
        assertGt(profit, 0, "Should have profit");
        assertEq(premium, (loanAmount * 9) / 10000, "Premium should be 0.09%");
        assertEq(finalAmount, 120 * 1e18, "Final amount should be 120 TKA");
        
        console.log("=== Simulation Results ===");
        console.log("Loan:", loanAmount / 1e18, "TKA");
        console.log("Final Amount:", finalAmount / 1e18, "TKA");
        console.log("Premium:", premium / 1e15, "milli-TKA");
        console.log("Profit:", profit / 1e18, "TKA");
        console.log("Profitable:", isProfitable);
    }
    
    /// @notice Test withdraw profit
    /// @notice 测试提取利润
    function testWithdrawProfit() public {
        // First execute arbitrage to generate profit / 先执行套利产生利润
        testFlashLoanArbitrageSuccess();
        
        uint256 contractBalance = tokenA.balanceOf(address(arbitrage));
        uint256 ownerBalanceBefore = tokenA.balanceOf(owner);
        
        // Withdraw profit / 提取利润
        arbitrage.withdrawProfit(address(tokenA));
        
        uint256 ownerBalanceAfter = tokenA.balanceOf(owner);
        
        assertEq(tokenA.balanceOf(address(arbitrage)), 0, "Contract balance should be zero");
        assertEq(ownerBalanceAfter - ownerBalanceBefore, contractBalance, "Owner should receive profit");
    }
}

// ============= Mock Contracts / 模拟合约 =============

/// @notice Mock ERC20 Token
/// @notice 模拟 ERC20 代币
contract MockERC20 {
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;
    
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    
    constructor(string memory _name, string memory _symbol, uint8 _decimals) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
    }
    
    function mint(address to, uint256 amount) external {
        balanceOf[to] += amount;
        totalSupply += amount;
    }
    
    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        return true;
    }
    
    function transfer(address to, uint256 amount) external returns (bool) {
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        allowance[from][msg.sender] -= amount;
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        return true;
    }
}

/// @notice Mock Uniswap V2 Router
/// @notice 模拟 Uniswap V2 路由器
contract MockRouter {
    // Exchange rates: tokenIn => tokenOut => rate (in 18 decimals)
    // 汇率: 输入代币 => 输出代币 => 汇率（18位小数）
    mapping(address => mapping(address => uint256)) public rates;
    
    function setRate(address tokenIn, address tokenOut, uint256 rate) external {
        rates[tokenIn][tokenOut] = rate;
    }
    
    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts) {
        require(deadline >= block.timestamp, "Expired");
        require(path.length == 2, "Invalid path");
        
        address tokenIn = path[0];
        address tokenOut = path[1];
        
        // Calculate output based on rate / 根据汇率计算输出
        uint256 rate = rates[tokenIn][tokenOut];
        require(rate > 0, "No rate set");
        
        uint256 amountOut = (amountIn * rate) / 1e18;
        require(amountOut >= amountOutMin, "Insufficient output");
        
        // Transfer tokens / 转账代币
        MockERC20(tokenIn).transferFrom(msg.sender, address(this), amountIn);
        MockERC20(tokenOut).transfer(to, amountOut);
        
        amounts = new uint[](2);
        amounts[0] = amountIn;
        amounts[1] = amountOut;
        
        return amounts;
    }
    
    function getAmountsOut(uint amountIn, address[] calldata path)
        external view returns (uint[] memory amounts)
    {
        require(path.length == 2, "Invalid path");
        
        address tokenIn = path[0];
        address tokenOut = path[1];
        uint256 rate = rates[tokenIn][tokenOut];
        
        uint256 amountOut = (amountIn * rate) / 1e18;
        
        amounts = new uint[](2);
        amounts[0] = amountIn;
        amounts[1] = amountOut;
        
        return amounts;
    }
}

/// @notice Mock Aave Pool
/// @notice 模拟 Aave 池
contract MockPool {
    address public receiver;
    uint256 public constant PREMIUM_BPS = 9; // 0.09% / 0.09%
    
    function setReceiver(address _receiver) external {
        receiver = _receiver;
    }
    
    function flashLoanSimple(
        address receiverAddress,
        address asset,
        uint256 amount,
        bytes calldata params,
        uint16 referralCode
    ) external {
        // Transfer loan to receiver / 将贷款转给接收者
        MockERC20(asset).transfer(receiverAddress, amount);
        
        // Call executeOperation callback / 调用 executeOperation 回调
        uint256 premium = (amount * PREMIUM_BPS) / 10000;
        
        bool success = FlashLoanArbitrage(payable(receiverAddress)).executeOperation(
            asset,
            amount,
            premium,
            receiverAddress,
            params
        );
        
        require(success, "Flash loan failed");
        
        // Pull back loan + premium / 收回贷款 + 手续费
        uint256 amountOwed = amount + premium;
        MockERC20(asset).transferFrom(receiverAddress, address(this), amountOwed);
    }
}

/// @notice Mock Aave Pool Addresses Provider
/// @notice 模拟 Aave 池地址提供者
contract MockPoolAddressesProvider {
    address public pool;
    
    constructor(address _pool) {
        pool = _pool;
    }
    
    function getPool() external view returns (address) {
        return pool;
    }
}
