// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// Uniswap V2 Interface - Minimal interface for swapping
// Uniswap V2 接口 - 最小化交易接口
interface IUniswapV2Router {
    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);
}

// ERC20 Interface - Token operations
// ERC20 接口 - 代币操作
interface IERC20 {
    function approve(address spender, uint256 amount) external returns (bool);
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

/// @title FlashArbitrage - Triangle Arbitrage Contract
/// @title FlashArbitrage - 三角套利合约
/// @notice Executes atomic arbitrage across multiple DEX pools
/// @notice 在多个 DEX 池子间执行原子套利
/// @dev All trades must succeed or the entire transaction reverts
/// @dev 所有交易必须成功，否则整个交易回滚
contract FlashArbitrage {
    // Contract owner / 合约所有者
    address public owner;
    
    // Events / 事件
    event ArbitrageExecuted(
        address indexed token,
        uint256 amountIn,
        uint256 amountOut,
        uint256 profit
    );
    
    event ProfitWithdrawn(address indexed token, uint256 amount);
    
    // Modifiers / 修饰器
    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        // 非合约所有者
        _;
    }
    
    /// @notice Constructor - Set contract owner
    /// @notice 构造函数 - 设置合约所有者
    constructor() {
        owner = msg.sender;
    }
    
    /// @notice Execute triangle arbitrage
    /// @notice 执行三角套利
    /// @dev Performs 3 swaps atomically: tokenA -> tokenB -> tokenC -> tokenA
    /// @dev 原子化执行3次交易：代币A -> 代币B -> 代币C -> 代币A
    /// @param routers Array of 3 DEX router addresses
    /// @param routers 3个 DEX 路由器地址数组
    /// @param tokens Array of 3 token addresses [tokenA, tokenB, tokenC]
    /// @param tokens 3个代币地址数组 [代币A, 代币B, 代币C]
    /// @param amountIn Initial amount to trade
    /// @param amountIn 初始交易金额
    /// @param minProfitBps Minimum profit in basis points (100 = 1%)
    /// @param minProfitBps 最小利润率（基点，100 = 1%）
    function executeArbitrage(
        address[3] calldata routers,
        address[3] calldata tokens,
        uint256 amountIn,
        uint256 minProfitBps
    ) external onlyOwner returns (uint256 finalAmount) {
        // Record initial balance / 记录初始余额
        uint256 initialBalance = IERC20(tokens[0]).balanceOf(address(this));
        require(initialBalance >= amountIn, "Insufficient balance");
        // 余额不足
        
        // Step 1: Swap tokenA -> tokenB
        // 步骤1: 交易 代币A -> 代币B
        uint256 amountB = _swap(
            routers[0],
            tokens[0],
            tokens[1],
            amountIn
        );
        
        // Step 2: Swap tokenB -> tokenC
        // 步骤2: 交易 代币B -> 代币C
        uint256 amountC = _swap(
            routers[1],
            tokens[1],
            tokens[2],
            amountB
        );
        
        // Step 3: Swap tokenC -> tokenA (complete the loop)
        // 步骤3: 交易 代币C -> 代币A (闭环)
        finalAmount = _swap(
            routers[2],
            tokens[2],
            tokens[0],
            amountC
        );
        
        // Calculate profit / 计算利润
        require(finalAmount > amountIn, "No profit");
        // 无利润
        uint256 profit = finalAmount - amountIn;
        
        // Check minimum profit requirement / 检查最低利润要求
        uint256 minProfit = (amountIn * minProfitBps) / 10000;
        require(profit >= minProfit, "Profit below minimum");
        // 利润低于最低要求
        
        emit ArbitrageExecuted(tokens[0], amountIn, finalAmount, profit);
        
        return finalAmount;
    }
    
    /// @notice Internal swap function
    /// @notice 内部交易函数
    /// @dev Executes a single token swap through a DEX router
    /// @dev 通过 DEX 路由器执行单次代币交易
    function _swap(
        address router,
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) internal returns (uint256 amountOut) {
        // Approve router to spend tokens / 授权路由器使用代币
        IERC20(tokenIn).approve(router, amountIn);
        
        // Build swap path / 构建交易路径
        address[] memory path = new address[](2);
        path[0] = tokenIn;
        path[1] = tokenOut;
        
        // Execute swap / 执行交易
        uint256[] memory amounts = IUniswapV2Router(router).swapExactTokensForTokens(
            amountIn,
            0, // Accept any amount (can be optimized with slippage protection)
               // 接受任意数量（可优化为滑点保护）
            path,
            address(this),
            block.timestamp + 300 // 5 minute deadline / 5分钟有效期
        );
        
        return amounts[1];
    }
    
    /// @notice Withdraw profits
    /// @notice 提取利润
    /// @param token Token address to withdraw
    /// @param token 要提取的代币地址
    function withdrawProfit(address token) external onlyOwner {
        uint256 balance = IERC20(token).balanceOf(address(this));
        require(balance > 0, "No balance");
        // 无余额
        
        IERC20(token).transfer(owner, balance);
        emit ProfitWithdrawn(token, balance);
    }
    
    /// @notice Withdraw ETH (if any)
    /// @notice 提取 ETH（如有）
    function withdrawETH() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No ETH balance");
        // 无 ETH 余额
        
        payable(owner).transfer(balance);
    }
    
    /// @notice Receive ETH
    /// @notice 接收 ETH
    receive() external payable {}
}
