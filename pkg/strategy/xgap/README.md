# XGap Strategy (Cross-Exchange Gap Trading)

## Overview

The XGap strategy is a sophisticated cross-exchange arbitrage strategy that exploits price gaps between different exchanges by providing liquidity on the trading exchange while monitoring price movements on a source exchange. It acts as a market maker that places both buy and sell orders at mid-price levels, profiting from the bid-ask spread while managing position risk through intelligent order placement and cancellation.

## How It Works

1. **Cross-Exchange Monitoring**: Monitors order books on both source and trading exchanges
2. **Gap Detection**: Identifies price gaps and spread opportunities between exchanges
3. **Liquidity Provision**: Places buy and sell orders at calculated mid-prices
4. **Position Management**: Automatically adjusts positions when profitable opportunities arise
5. **Risk Control**: Implements spread thresholds and position limits
6. **Volume Simulation**: Can simulate trading volume based on source exchange activity

## Key Features

- **Cross-Exchange Arbitrage**: Exploits price differences between exchanges
- **Intelligent Market Making**: Places orders at optimal mid-price levels
- **Position Adjustment**: Automatically closes profitable positions
- **Spread Management**: Configurable minimum spread requirements
- **Volume Simulation**: Mimics trading patterns from source exchange
- **Risk Controls**: Multiple safeguards including balance checks and dust quantity filters
- **Flexible Pricing**: Options for mid-price or below-ask pricing strategies
- **Make Spread Feature**: Can actively create spreads when markets are too tight

## Strategy Logic

### Cross-Exchange Price Monitoring

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) CrossSubscribe(sessions map[string]*bbgo.ExchangeSession) {
    if len(s.SourceExchange) > 0 && len(s.SourceSymbol) > 0 {
        sourceSession, ok := sessions[s.SourceExchange]
        if !ok {
            panic(fmt.Errorf("source session %s is not defined", s.SourceExchange))
        }
        
        sourceSession.Subscribe(types.KLineChannel, s.SourceSymbol, types.SubscribeOptions{Interval: "1m"})
        sourceSession.Subscribe(types.BookChannel, s.SourceSymbol, types.SubscribeOptions{Depth: types.DepthLevelFull})
    }
    
    tradingSession.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: "1m"})
    tradingSession.Subscribe(types.BookChannel, s.Symbol, types.SubscribeOptions{Depth: types.DepthLevel5})
}
```
</augment_code_snippet>

### Position Adjustment Logic

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func buildAdjustPositionOrder(
    symbol string,
    positionSnapshot *types.Position,
    bestBid, bestAsk types.PriceVolume,
) (ok bool, order types.SubmitOrder) {
    if positionSnapshot.IsClosed() {
        return
    }
    
    var pv types.PriceVolume
    var side types.SideType
    if positionSnapshot.IsShort() && bestAsk.Price.Compare(positionSnapshot.AverageCost) < 0 {
        // In short position and best ask price is less than average cost
        pv = bestAsk
        side = types.SideTypeBuy
    } else if positionSnapshot.IsLong() && bestBid.Price.Compare(positionSnapshot.AverageCost) > 0 {
        // In long position and best bid price is greater than average cost
        pv = bestBid
        side = types.SideTypeSell
    }
    
    price := pv.Price
    quantity := fixedpoint.Min(positionSnapshot.Base.Abs(), pv.Volume)
    
    if !pv.IsZero() {
        order = types.SubmitOrder{
            Symbol:      symbol,
            Side:        side,
            Type:        types.OrderTypeLimit,
            Quantity:    quantity,
            Price:       price,
            TimeInForce: types.TimeInForceIOC,
        }
        ok = true
    }
    return
}
```
</augment_code_snippet>

### Order Placement Strategy

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) placeOrders(ctx context.Context) {
    bestBid, bestAsk, hasPrice := s.tradingBook.BestBidAndAsk()
    
    if hasPrice {
        var spread = bestAsk.Price.Sub(bestBid.Price)
        var spreadPercentage = spread.Div(bestAsk.Price)
        
        // Use source book price if spread is too large (>5%)
        if s.SimulatePrice && s.sourceBook != nil && spreadPercentage.Compare(maxStepPercentageGap) > 0 {
            bestBid, bestAsk, hasPrice = s.sourceBook.BestBidAndAsk()
        }
        
        // Check minimum spread requirement
        if s.MinSpread.Sign() > 0 && spreadPercentage.Compare(s.MinSpread) < 0 {
            if s.MakeSpread.Enabled {
                s.makeSpread(ctx, bestBid, bestAsk)
            }
            return
        }
    }
    
    var midPrice = bestAsk.Price.Add(bestBid.Price).Div(Two)
    var price fixedpoint.Value
    
    if s.SellBelowBestAsk {
        price = bestAsk.Price.Sub(s.tradingMarket.TickSize)
    } else {
        price = adjustPrice(midPrice, s.tradingMarket.PricePrecision)
    }
    
    // Place both buy and sell orders at the calculated price
    orderForms := []types.SubmitOrder{
        {
            Symbol:   s.Symbol,
            Side:     types.SideTypeBuy,
            Type:     types.OrderTypeLimit,
            Quantity: quantity,
            Price:    price,
        },
        {
            Symbol:      s.Symbol,
            Side:        types.SideTypeSell,
            Type:        types.OrderTypeLimit,
            Quantity:    quantity,
            Price:       price,
            TimeInForce: types.TimeInForceIOC,
        },
    }
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol on trading exchange |
| `tradingExchange` | string | Yes | Exchange where orders will be placed |
| `sourceExchange` | string | Yes | Exchange to monitor for price reference |
| `sourceSymbol` | string | No | Symbol on source exchange (defaults to symbol) |

### Spread and Pricing
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `minSpread` | percentage | No | Minimum spread required to place orders |
| `sellBelowBestAsk` | boolean | No | Place sell orders 1 tick below best ask |
| `simulatePrice` | boolean | No | Use source exchange price when spread is large |

### Order Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quantity` | number | No | Fixed order quantity (if not set, uses calculated quantity) |
| `maxJitterQuantity` | number | No | Maximum quantity jitter for randomization |
| `updateInterval` | duration | No | Interval between order updates (default: 1s) |

### Make Spread Feature
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `makeSpread.enabled` | boolean | No | Enable active spread making |
| `makeSpread.skipLargeQuantityThreshold` | number | No | Skip making spread if quantity exceeds threshold |

### Volume Control
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `dailyMaxVolume` | number | No | Maximum daily trading volume |
| `dailyTargetVolume` | number | No | Target daily trading volume |
| `simulateVolume` | boolean | No | Simulate volume based on source exchange |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `dryRun` | boolean | No | Enable dry run mode (no actual orders) |

## Configuration Example

```yaml
crossExchangeStrategies:
- xgap:
    # Basic settings
    symbol: BTCUSDT
    tradingExchange: max          # Exchange to place orders
    sourceExchange: binance       # Exchange to monitor prices
    sourceSymbol: BTCUSDT         # Symbol on source exchange
    
    # Spread and pricing
    minSpread: 0.001              # 0.1% minimum spread
    sellBelowBestAsk: false       # Use mid-price strategy
    simulatePrice: true           # Use source price when spread is large
    
    # Order management
    quantity: 0.01                # Fixed 0.01 BTC per order
    maxJitterQuantity: 0.005      # Add up to 0.005 BTC jitter
    updateInterval: 30s           # Update orders every 30 seconds
    
    # Make spread feature
    makeSpread:
      enabled: true               # Enable spread making
      skipLargeQuantityThreshold: 1.0  # Skip if quantity > 1.0 BTC
    
    # Volume control
    dailyMaxVolume: 10.0          # Maximum 10 BTC daily volume
    simulateVolume: true          # Match source exchange volume
    
    # Risk management
    dryRun: false                 # Live trading mode
    
    # Fee budget (inherited from common strategy)
    dailyFeeBudgets:
      MAX: 100                    # $100 daily fee budget
```

## Strategy Workflow

### 1. Market Data Collection
- Subscribes to order book data from both source and trading exchanges
- Monitors 1-minute klines for volume analysis
- Maintains real-time best bid/ask prices

### 2. Spread Analysis
- Calculates spread between best bid and ask on trading exchange
- Compares with minimum spread requirements
- Falls back to source exchange prices if trading spread is too large (>5%)

### 3. Position Assessment
- Checks current position status
- Attempts to place position adjustment orders if profitable
- Prioritizes closing profitable positions over opening new ones

### 4. Order Placement
- Calculates optimal order price (mid-price or below best ask)
- Determines order quantity based on configuration or volume simulation
- Places both buy and sell orders simultaneously
- Uses IOC (Immediate or Cancel) for sell orders to prevent self-trading

### 5. Order Management
- Automatically cancels orders after 1 second
- Continuously updates orders based on market conditions
- Implements jitter timing to avoid predictable patterns

## Pricing Strategies

### Mid-Price Strategy (Default)
```
Price = (Best Bid + Best Ask) / 2
```
- Places orders at the middle of the spread
- Maximizes probability of execution from both sides

### Below Best Ask Strategy
```
Price = Best Ask - 1 Tick
```
- Places orders just below the best ask price
- More aggressive pricing for faster execution

## Volume Simulation

### Fixed Quantity Mode
- Uses configured `quantity` for all orders
- Adds random jitter if `maxJitterQuantity` is set

### Volume Simulation Mode
- Monitors volume difference between source and trading exchanges
- Adjusts order quantity based on volume gaps
- Helps maintain similar trading activity across exchanges

### Daily Target Mode
- Calculates quantity to achieve `dailyTargetVolume`
- Distributes volume evenly across trading intervals

## Risk Management Features

### 1. Spread Controls
- **Minimum Spread**: Prevents trading when spreads are too tight
- **Maximum Spread**: Falls back to source pricing when spreads are too wide
- **Tick Size Validation**: Ensures spreads are at least 2 ticks wide

### 2. Position Management
- **Automatic Adjustment**: Closes profitable positions immediately
- **Balance Validation**: Ensures sufficient balance before placing orders
- **Dust Quantity Filter**: Prevents orders below minimum thresholds

### 3. Order Safeguards
- **Self-Trading Prevention**: Uses IOC orders to prevent matching own orders
- **Price Validation**: Ensures orders are within bid-ask range
- **Balance Checks**: Validates sufficient funds for each order

### 4. Make Spread Feature
- **Liquidity Injection**: Actively creates spreads when markets are too tight
- **Quantity Limits**: Skips large quantities to avoid market impact
- **Balance Protection**: Ensures sufficient balance before making spreads

## Performance Optimization

### Timing and Jitter
- Uses jittered update intervals to avoid predictable patterns
- Implements random delays to reduce market impact
- Balances responsiveness with stability

### Order Lifecycle
- Short-lived orders (1 second) to minimize exposure
- Immediate cancellation and replacement for fresh pricing
- IOC orders for sell side to prevent accumulation

## Common Use Cases

### 1. Conservative Arbitrage
```yaml
minSpread: 0.002              # 0.2% minimum spread
quantity: 0.001               # Small fixed quantity
updateInterval: 60s           # Slower updates
makeSpread:
  enabled: false              # No active spread making
```

### 2. Aggressive Market Making
```yaml
minSpread: 0.0005             # 0.05% minimum spread
simulateVolume: true          # Match source volume
updateInterval: 10s           # Faster updates
makeSpread:
  enabled: true               # Active spread making
```

### 3. Volume Matching
```yaml
simulateVolume: true          # Match source exchange volume
dailyTargetVolume: 50.0       # Target 50 BTC daily
maxJitterQuantity: 0.01       # Add quantity randomization
```

## Best Practices

1. **Exchange Selection**: Choose exchanges with good API reliability and low latency
2. **Spread Tuning**: Set minimum spreads based on typical market conditions
3. **Volume Management**: Use daily limits to control exposure
4. **Fee Monitoring**: Set appropriate fee budgets to maintain profitability
5. **Risk Limits**: Start with small quantities and gradually increase
6. **Market Hours**: Consider different trading hours between exchanges

## Limitations

1. **Latency Sensitivity**: Performance depends on network latency between exchanges
2. **API Rate Limits**: Subject to exchange API rate limiting
3. **Market Conditions**: Less effective in highly volatile or illiquid markets
4. **Exchange Risk**: Exposure to exchange-specific risks and downtime
5. **Regulatory Compliance**: Must comply with regulations on both exchanges

## Troubleshooting

### Common Issues

**No Orders Placed**
- Check minimum spread requirements
- Verify sufficient balance on trading exchange
- Ensure spread is at least 2 ticks wide

**Frequent Order Cancellations**
- Adjust update interval for market conditions
- Check if spreads are meeting minimum requirements
- Verify network connectivity to both exchanges

**Position Accumulation**
- Enable position adjustment feature
- Check if adjustment orders are being filled
- Review spread thresholds and market conditions

**Low Profitability**
- Increase minimum spread requirements
- Optimize order quantities
- Review fee structures on both exchanges

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/xgap.yaml)
- [Cross-Exchange Trading Guide](../../doc/topics/cross-exchange.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
