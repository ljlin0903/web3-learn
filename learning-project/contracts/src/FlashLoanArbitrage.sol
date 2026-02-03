// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// Aave V3 Flash Loan Interface
// Aave V3 闪电贷接口
interface IPoolAddressesProvider {
    function getPool() external view returns (address);
}

interface IPool {
    function flashLoanSimple(
        address receiverAddress,
        address asset,
        uint256 amount,
        bytes calldata params,
        uint16 referralCode
    ) external;
}

// Uniswap V2 Router Interface
// Uniswap V2 路由器接口
interface IUniswapV2Router {
    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);
    
    function getAmountsOut(uint amountIn, address[] calldata path)
        external view returns (uint[] memory amounts);
}

// ERC20 Interface
// ERC20 接口
interface IERC20 {
    function approve(address spender, uint256 amount) external returns (bool);
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

/// @title FlashLoanArbitrage - Flash Loan Triangle Arbitrage Contract
/// @title FlashLoanArbitrage - 闪电贷三角套利合约
/// @notice Executes arbitrage using Aave flash loans (no upfront capital needed)
/// @notice 使用 Aave 闪电贷执行套利（无需前期资金）
/// @dev Borrows funds, executes arbitrage, repays loan + fee in a single transaction
/// @dev 借入资金、执行套利、归还贷款+手续费，全部在单笔交易中完成
contract FlashLoanArbitrage {
    // Contract owner / 合约所有者
    address public owner;
    
    // Aave Pool Addresses Provider / Aave 池地址提供者
    IPoolAddressesProvider public immutable ADDRESSES_PROVIDER;
    IPool public immutable POOL;
    
    // Events / 事件
    event ArbitrageExecuted(
        address indexed token,
        uint256 loanAmount,
        uint256 profit,
        uint256 premium
    );
    
    event ProfitWithdrawn(address indexed token, uint256 amount);
    
    // Modifiers / 修饰器
    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }
    
    /// @notice Constructor
    /// @notice 构造函数
    /// @param _addressProvider Aave PoolAddressesProvider address
    /// @param _addressProvider Aave 池地址提供者地址
    constructor(address _addressProvider) {
        owner = msg.sender;
        ADDRESSES_PROVIDER = IPoolAddressesProvider(_addressProvider);
        POOL = IPool(ADDRESSES_PROVIDER.getPool());
    }
    
    /// @notice Execute flash loan arbitrage
    /// @notice 执行闪电贷套利
    /// @dev Initiates flash loan, arbitrage executed in executeOperation callback
    /// @dev 发起闪电贷，套利在 executeOperation 回调中执行
    /// @param asset Token to borrow (e.g., WETH)
    /// @param asset 要借入的代币（例如 WETH）
    /// @param loanAmount Amount to borrow
    /// @param loanAmount 借入金额
    /// @param routers Array of 3 DEX router addresses
    /// @param routers 3个 DEX 路由器地址数组
    /// @param tokens Array of 3 token addresses for arbitrage path
    /// @param tokens 套利路径的3个代币地址数组
    /// @param minProfitBps Minimum profit in basis points (100 = 1%)
    /// @param minProfitBps 最小利润率（基点，100 = 1%）
    function executeFlashLoanArbitrage(
        address asset,
        uint256 loanAmount,
        address[3] calldata routers,
        address[3] calldata tokens,
        uint256 minProfitBps
    ) external onlyOwner {
        // Ensure first token matches borrowed asset
        // 确保第一个代币与借入资产匹配
        require(tokens[0] == asset, "First token must match borrowed asset");
        
        // Encode parameters for callback
        // 编码回调参数
        bytes memory params = abi.encode(routers, tokens, minProfitBps);
        
        // Initiate flash loan
        // 发起闪电贷
        POOL.flashLoanSimple(
            address(this),
            asset,
            loanAmount,
            params,
            0 // referral code
        );
    }
    
    /// @notice Flash loan callback - Aave calls this after transferring loan
    /// @notice 闪电贷回调 - Aave 在转账借款后调用此函数
    /// @dev MUST repay loan + premium before returning, or transaction reverts
    /// @dev 必须在返回前归还贷款+手续费，否则交易回滚
    /// @param asset Borrowed token address
    /// @param asset 借入代币地址
    /// @param amount Loan amount
    /// @param amount 借款金额
    /// @param premium Loan fee (typically 0.09%)
    /// @param premium 借款手续费（通常为 0.09%）
    /// @param initiator Address that initiated the flash loan
    /// @param initiator 发起闪电贷的地址
    /// @param params Encoded arbitrage parameters
    /// @param params 编码的套利参数
    function executeOperation(
        address asset,
        uint256 amount,
        uint256 premium,
        address initiator,
        bytes calldata params
    ) external returns (bool) {
        // Security: Only Aave Pool can call this
        // 安全检查：仅 Aave 池可调用
        require(msg.sender == address(POOL), "Caller must be Pool");
        require(initiator == address(this), "Initiator must be this contract");
        
        // Decode parameters
        // 解码参数
        (address[3] memory routers, address[3] memory tokens, uint256 minProfitBps) = 
            abi.decode(params, (address[3], address[3], uint256));
        
        // Execute triangle arbitrage
        // 执行三角套利
        uint256 finalAmount = _executeArbitrage(routers, tokens, amount);
        
        // Calculate total amount owed (loan + premium)
        // 计算应还总额（贷款 + 手续费）
        uint256 amountOwed = amount + premium;
        
        // Verify profit after repaying loan
        // 验证归还贷款后的利润
        require(finalAmount > amountOwed, "No profit after loan repayment");
        uint256 profit = finalAmount - amountOwed;
        
        // Check minimum profit requirement
        // 检查最低利润要求
        uint256 minProfit = (amount * minProfitBps) / 10000;
        require(profit >= minProfit, "Profit below minimum");
        
        // Approve Pool to pull the owed amount
        // 授权池扣除应还金额
        IERC20(asset).approve(address(POOL), amountOwed);
        
        emit ArbitrageExecuted(asset, amount, profit, premium);
        
        return true;
    }
    
    /// @notice Execute triangle arbitrage
    /// @notice 执行三角套利
    /// @dev Internal function to perform 3-step arbitrage
    /// @dev 执行3步套利的内部函数
    function _executeArbitrage(
        address[3] memory routers,
        address[3] memory tokens,
        uint256 amountIn
    ) internal returns (uint256 finalAmount) {
        // Step 1: Swap tokenA -> tokenB
        // 步骤1: 交易 代币A -> 代币B
        uint256 amountB = _swap(routers[0], tokens[0], tokens[1], amountIn);
        
        // Step 2: Swap tokenB -> tokenC
        // 步骤2: 交易 代币B -> 代币C
        uint256 amountC = _swap(routers[1], tokens[1], tokens[2], amountB);
        
        // Step 3: Swap tokenC -> tokenA (complete the loop)
        // 步骤3: 交易 代币C -> 代币A (闭环)
        finalAmount = _swap(routers[2], tokens[2], tokens[0], amountC);
        
        return finalAmount;
    }
    
    /// @notice Internal swap function
    /// @notice 内部交易函数
    function _swap(
        address router,
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) internal returns (uint256 amountOut) {
        // Approve router to spend tokens
        // 授权路由器使用代币
        IERC20(tokenIn).approve(router, amountIn);
        
        // Build swap path
        // 构建交易路径
        address[] memory path = new address[](2);
        path[0] = tokenIn;
        path[1] = tokenOut;
        
        // Execute swap
        // 执行交易
        uint256[] memory amounts = IUniswapV2Router(router).swapExactTokensForTokens(
            amountIn,
            0, // Accept any amount (production should use slippage protection)
               // 接受任意数量（生产环境应使用滑点保护）
            path,
            address(this),
            block.timestamp + 300 // 5 minute deadline / 5分钟有效期
        );
        
        return amounts[1];
    }
    
    /// @notice Simulate arbitrage profit (off-chain view function)
    /// @notice 模拟套利利润（链下视图函数）
    /// @dev Use this to check profitability before executing flash loan
    /// @dev 在执行闪电贷前使用此函数检查盈利性
    function simulateArbitrage(
        address[3] calldata routers,
        address[3] calldata tokens,
        uint256 amountIn,
        uint256 premiumBps // Aave premium in basis points (e.g., 9 for 0.09%)
                          // Aave 手续费基点（例如 9 表示 0.09%）
    ) external view returns (
        uint256 finalAmount,
        uint256 profit,
        uint256 premium,
        bool isProfitable
    ) {
        // Simulate 3 swaps
        // 模拟3次交易
        address[] memory path1 = new address[](2);
        path1[0] = tokens[0];
        path1[1] = tokens[1];
        uint256[] memory amounts1 = IUniswapV2Router(routers[0]).getAmountsOut(amountIn, path1);
        
        address[] memory path2 = new address[](2);
        path2[0] = tokens[1];
        path2[1] = tokens[2];
        uint256[] memory amounts2 = IUniswapV2Router(routers[1]).getAmountsOut(amounts1[1], path2);
        
        address[] memory path3 = new address[](2);
        path3[0] = tokens[2];
        path3[1] = tokens[0];
        uint256[] memory amounts3 = IUniswapV2Router(routers[2]).getAmountsOut(amounts2[1], path3);
        
        finalAmount = amounts3[1];
        premium = (amountIn * premiumBps) / 10000;
        uint256 amountOwed = amountIn + premium;
        
        if (finalAmount > amountOwed) {
            profit = finalAmount - amountOwed;
            isProfitable = true;
        } else {
            profit = 0;
            isProfitable = false;
        }
        
        return (finalAmount, profit, premium, isProfitable);
    }
    
    /// @notice Withdraw profits
    /// @notice 提取利润
    function withdrawProfit(address token) external onlyOwner {
        uint256 balance = IERC20(token).balanceOf(address(this));
        require(balance > 0, "No balance");
        
        IERC20(token).transfer(owner, balance);
        emit ProfitWithdrawn(token, balance);
    }
    
    /// @notice Withdraw ETH (if any)
    /// @notice 提取 ETH（如有）
    function withdrawETH() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No ETH balance");
        
        payable(owner).transfer(balance);
    }
    
    /// @notice Receive ETH
    /// @notice 接收 ETH
    receive() external payable {}
}
