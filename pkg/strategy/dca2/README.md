# DCA2 Strategy (Dollar Cost Averaging v2)

## Overview

The DCA2 strategy is an advanced Dollar Cost Averaging implementation that automatically places multiple buy orders at decreasing price levels and takes profit when the average cost reaches a target profit ratio. It's designed for accumulating assets during market downturns while maintaining strict risk management through systematic position building and profit-taking.

## How It Works

1. **Multi-Level Buy Orders**: Places multiple buy orders at decreasing price levels based on price deviation
2. **Equal Notional Distribution**: Distributes quote investment equally across all buy orders
3. **Automatic Take Profit**: Calculates take profit price based on average cost and profit ratio
4. **State Machine Management**: Uses a sophisticated state machine to manage order lifecycle
5. **Round-Based Trading**: Operates in rounds with cooldown periods between rounds
6. **Profit Reinvestment**: Automatically reinvests profits into the next round

## Key Features

- **Systematic DCA Approach**: Places orders at predetermined price levels below current market price
- **State Machine Architecture**: Robust state management for order lifecycle
- **Automatic Recovery**: Can recover from interruptions and continue trading
- **Profit Tracking**: Comprehensive profit statistics with round-based tracking
- **Risk Management**: Built-in safeguards and validation mechanisms
- **Flexible Configuration**: Highly configurable parameters for different market conditions
- **Order Group Management**: Uses order groups for efficient order management
- **Cooldown Mechanism**: Prevents overtrading with configurable cooldown periods

## Strategy Logic

### State Machine Flow

<augment_code_snippet path="pkg/strategy/dca2/state.go" mode="EXCERPT">
```go
type State int64

const (
    None State = iota
    IdleWaiting
    OpenPositionReady
    OpenPositionOrderFilled
    OpenPositionFinished
    TakeProfitReady
)

var stateTransition map[State]State = map[State]State{
    IdleWaiting:             OpenPositionReady,
    OpenPositionReady:       OpenPositionOrderFilled,
    OpenPositionOrderFilled: OpenPositionFinished,
    OpenPositionFinished:    TakeProfitReady,
    TakeProfitReady:         IdleWaiting,
}
```
</augment_code_snippet>

### Open Position Order Generation

<augment_code_snippet path="pkg/strategy/dca2/open_position.go" mode="EXCERPT">
```go
func generateOpenPositionOrders(market types.Market, enableQuoteInvestmentReallocate bool, quoteInvestment, profit, price, priceDeviation fixedpoint.Value, maxOrderCount int64, orderGroupID uint32) ([]types.SubmitOrder, error) {
    factor := fixedpoint.One.Sub(priceDeviation)
    
    // Calculate all valid prices
    var prices []fixedpoint.Value
    for i := 0; i < int(maxOrderCount); i++ {
        if i > 0 {
            price = price.Mul(factor)
        }
        price = market.TruncatePrice(price)
        if price.Compare(market.MinPrice) < 0 {
            break
        }
        prices = append(prices, price)
    }
    
    notional, orderNum := calculateNotionalAndNumOrders(market, quoteInvestment, prices)
    
    // Generate submit orders with equal notional
    var submitOrders []types.SubmitOrder
    for i := 0; i < orderNum; i++ {
        var quantity fixedpoint.Value
        if i == 0 {
            // First order includes accumulated profit
            quantity = market.TruncateQuantity(notional.Add(profit).Div(prices[i]))
        } else {
            quantity = market.TruncateQuantity(notional.Div(prices[i]))
        }
        submitOrders = append(submitOrders, types.SubmitOrder{
            Symbol:      market.Symbol,
            Type:        types.OrderTypeLimit,
            Price:       prices[i],
            Side:        types.SideTypeBuy,
            Quantity:    quantity,
            GroupID:     orderGroupID,
        })
    }
    
    return submitOrders, nil
}
```
</augment_code_snippet>

### Take Profit Calculation

<augment_code_snippet path="pkg/strategy/dca2/take_profit.go" mode="EXCERPT">
```go
func generateTakeProfitOrder(market types.Market, takeProfitRatio fixedpoint.Value, position *types.Position, orderGroupID uint32) types.SubmitOrder {
    takeProfitPrice := market.TruncatePrice(position.AverageCost.Mul(fixedpoint.One.Add(takeProfitRatio)))
    return types.SubmitOrder{
        Symbol:      market.Symbol,
        Type:        types.OrderTypeLimit,
        Price:       takeProfitPrice,
        Side:        types.SideTypeSell,
        Quantity:    position.GetBase().Abs(),
        GroupID:     orderGroupID,
    }
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `quoteInvestment` | number | Yes | Total quote currency to invest per round |
| `maxOrderCount` | int | Yes | Maximum number of buy orders to place |
| `priceDeviation` | percentage | Yes | Price deviation between consecutive orders |
| `takeProfitRatio` | percentage | Yes | Profit ratio for take profit calculation |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `coolDownInterval` | duration | Yes | Cooldown period between rounds (seconds) |
| `orderGroupID` | int | No | Custom order group ID for order management |
| `disableOrderGroupIDFilter` | boolean | No | Disable order group ID filtering |

### Recovery and Persistence
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `recoverWhenStart` | boolean | No | Enable recovery on strategy start |
| `disableProfitStatsRecover` | boolean | No | Disable profit stats recovery |
| `disablePositionRecover` | boolean | No | Disable position recovery |
| `persistenceTTL` | duration | No | Time-to-live for persistence data |

### Advanced Options
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `enableQuoteInvestmentReallocate` | boolean | No | Allow reallocation when orders are under minimum |
| `keepOrdersWhenShutdown` | boolean | No | Keep orders active when shutting down |
| `useCancelAllOrdersApiWhenClose` | boolean | No | Use cancel all orders API when closing |

### Monitoring and Debugging
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `logFields` | object | No | Additional log fields for debugging |
| `prometheusLabels` | object | No | Custom Prometheus labels for metrics |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  dca2:
    symbol: BTCUSDT
    
    # Investment parameters
    quoteInvestment: 1000        # Invest $1000 per round
    maxOrderCount: 5             # Place 5 buy orders
    priceDeviation: 0.02         # 2% price deviation between orders
    takeProfitRatio: 0.005       # 0.5% take profit ratio
    
    # Risk management
    coolDownInterval: 3600       # 1 hour cooldown between rounds
    
    # Recovery settings
    recoverWhenStart: true       # Enable recovery on start
    disableProfitStatsRecover: false
    disablePositionRecover: false
    
    # Order management
    keepOrdersWhenShutdown: false
    useCancelAllOrdersApiWhenClose: false
    enableQuoteInvestmentReallocate: true
    
    # Persistence
    persistenceTTL: 24h          # Keep data for 24 hours
    
    # Monitoring
    logFields:
      environment: "production"
    prometheusLabels:
      instance: "dca-btc-001"
```

## Strategy Workflow

### 1. Round Initialization (IdleWaiting → OpenPositionReady)
- Waits for cooldown period to complete
- Validates no existing open position orders
- Calculates order prices based on current market price and price deviation
- Places multiple buy orders at decreasing price levels

### 2. Position Building (OpenPositionReady → OpenPositionOrderFilled)
- Monitors buy order fills
- Updates position as orders are filled
- Calculates running average cost

### 3. Position Completion (OpenPositionOrderFilled → OpenPositionFinished)
- Triggered when price reaches take profit level or all orders are filled
- Transitions to take profit stage

### 4. Take Profit (OpenPositionFinished → TakeProfitReady)
- Cancels remaining buy orders
- Places take profit sell order at calculated price
- Take profit price = Average Cost × (1 + Take Profit Ratio)

### 5. Round Completion (TakeProfitReady → IdleWaiting)
- Monitors take profit order execution
- Updates profit statistics
- Resets position for next round
- Starts cooldown period

## Order Placement Logic

### Price Calculation
```
Order 1: Current Price
Order 2: Current Price × (1 - Price Deviation)
Order 3: Current Price × (1 - Price Deviation)²
Order 4: Current Price × (1 - Price Deviation)³
Order 5: Current Price × (1 - Price Deviation)⁴
```

### Quantity Calculation
```
Notional per Order = Quote Investment ÷ Number of Orders
Quantity = Notional ÷ Order Price

Special case for first order:
First Order Quantity = (Notional + Accumulated Profit) ÷ First Order Price
```

## Risk Management Features

### 1. Equal Notional Distribution
- Each order has the same notional value
- Ensures balanced position building
- Prevents over-concentration at any price level

### 2. Minimum Requirements Validation
- Validates orders meet exchange minimum notional requirements
- Automatically adjusts order count if needed (when reallocation enabled)
- Prevents failed order submissions

### 3. Balance Verification
- Verifies sufficient balance before placing take profit orders
- Prevents overselling scenarios
- Warns about balance discrepancies

### 4. State Machine Safeguards
- Validates state transitions
- Prevents invalid operations
- Ensures proper order lifecycle management

## Performance Metrics

### Profit Statistics
- **Total Profit**: Cumulative profit across all rounds
- **Current Round Profit**: Profit from current round
- **Quote Investment**: Total investment including reinvested profits
- **Round Count**: Number of completed rounds

### Order Metrics
- **Active Orders**: Number of currently active orders
- **Fill Rate**: Percentage of orders filled
- **Average Fill Price**: Volume-weighted average fill price

## Common Use Cases

### 1. Conservative DCA Setup
```yaml
quoteInvestment: 500
maxOrderCount: 3
priceDeviation: 0.05         # 5% between orders
takeProfitRatio: 0.01        # 1% take profit
coolDownInterval: 86400      # 24 hours cooldown
```

### 2. Aggressive DCA Setup
```yaml
quoteInvestment: 1000
maxOrderCount: 7
priceDeviation: 0.015        # 1.5% between orders
takeProfitRatio: 0.003       # 0.3% take profit
coolDownInterval: 3600       # 1 hour cooldown
```

### 3. Conservative Long-term Setup
```yaml
quoteInvestment: 2000
maxOrderCount: 4
priceDeviation: 0.08         # 8% between orders
takeProfitRatio: 0.02        # 2% take profit
coolDownInterval: 604800     # 1 week cooldown
```

## Best Practices

1. **Price Deviation Tuning**: Set price deviation based on asset volatility
2. **Take Profit Optimization**: Balance between frequent profits and profit size
3. **Cooldown Management**: Use longer cooldowns in trending markets
4. **Investment Sizing**: Size investment based on available capital and risk tolerance
5. **Recovery Settings**: Enable recovery for production environments
6. **Monitoring**: Use Prometheus labels for comprehensive monitoring

## Limitations

1. **Downtrend Dependency**: Works best in ranging or slightly declining markets
2. **Capital Requirements**: Requires sufficient capital for multiple orders
3. **Exchange Limits**: Subject to exchange minimum notional and quantity limits
4. **Market Impact**: Large orders may impact market price
5. **Timing Risk**: May miss opportunities during rapid price movements

## Troubleshooting

### Common Issues

**Orders Not Placed**
- Check quote investment meets minimum requirements
- Verify price deviation doesn't create prices below minimum
- Ensure sufficient account balance

**Take Profit Not Triggered**
- Verify take profit ratio is reasonable for market conditions
- Check if price has reached calculated take profit level
- Ensure take profit order was placed successfully

**Recovery Failures**
- Check persistence configuration
- Verify order group ID consistency
- Review log files for specific error messages

**State Machine Stuck**
- Monitor state transitions in logs
- Check for network connectivity issues
- Verify exchange API responses

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/dca2.yaml)
- [Original DCA Strategy](../dca/README.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
