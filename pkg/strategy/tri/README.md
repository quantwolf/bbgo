# Triangular Arbitrage Strategy

## Overview

The Triangular Arbitrage (tri) strategy is a sophisticated arbitrage strategy that exploits price inefficiencies across three related trading pairs on the same exchange. It automatically detects profitable triangular arbitrage opportunities and executes three sequential trades to capture risk-free profits from temporary price discrepancies between currency pairs.

## How It Works

1. **Path Detection**: Identifies triangular paths between three related currency pairs
2. **Rate Calculation**: Continuously calculates forward and backward arbitrage ratios
3. **Opportunity Identification**: Detects when arbitrage ratios exceed minimum profit thresholds
4. **Sequential Execution**: Executes three trades in sequence using IOC (Immediate-or-Cancel) orders
5. **Position Management**: Tracks multi-currency positions and calculates profits
6. **Risk Control**: Implements balance limits and protective order pricing

## Key Features

- **Automatic Path Discovery**: Automatically determines trading directions for triangular paths
- **Real-Time Monitoring**: Continuously monitors order books for arbitrage opportunities
- **IOC Order Strategy**: Uses Immediate-or-Cancel orders to minimize execution risk
- **Multi-Currency Position Tracking**: Tracks positions across multiple currencies
- **Protective Pricing**: Applies protective ratios to market orders to prevent slippage
- **Balance Management**: Implements balance limits and buffers for risk control
- **Performance Analytics**: Tracks IOC winning ratios and trade statistics
- **Separate Streams**: Optional separate WebSocket streams for better performance

## Strategy Logic

### Triangular Path Structure

<augment_code_snippet path="pkg/strategy/tri/path.go" mode="EXCERPT">
```go
type Path struct {
    marketA, marketB, marketC *ArbMarket
    dirA, dirB, dirC          int
}

func (p *Path) solveDirection() error {
    // Automatically determine trading directions based on currency relationships
    if p.marketA.QuoteCurrency == p.marketB.BaseCurrency || p.marketA.QuoteCurrency == p.marketB.QuoteCurrency {
        p.dirA = 1  // Sell direction
    } else if p.marketA.BaseCurrency == p.marketB.BaseCurrency || p.marketA.BaseCurrency == p.marketB.QuoteCurrency {
        p.dirA = -1 // Buy direction
    }
    // Similar logic for marketB and marketC...
}
```
</augment_code_snippet>

### Arbitrage Ratio Calculation

<augment_code_snippet path="pkg/strategy/tri/strategy.go" mode="EXCERPT">
```go
// Forward arbitrage: A -> B -> C -> A
func calculateForwardRatio(p *Path) float64 {
    var ratio = 1.0
    ratio *= p.marketA.calculateRatio(p.dirA)
    ratio *= p.marketB.calculateRatio(p.dirB)
    ratio *= p.marketC.calculateRatio(p.dirC)
    return ratio
}

// Backward arbitrage: A <- B <- C <- A
func calculateBackwardRate(p *Path) float64 {
    var ratio = 1.0
    ratio *= p.marketA.calculateRatio(-p.dirA)
    ratio *= p.marketB.calculateRatio(-p.dirB)
    ratio *= p.marketC.calculateRatio(-p.dirC)
    return ratio
}
```
</augment_code_snippet>

### IOC Order Execution

<augment_code_snippet path="pkg/strategy/tri/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) iocOrderExecution(
    ctx context.Context, session *bbgo.ExchangeSession, orders [3]types.SubmitOrder, ratio float64,
) (types.OrderSlice, error) {
    // Execute first order as IOC
    orders[0].Type = types.OrderTypeLimit
    orders[0].TimeInForce = types.TimeInForceIOC
    
    iocOrder := s.executeOrder(ctx, orders[0])
    if iocOrder == nil {
        return nil, errors.New("ioc order submit error")
    }
    
    // Wait for IOC order completion
    o := <-iocOrderC
    filledQuantity := o.ExecutedQuantity
    
    if filledQuantity.IsZero() {
        s.State.IOCLossTimes++
        return nil, nil
    }
    
    // Adjust subsequent orders based on filled quantity
    filledRatio := filledQuantity.Div(iocOrder.Quantity)
    orders[1].Quantity = orders[1].Quantity.Mul(filledRatio)
    orders[2].Quantity = orders[2].Quantity.Mul(filledRatio)
    
    // Execute remaining orders as market orders with protective pricing
    orders[1] = s.toProtectiveMarketOrder(orders[1], s.MarketOrderProtectiveRatio)
    orders[2] = s.toProtectiveMarketOrder(orders[2], s.MarketOrderProtectiveRatio)
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbols` | array | No | List of trading symbols (auto-detected from paths if not specified) |
| `paths` | array | Yes | Array of triangular paths, each containing 3 symbols |
| `minSpreadRatio` | number | No | Minimum profit ratio required (default: 1.002 = 0.2%) |

### Execution Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `separateStream` | boolean | No | Use separate WebSocket streams for each symbol |
| `marketOrderProtectiveRatio` | number | No | Protective ratio for market orders (default: 0.008) |
| `iocOrderRatio` | number | No | Protective ratio for IOC orders |
| `coolingDownTime` | duration | No | Cooldown period between arbitrage executions |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limits` | object | No | Maximum balance limits per currency |
| `resetPosition` | boolean | No | Reset position tracking on strategy start |
| `dryRun` | boolean | No | Enable dry run mode (no actual orders) |

### Monitoring
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `notifyTrade` | boolean | No | Send notifications for executed trades |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  tri:
    # Minimum profit threshold
    minSpreadRatio: 1.0015        # 0.15% minimum profit
    
    # Performance optimization
    separateStream: true          # Use separate streams for better performance
    
    # Risk management
    limits:
      BTC: 0.01                   # Maximum 0.01 BTC exposure
      ETH: 0.1                    # Maximum 0.1 ETH exposure
      USDT: 1000.0                # Maximum $1000 USDT exposure
    
    # Execution settings
    marketOrderProtectiveRatio: 0.008  # 0.8% protective ratio
    iocOrderRatio: 0.005              # 0.5% IOC protective ratio
    coolingDownTime: 1s               # 1 second cooldown
    
    # Monitoring
    notifyTrade: true             # Enable trade notifications
    resetPosition: false          # Don't reset positions on restart
    dryRun: false                 # Live trading mode
    
    # Trading symbols (auto-detected from paths)
    symbols:
    - BTCUSDT
    - ETHUSDT
    - ETHBTC
    - BNBUSDT
    - BNBBTC
    - BNBETH
    
    # Triangular arbitrage paths
    paths:
    - [BTCUSDT, ETHBTC, ETHUSDT]    # BTC -> ETH -> USDT -> BTC
    - [BNBBTC, BNBUSDT, BTCUSDT]    # BNB -> BTC -> USDT -> BNB
    - [BNBETH, BNBUSDT, ETHUSDT]    # BNB -> ETH -> USDT -> BNB
```

## Triangular Arbitrage Paths

### Path Structure
Each path consists of three trading pairs that form a closed loop:
```
Path: [BTCUSDT, ETHBTC, ETHUSDT]
Direction: BTC -> ETH -> USDT -> BTC
```

### Automatic Direction Detection
The strategy automatically determines trading directions:
- **Forward Path**: A → B → C → A
- **Backward Path**: A ← B ← C ← A

### Example Calculations

**Forward Arbitrage Example:**
```
Start: 1 BTC
1. BTCUSDT: Sell 1 BTC → Get 50,000 USDT
2. ETHUSDT: Buy ETH with 50,000 USDT → Get 20 ETH  
3. ETHBTC: Sell 20 ETH → Get 1.002 BTC
Profit: 0.002 BTC (0.2%)
```

**Backward Arbitrage Example:**
```
Start: 1 BTC
1. ETHBTC: Buy 20 ETH with 1 BTC
2. ETHUSDT: Sell 20 ETH → Get 50,100 USDT
3. BTCUSDT: Buy BTC with 50,100 USDT → Get 1.002 BTC
Profit: 0.002 BTC (0.2%)
```

## Execution Workflow

### 1. Market Data Monitoring
- Subscribes to order book data for all symbols in paths
- Continuously calculates best bid/ask prices
- Updates arbitrage ratios in real-time

### 2. Opportunity Detection
- Calculates forward and backward arbitrage ratios
- Compares ratios against minimum spread threshold
- Ranks opportunities by profitability

### 3. Order Execution Sequence
1. **IOC Order**: Places first order as Immediate-or-Cancel
2. **Fill Verification**: Waits for IOC order completion
3. **Quantity Adjustment**: Adjusts remaining orders based on filled quantity
4. **Market Orders**: Executes remaining orders as protective market orders
5. **Trade Collection**: Collects and analyzes all trades

### 4. Position and Profit Tracking
- Updates multi-currency position tracking
- Calculates profits in USD equivalent
- Updates performance statistics

## Risk Management Features

### 1. Balance Controls
- **Balance Limits**: Maximum exposure per currency
- **Balance Buffer**: Reserves small buffer to prevent over-trading
- **Minimum Quantity**: Validates orders meet exchange minimums

### 2. Protective Pricing
- **Market Order Protection**: Applies protective ratios to prevent slippage
- **IOC Order Protection**: Optional protective pricing for IOC orders
- **Price Validation**: Ensures orders are within reasonable price ranges

### 3. Execution Safeguards
- **IOC Strategy**: Uses IOC orders to minimize execution risk
- **Quantity Adjustment**: Adjusts subsequent orders based on actual fills
- **Order Validation**: Validates all orders before submission

### 4. Performance Monitoring
- **IOC Win Rate**: Tracks success rate of IOC orders
- **Trade Statistics**: Comprehensive trade performance analytics
- **Position Tracking**: Multi-currency position management

## Performance Optimization

### Separate Streams
```yaml
separateStream: true
```
- Creates dedicated WebSocket connections for each symbol
- Reduces latency and improves data freshness
- Better performance for high-frequency arbitrage

### Cooling Down
```yaml
coolingDownTime: 1s
```
- Prevents over-trading and exchange rate limiting
- Allows positions to settle between arbitrage cycles
- Reduces market impact

## Common Use Cases

### 1. Conservative Arbitrage
```yaml
minSpreadRatio: 1.005           # 0.5% minimum profit
coolingDownTime: 5s             # 5 second cooldown
limits:
  BTC: 0.001                    # Small position limits
  ETH: 0.01
  USDT: 100.0
```

### 2. Aggressive Arbitrage
```yaml
minSpreadRatio: 1.001           # 0.1% minimum profit
coolingDownTime: 1s             # 1 second cooldown
separateStream: true            # Optimize for speed
limits:
  BTC: 0.01                     # Larger position limits
  ETH: 0.1
  USDT: 1000.0
```

### 3. High-Volume Arbitrage
```yaml
minSpreadRatio: 1.002           # 0.2% minimum profit
separateStream: true            # Maximum performance
marketOrderProtectiveRatio: 0.005  # Tighter protective ratios
limits:
  BTC: 0.1                      # High position limits
  ETH: 1.0
  USDT: 10000.0
```

## Best Practices

1. **Path Selection**: Choose paths with high liquidity and tight spreads
2. **Spread Threshold**: Set minimum spread ratios based on typical market conditions
3. **Balance Management**: Use appropriate balance limits to control risk
4. **Stream Optimization**: Enable separate streams for better performance
5. **Cooling Down**: Use cooldown periods to prevent over-trading
6. **Monitoring**: Enable notifications to track performance

## Limitations

1. **Exchange Dependency**: Only works within a single exchange
2. **Liquidity Requirements**: Requires sufficient liquidity in all three pairs
3. **Latency Sensitivity**: Performance depends on low-latency market data
4. **Market Conditions**: Less effective during high volatility periods
5. **Competition**: Arbitrage opportunities may be quickly eliminated by other traders

## Troubleshooting

### Common Issues

**No Arbitrage Opportunities**
- Check if minimum spread ratio is too high
- Verify all symbols have sufficient liquidity
- Ensure market data streams are connected

**IOC Orders Not Filling**
- Reduce IOC protective ratio
- Check order book depth
- Verify minimum quantity requirements

**Frequent Execution Failures**
- Increase balance limits
- Check exchange API rate limits
- Verify network connectivity

**Low Profitability**
- Reduce minimum spread ratio
- Optimize protective ratios
- Consider transaction fees in calculations

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/tri.yaml)
- [Arbitrage Trading Guide](../../doc/topics/arbitrage.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
