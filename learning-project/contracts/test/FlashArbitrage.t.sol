// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/FlashArbitrage.sol";

// Mock ERC20 Token for testing
// 模拟 ERC20 代币用于测试
contract MockERC20 {
    string public name;
    string public symbol;
    uint8 public decimals = 18;
    uint256 public totalSupply;
    
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    
    constructor(string memory _name, string memory _symbol, uint256 _supply) {
        name = _name;
        symbol = _symbol;
        totalSupply = _supply;
        balanceOf[msg.sender] = _supply;
    }
    
    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        return true;
    }
    
    function transfer(address to, uint256 amount) external returns (bool) {
        require(balanceOf[msg.sender] >= amount, "Insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        require(balanceOf[from] >= amount, "Insufficient balance");
        require(allowance[from][msg.sender] >= amount, "Insufficient allowance");
        
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        allowance[from][msg.sender] -= amount;
        return true;
    }
    
    function mint(address to, uint256 amount) external {
        totalSupply += amount;
        balanceOf[to] += amount;
    }
}

// Mock Uniswap V2 Router for testing
// 模拟 Uniswap V2 路由器用于测试
contract MockRouter {
    // Simulate exchange rate / 模拟汇率
    uint256 public rate; // in basis points, 10000 = 1:1
    
    constructor(uint256 _rate) {
        rate = _rate;
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
        
        MockERC20 tokenIn = MockERC20(path[0]);
        MockERC20 tokenOut = MockERC20(path[1]);
        
        // Transfer tokens in
        tokenIn.transferFrom(msg.sender, address(this), amountIn);
        
        // Calculate output (applying rate)
        uint256 amountOut = (amountIn * rate) / 10000;
        require(amountOut >= amountOutMin, "Insufficient output");
        
        // Transfer tokens out
        tokenOut.transfer(to, amountOut);
        
        amounts = new uint[](2);
        amounts[0] = amountIn;
        amounts[1] = amountOut;
        
        return amounts;
    }
}

/// @title FlashArbitrageTest - Test suite for FlashArbitrage contract
/// @title FlashArbitrageTest - FlashArbitrage 合约测试套件
contract FlashArbitrageTest is Test {
    FlashArbitrage public arbitrage;
    
    MockERC20 public tokenA;
    MockERC20 public tokenB;
    MockERC20 public tokenC;
    
    MockRouter public router1;
    MockRouter public router2;
    MockRouter public router3;
    
    address public owner;
    
    function setUp() public {
        owner = address(this);
        
        // Deploy arbitrage contract / 部署套利合约
        arbitrage = new FlashArbitrage();
        
        // Deploy mock tokens / 部署模拟代币
        tokenA = new MockERC20("Token A", "TKA", 1000000 ether);
        tokenB = new MockERC20("Token B", "TKB", 1000000 ether);
        tokenC = new MockERC20("Token C", "TKC", 1000000 ether);
        
        // Deploy mock routers with different rates / 部署不同汇率的模拟路由器
        // Router 1: 1 TKA = 1.01 TKB (1% premium)
        router1 = new MockRouter(10100);
        
        // Router 2: 1 TKB = 1.02 TKC (2% premium)
        router2 = new MockRouter(10200);
        
        // Router 3: 1 TKC = 1.01 TKA (1% premium, creates arbitrage)
        router3 = new MockRouter(10100);
        
        // Fund routers with tokens / 为路由器提供代币资金
        tokenB.transfer(address(router1), 100000 ether);
        tokenC.transfer(address(router2), 100000 ether);
        tokenA.transfer(address(router3), 100000 ether);
        
        // Fund arbitrage contract / 为套利合约提供资金
        tokenA.transfer(address(arbitrage), 10 ether);
    }
    
    /// @notice Test successful arbitrage execution
    /// @notice 测试成功的套利执行
    function testSuccessfulArbitrage() public {
        uint256 initialBalance = tokenA.balanceOf(address(arbitrage));
        uint256 amountIn = 1 ether;
        
        // Prepare parameters / 准备参数
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
        
        // Execute arbitrage with 1% minimum profit / 执行套利，最低利润1%
        uint256 finalAmount = arbitrage.executeArbitrage(
            routers,
            tokens,
            amountIn,
            100 // 1% minimum profit
        );
        
        // Verify profit / 验证利润
        assertTrue(finalAmount > amountIn, "Should have profit");
        
        uint256 finalBalance = tokenA.balanceOf(address(arbitrage));
        uint256 expectedBalance = initialBalance - amountIn + finalAmount;
        assertEq(finalBalance, expectedBalance, "Balance should update correctly");
        
        console.log("Amount In:", amountIn);
        console.log("Amount Out:", finalAmount);
        console.log("Profit:", finalAmount - amountIn);
        console.log("Profit %:", ((finalAmount - amountIn) * 10000) / amountIn, "bps");
    }
    
    /// @notice Test arbitrage with insufficient profit
    /// @notice 测试利润不足的套利
    function testInsufficientProfit() public {
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
        
        // Expect revert when requiring 10% profit (won't achieve)
        // 期望在要求10%利润时回滚（无法达到）
        vm.expectRevert("Profit below minimum");
        arbitrage.executeArbitrage(
            routers,
            tokens,
            1 ether,
            1000 // 10% minimum profit (unrealistic)
        );
    }
    
    /// @notice Test profit withdrawal
    /// @notice 测试利润提取
    function testWithdrawProfit() public {
        uint256 contractBalance = tokenA.balanceOf(address(arbitrage));
        uint256 ownerBalanceBefore = tokenA.balanceOf(owner);
        
        arbitrage.withdrawProfit(address(tokenA));
        
        uint256 ownerBalanceAfter = tokenA.balanceOf(owner);
        assertEq(ownerBalanceAfter - ownerBalanceBefore, contractBalance, "Should withdraw all");
        assertEq(tokenA.balanceOf(address(arbitrage)), 0, "Contract should be empty");
    }
}
