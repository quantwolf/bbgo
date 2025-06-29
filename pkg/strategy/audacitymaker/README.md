# Audacity Maker Strategy

## Overview

The Audacity Maker strategy is an advanced order flow-based trading strategy that analyzes real-time market microstructure to identify trading opportunities. It uses statistical analysis of buyer-initiated vs. seller-initiated trading volume and trade counts to detect market imbalances and place directional trades. The strategy employs outlier detection techniques to identify significant deviations in order flow patterns, making it particularly effective in capturing short-term momentum shifts.

## How It Works

1. **Order Flow Analysis**: Continuously monitors buyer-initiated and seller-initiated trades
2. **Volume and Count Tracking**: Tracks both trading volume and number of trades for each side
3. **Statistical Normalization**: Applies min-max scaling to order flow ratios over rolling windows
4. **Outlier Detection**: Uses standard deviation-based outlier detection to identify significant imbalances
5. **Directional Trading**: Places limit orders at best bid/ask when significant order flow imbalances are detected
6. **Position Management**: Integrates with exit methods for comprehensive risk management

## Key Features

- **Real-Time Order Flow Analysis**: Processes every market trade to build order flow metrics
- **Dual Metric System**: Analyzes both volume-based and count-based order flow
- **Statistical Outlier Detection**: Uses 3-sigma threshold for signal generation
- **Min-Max Normalization**: Scales order flow ratios to 0-1 range for consistent analysis
- **Limit Maker Orders**: Places orders at best bid/ask to capture spread
- **Rolling Window Analysis**: Uses 100-period and 200-period rolling windows
- **Market Microstructure Focus**: Exploits short-term market inefficiencies
- **Exit Method Integration**: Compatible with various risk management strategies

## Strategy Logic

### Order Flow Calculation

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// Track buyer and seller initiated trades
if trade.Side == types.SideTypeBuy {
    // Buyer-initiated trade (aggressive buy)
    buyTradeSize.Update(trade.Quantity.Float64())
    sellTradeSize.Update(0)
    buyTradesNumber.Update(1)
    sellTradesNumber.Update(0)
} else if trade.Side == types.SideTypeSell {
    // Seller-initiated trade (aggressive sell)
    buyTradeSize.Update(0)
    sellTradeSize.Update(trade.Quantity.Float64())
    buyTradesNumber.Update(0)
    sellTradesNumber.Update(1)
}

// Calculate order flow ratios
sizeFraction := buyTradeSize.Sum() / sellTradeSize.Sum()
numberFraction := buyTradesNumber.Sum() / sellTradesNumber.Sum()
```
</augment_code_snippet>

### Statistical Normalization

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// Min-max scaling for order flow size
if orderFlowSize.Length() > 100 {
    ofsMax := orderFlowSize.Tail(100).Max()
    ofsMin := orderFlowSize.Tail(100).Min()
    ofsMinMax := (orderFlowSize.Last(0) - ofsMin) / (ofsMax - ofsMin)
    orderFlowSizeMinMax.Push(ofsMinMax)
}

// Min-max scaling for order flow number
if orderFlowNumber.Length() > 100 {
    ofnMax := orderFlowNumber.Tail(100).Max()
    ofnMin := orderFlowNumber.Tail(100).Min()
    ofnMinMax := (orderFlowNumber.Last(0) - ofnMin) / (ofnMax - ofnMin)
    orderFlowNumberMinMax.Push(ofnMinMax)
}
```
</augment_code_snippet>

### Outlier Detection and Trading Logic

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// Detect outliers using 3-sigma threshold
func outlier(fs floats.Slice, multiplier float64) int {
    stddev := stat.StdDev(fs, nil)
    if fs.Last(0) > fs.Mean()+multiplier*stddev {
        return 1  // Positive outlier
    } else if fs.Last(0) < fs.Mean()-multiplier*stddev {
        return -1 // Negative outlier
    }
    return 0 // No outlier
}

// Trading decisions based on dual outlier confirmation
if outlier(orderFlowSizeMinMax.Tail(100), threshold) > 0 && 
   outlier(orderFlowNumberMinMax.Tail(100), threshold) > 0 {
    // Strong buying pressure detected
    _ = s.placeOrder(ctx, types.SideTypeBuy, s.Quantity, bid.Price, symbol)
} else if outlier(orderFlowSizeMinMax.Tail(100), threshold) < 0 && 
          outlier(orderFlowNumberMinMax.Tail(100), threshold) < 0 {
    // Strong selling pressure detected
    _ = s.placeOrder(ctx, types.SideTypeSell, s.Quantity, ask.Price, symbol)
}
```
</augment_code_snippet>

### Order Placement

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
func (s *PerTrade) placeOrder(
    ctx context.Context, side types.SideType, quantity fixedpoint.Value, price fixedpoint.Value, symbol string,
) error {
    market, _ := s.session.Market(symbol)
    _, err := s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
        Symbol:   symbol,
        Market:   market,
        Side:     side,
        Type:     types.OrderTypeLimitMaker,  // Maker orders for spread capture
        Quantity: quantity,
        Price:    price,
        Tag:      "audacity-limit",
    })
    return err
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "ETHBUSD") |
| `orderFlow.interval` | interval | Yes | K-line interval for monitoring (e.g., "1m") |
| `orderFlow.quantity` | number | Yes | Fixed quantity for each trade |

### Advanced Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `orderFlow.marketOrder` | boolean | No | Enable market order execution (default: false) |

### Exit Methods
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `exits` | array | No | Array of exit method configurations |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  audacitymaker:
    symbol: ETHBUSD
    
    # Order flow configuration
    orderFlow:
      interval: 1m            # Monitor 1-minute intervals
      quantity: 0.01          # 0.01 ETH per trade
      marketOrder: false      # Use limit maker orders
    
    # Risk management (optional)
    exits:
    - roiStopLoss:
        percentage: 0.02      # 2% stop loss
    - roiTakeProfit:
        percentage: 0.01      # 1% take profit
    - trailingStop:
        callbackRate: 0.005   # 0.5% trailing stop
```

## Order Flow Analysis Explained

### What is Order Flow?
Order flow represents the difference between buyer-initiated and seller-initiated trading activity. It provides insights into:
- **Market Sentiment**: Whether buyers or sellers are more aggressive
- **Liquidity Dynamics**: How market participants are interacting with the order book
- **Short-term Momentum**: Immediate directional pressure in the market

### Dual Metric Approach
The strategy analyzes two complementary metrics:

1. **Volume-Based Order Flow**: `buyVolume / sellVolume`
   - Measures the size of aggressive trades
   - Indicates institutional or large trader activity
   - More sensitive to significant market moves

2. **Count-Based Order Flow**: `buyTrades / sellTrades`
   - Measures the frequency of aggressive trades
   - Indicates retail or algorithmic activity
   - More sensitive to sustained directional pressure

### Statistical Processing Pipeline

#### 1. Raw Data Collection
- Tracks buyer and seller initiated trades in real-time
- Uses 200-period rolling queues for volume and trade counts
- Maintains separate queues for buy and sell activity

#### 2. Ratio Calculation
```
Size Ratio = Sum(Buy Volume) / Sum(Sell Volume)
Count Ratio = Sum(Buy Trades) / Sum(Sell Trades)
```

#### 3. Min-Max Normalization
```
Normalized Value = (Current - Min) / (Max - Min)
```
- Scales ratios to 0-1 range over 100-period windows
- Enables consistent comparison across different market conditions
- Preserves temporal relationships in the data

#### 4. Outlier Detection
```
Outlier Threshold = Mean ± (3 × Standard Deviation)
```
- Uses 3-sigma rule for outlier identification
- Positive outliers indicate strong buying pressure
- Negative outliers indicate strong selling pressure

## Trading Logic

### Signal Generation
The strategy requires **dual confirmation** for trade signals:
- Both volume-based AND count-based metrics must show outliers
- Same direction outliers confirm the signal strength
- Opposite direction outliers are ignored (no trade)

### Entry Conditions

**Long Entry:**
1. Volume-based order flow shows positive outlier (> mean + 3σ)
2. Count-based order flow shows positive outlier (> mean + 3σ)
3. Places limit buy order at current best bid price

**Short Entry:**
1. Volume-based order flow shows negative outlier (< mean - 3σ)
2. Count-based order flow shows negative outlier (< mean - 3σ)
3. Places limit sell order at current best ask price

### Order Management
- **Order Type**: Limit Maker orders to capture spread
- **Price**: Best bid for buys, best ask for sells
- **Cancellation**: Cancels existing orders before placing new ones
- **Quantity**: Fixed quantity per configuration

## Performance Characteristics

### Market Conditions
- **High-Frequency Markets**: Excels in markets with frequent trades
- **Liquid Markets**: Requires sufficient trade flow for analysis
- **Volatile Markets**: Benefits from clear directional moves
- **Trending Markets**: Can capture momentum shifts effectively

### Time Sensitivity
- **Real-Time Processing**: Processes every market trade immediately
- **Short-Term Focus**: Designed for capturing quick momentum shifts
- **Scalping Nature**: Typically holds positions for short periods

### Statistical Requirements
- **Minimum Data**: Requires 100+ trades for reliable signals
- **Rolling Windows**: Uses 100-200 period windows for stability
- **Outlier Threshold**: 3-sigma threshold balances sensitivity and noise

## Risk Management

### Built-in Protections
- **Dual Confirmation**: Requires both metrics to agree
- **Limit Orders**: Uses maker orders to avoid immediate market impact
- **Order Cancellation**: Cancels conflicting orders before new trades

### Recommended Risk Controls
- **Position Sizing**: Use small fixed quantities
- **Stop Losses**: Implement tight stop losses (1-2%)
- **Take Profits**: Quick profit taking (0.5-1%)
- **Time Limits**: Consider time-based exits for stale positions

## Common Use Cases

### 1. High-Frequency Scalping
```yaml
orderFlow:
  interval: 1m
  quantity: 0.001
exits:
- roiTakeProfit:
    percentage: 0.005  # 0.5% quick profits
```

### 2. Momentum Capture
```yaml
orderFlow:
  interval: 1m
  quantity: 0.01
exits:
- roiTakeProfit:
    percentage: 0.01   # 1% momentum profits
- roiStopLoss:
    percentage: 0.02   # 2% risk control
```

### 3. Conservative Order Flow
```yaml
orderFlow:
  interval: 5m
  quantity: 0.005
exits:
- trailingStop:
    callbackRate: 0.01 # 1% trailing stop
```

## Best Practices

1. **Market Selection**: Choose highly liquid markets with frequent trades
2. **Quantity Sizing**: Start with small quantities to test effectiveness
3. **Risk Management**: Always use exit methods for protection
4. **Monitoring**: Watch for changes in market microstructure
5. **Backtesting**: Test thoroughly on historical data before live trading
6. **Parameter Tuning**: Consider adjusting outlier threshold for different markets

## Limitations

1. **Data Dependency**: Requires high-frequency trade data
2. **Market Structure**: Performance varies with market microstructure changes
3. **Latency Sensitivity**: Effectiveness depends on low-latency execution
4. **False Signals**: May generate false signals in choppy markets
5. **Transaction Costs**: High-frequency nature may incur significant fees
6. **Market Impact**: Large quantities may affect the very patterns being exploited

## Troubleshooting

### Common Issues

**No Trading Signals**
- Check if market has sufficient trade frequency
- Verify order flow data is being collected
- Ensure 100+ trades have occurred for analysis

**Excessive False Signals**
- Consider increasing outlier threshold (> 3 sigma)
- Add additional confirmation filters
- Implement minimum holding periods

**Poor Fill Rates**
- Check if limit orders are competitive
- Consider using market orders for critical signals
- Verify order book depth and spreads

**High Transaction Costs**
- Reduce trading frequency with higher thresholds
- Optimize quantity sizing
- Consider maker rebate programs

## Advanced Concepts

### Market Microstructure Theory
The strategy is based on the principle that:
- Aggressive trades reveal private information
- Order flow imbalances predict short-term price movements
- Statistical outliers indicate significant market events

### Statistical Foundation
- **Central Limit Theorem**: Assumes order flow ratios follow normal distribution
- **Outlier Detection**: Uses standard deviation to identify rare events
- **Min-Max Scaling**: Normalizes data for consistent analysis across time

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/audacitymaker.yaml)
- [Order Flow Analysis Guide](../../doc/topics/order-flow.md)
- [Market Microstructure Trading](../../doc/topics/microstructure.md)
