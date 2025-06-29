# Drift Strategy

## Overview

The Drift strategy is a sophisticated momentum-based trading strategy that uses a custom Drift Moving Average (DriftMA) indicator to identify market momentum shifts and trend changes. It combines weighted drift calculations with multiple smoothing techniques and advanced risk management features including trailing stops, ATR-based stop losses, and dynamic rebalancing.

## How It Works

1. **Drift Calculation**: Uses a weighted drift indicator to measure momentum changes in price movement
2. **Signal Smoothing**: Applies multiple layers of smoothing using EWMA and Fisher Transform
3. **Trend Detection**: Identifies momentum shifts through drift crossovers and directional changes
4. **Dynamic Positioning**: Adjusts position sizes based on market volatility and trend strength
5. **Risk Management**: Implements multiple stop loss mechanisms and trailing stops
6. **Smart Order Management**: Uses intelligent order cancellation and replacement logic

## Key Features

- **Custom DriftMA Indicator**: Proprietary momentum indicator combining weighted drift with smoothing
- **Multi-Timeframe Analysis**: Operates on both main interval and minimum interval for precise timing
- **Advanced Stop Loss**: Multiple stop loss types including percentage, ATR-based, and trailing stops
- **Dynamic Rebalancing**: Automatic position rebalancing based on trend line regression
- **Smart Order Management**: Intelligent order cancellation based on market conditions
- **Comprehensive Analytics**: Detailed performance tracking and graph generation
- **High-Frequency Trading**: Supports very short intervals (1s) for scalping strategies
- **Volatility-Based Pricing**: Uses standard deviation to adjust order prices

## Strategy Logic

### DriftMA Indicator

<augment_code_snippet path="pkg/strategy/drift/driftma.go" mode="EXCERPT">
```go
type DriftMA struct {
    types.SeriesBase
    drift *indicator.WeightedDrift
    ma1   types.UpdatableSeriesExtend
    ma2   types.UpdatableSeriesExtend
}

func (s *DriftMA) Update(value, weight float64) {
    s.ma1.Update(value)
    if s.ma1.Length() == 0 {
        return
    }
    s.drift.Update(s.ma1.Last(0), weight)
    if s.drift.Length() == 0 {
        return
    }
    s.ma2.Update(s.drift.Last(0))
}
```
</augment_code_snippet>

### Signal Generation

<augment_code_snippet path="pkg/strategy/drift/strategy.go" mode="EXCERPT">
```go
drift := s.drift.Array(2)
ddrift := s.drift.drift.Array(2)

shortCondition := drift[1] >= 0 && drift[0] <= 0 || 
                 (drift[1] >= drift[0] && drift[1] <= 0) || 
                 ddrift[1] >= 0 && ddrift[0] <= 0 || 
                 (ddrift[1] >= ddrift[0] && ddrift[1] <= 0)

longCondition := drift[1] <= 0 && drift[0] >= 0 || 
                (drift[1] <= drift[0] && drift[1] >= 0) || 
                ddrift[1] <= 0 && ddrift[0] >= 0 || 
                (ddrift[1] <= ddrift[0] && ddrift[1] >= 0)

// Resolve conflicts using price direction
if shortCondition && longCondition {
    if s.priceLines.Index(1) > s.priceLines.Last(0) {
        longCondition = false
    } else {
        shortCondition = false
    }
}
```
</augment_code_snippet>

### Stop Loss Implementation

<augment_code_snippet path="pkg/strategy/drift/stoploss.go" mode="EXCERPT">
```go
func (s *Strategy) CheckStopLoss() bool {
    if s.UseStopLoss {
        stoploss := s.StopLoss.Float64()
        if s.sellPrice > 0 && s.sellPrice*(1.+stoploss) <= s.highestPrice ||
           s.buyPrice > 0 && s.buyPrice*(1.-stoploss) >= s.lowestPrice {
            return true
        }
    }
    if s.UseAtr {
        atr := s.atr.Last(0)
        if s.sellPrice > 0 && s.sellPrice+atr <= s.highestPrice ||
           s.buyPrice > 0 && s.buyPrice-atr >= s.lowestPrice {
            return true
        }
    }
    return false
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `interval` | interval | Yes | Main analysis interval (e.g., "1s", "1m") |
| `minInterval` | interval | Yes | Minimum interval for stop loss checks |
| `window` | int | Yes | Window size for main moving average |

### Drift Indicator Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `predictOffset` | int | Yes | Lookback length for linear regression prediction |
| `smootherWindow` | int | Yes | Window for EWMA smoothing of drift |
| `fisherTransformWindow` | int | Yes | Window for Fisher Transform filtering |
| `hlRangeWindow` | int | Yes | Window for high/low variance calculation |
| `hlVarianceMultiplier` | float | Yes | Multiplier for price adjustment based on volatility |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `useStopLoss` | boolean | No | Enable percentage-based stop loss |
| `stoploss` | percentage | No | Stop loss percentage from entry price |
| `useAtr` | boolean | No | Enable ATR-based stop loss |
| `atrWindow` | int | Yes | Window for ATR calculation |
| `noTrailingStopLoss` | boolean | No | Disable trailing stop loss |

### Trailing Stop Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `trailingActivationRatio` | array | No | Activation ratios for trailing stops |
| `trailingCallbackRate` | array | No | Callback rates for trailing stops |

### Order Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `pendingMinInterval` | int | Yes | Time before canceling unfilled orders |
| `limitOrder` | boolean | No | Use limit orders instead of market orders |

### Rebalancing
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `noRebalance` | boolean | No | Disable automatic rebalancing |
| `trendWindow` | int | Yes | Window for trend line calculation |
| `rebalanceFilter` | float | Yes | Beta filter for rebalancing decisions |

### Analytics and Debugging
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `debug` | boolean | No | Enable debug logging |
| `generateGraph` | boolean | No | Generate performance graphs |
| `graphPNLDeductFee` | boolean | No | Deduct fees from PnL graphs |
| `canvasPath` | string | No | Path for indicator graphs |
| `graphPNLPath` | string | No | Path for PnL graphs |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  drift:
    symbol: BTCUSDT
    
    # Time intervals
    interval: 1s              # Main analysis interval
    minInterval: 1s           # Minimum interval for stop checks
    window: 2                 # Main MA window
    
    # Drift indicator settings
    predictOffset: 2          # Linear regression lookback
    smootherWindow: 10        # EWMA smoothing window
    fisherTransformWindow: 45 # Fisher transform window
    hlRangeWindow: 6          # High/low variance window
    hlVarianceMultiplier: 0.7 # Price adjustment multiplier
    
    # Risk management
    useStopLoss: true         # Enable percentage stop loss
    stoploss: 0.01%           # 0.01% stop loss
    useAtr: true              # Enable ATR stop loss
    atrWindow: 24             # ATR calculation window
    noTrailingStopLoss: false # Enable trailing stops
    
    # Trailing stops (multiple levels)
    trailingActivationRatio: [0.0008, 0.002, 0.01]
    trailingCallbackRate: [0.00014, 0.0003, 0.0016]
    
    # Order management
    pendingMinInterval: 6     # Cancel orders after 6 intervals
    limitOrder: true          # Use limit orders
    
    # Rebalancing
    noRebalance: false        # Enable rebalancing
    trendWindow: 4            # Trend line window
    rebalanceFilter: 2        # Beta filter threshold
    
    # Analytics
    debug: false              # Disable debug logging
    generateGraph: true       # Generate performance graphs
    graphPNLDeductFee: false  # Don't deduct fees from graphs
    canvasPath: "./output.png"
    graphPNLPath: "./pnl.png"
```

## Strategy Components

### 1. DriftMA Indicator
- **First Smoothing**: EWMA applied to price source
- **Drift Calculation**: Weighted drift of smoothed prices
- **Second Smoothing**: Fisher Transform applied to drift values
- **Signal Generation**: Crossovers and directional changes in drift

### 2. Volatility-Based Pricing
- **High/Low Tracking**: Monitors price volatility using standard deviation
- **Dynamic Adjustment**: Adjusts order prices based on market volatility
- **Risk-Adjusted Entry**: Places orders at optimal prices considering market conditions

### 3. Multi-Level Trailing Stops
- **Activation Ratios**: Multiple levels of profit before trailing activation
- **Callback Rates**: Different callback rates for each activation level
- **Dynamic Tracking**: Continuously updates highest/lowest prices for trailing

### 4. Smart Order Management
- **Time-Based Cancellation**: Cancels orders after specified time intervals
- **Price-Based Cancellation**: Cancels orders when market moves significantly
- **Intelligent Replacement**: Replaces cancelled orders with updated prices

## Entry and Exit Rules

### Entry Conditions

**Long Entry**:
1. Drift crosses from negative to positive OR
2. Drift is positive and increasing OR
3. Drift derivative crosses from negative to positive OR
4. Drift derivative is positive and increasing
5. Price adjusted down by volatility for better entry

**Short Entry**:
1. Drift crosses from positive to negative OR
2. Drift is negative and decreasing OR
3. Drift derivative crosses from positive to negative OR
4. Drift derivative is negative and decreasing
5. Price adjusted up by volatility for better entry

### Exit Conditions

**Stop Loss**:
1. Percentage-based stop loss from entry price
2. ATR-based stop loss using market volatility
3. Multi-level trailing stops with different activation ratios

**Trailing Stop Logic**:
- Activates when profit exceeds activation ratio
- Trails with specified callback rate
- Multiple levels for different profit ranges

## Risk Management Features

### 1. Multiple Stop Loss Types
- **Percentage Stop**: Fixed percentage from entry price
- **ATR Stop**: Dynamic stop based on market volatility
- **Trailing Stop**: Profit-protecting trailing mechanism

### 2. Dynamic Rebalancing
- **Trend Analysis**: Uses linear regression on trend line
- **Beta Filtering**: Rebalances only when trend strength changes significantly
- **Position Adjustment**: Automatically adjusts position based on trend direction

### 3. Order Management
- **Smart Cancellation**: Cancels orders based on time and price movement
- **Volatility Adjustment**: Adjusts order prices based on market conditions
- **Execution Optimization**: Chooses between limit and market orders intelligently

## Performance Analytics

### Real-Time Monitoring
- **Position Tracking**: Continuous monitoring of highest/lowest prices
- **Performance Metrics**: Real-time PnL calculation and tracking
- **Trade Statistics**: Comprehensive trade analysis and reporting

### Graph Generation
- **Indicator Graphs**: Visual representation of drift and other indicators
- **PnL Graphs**: Profit/loss visualization over time
- **Cumulative Performance**: Asset value changes over time
- **Execution Time**: Performance monitoring of strategy execution

## Common Use Cases

### 1. High-Frequency Scalping
```yaml
interval: 1s
minInterval: 1s
window: 2
smootherWindow: 5
pendingMinInterval: 3
trailingActivationRatio: [0.0005, 0.001, 0.005]
trailingCallbackRate: [0.0001, 0.0002, 0.001]
```

### 2. Short-Term Momentum Trading
```yaml
interval: 1m
minInterval: 1s
window: 5
smootherWindow: 15
pendingMinInterval: 10
trailingActivationRatio: [0.002, 0.005, 0.015]
trailingCallbackRate: [0.0005, 0.001, 0.003]
```

### 3. Medium-Term Trend Following
```yaml
interval: 5m
minInterval: 1m
window: 10
smootherWindow: 30
pendingMinInterval: 20
trailingActivationRatio: [0.01, 0.02, 0.05]
trailingCallbackRate: [0.002, 0.005, 0.01]
```

## Best Practices

1. **Interval Selection**: Use shorter intervals for scalping, longer for trend following
2. **Volatility Adjustment**: Tune `hlVarianceMultiplier` based on market volatility
3. **Stop Loss Tuning**: Combine percentage and ATR stops for optimal risk management
4. **Trailing Stop Levels**: Use multiple levels for different profit ranges
5. **Rebalancing**: Enable rebalancing in trending markets, disable in ranging markets
6. **Order Management**: Adjust `pendingMinInterval` based on market liquidity

## Limitations

1. **High Frequency Requirements**: Works best with low-latency connections
2. **Market Dependency**: Performance varies significantly across different market conditions
3. **Parameter Sensitivity**: Requires careful tuning for optimal performance
4. **Computational Intensity**: Multiple indicators and frequent calculations
5. **Slippage Risk**: High-frequency trading may experience slippage in volatile markets

## Troubleshooting

### Common Issues

**No Trades Generated**
- Check if drift values are crossing zero
- Verify that prediction offset has sufficient data
- Ensure market volatility is adequate for signal generation

**Excessive Stop Losses**
- Increase stop loss percentage
- Adjust ATR window for smoother volatility measurement
- Review trailing stop activation ratios

**Poor Order Fill Rates**
- Increase `hlVarianceMultiplier` for better pricing
- Reduce `pendingMinInterval` for faster order updates
- Consider using market orders in volatile conditions

**Frequent Rebalancing**
- Increase `rebalanceFilter` threshold
- Extend `trendWindow` for more stable trend detection
- Consider disabling rebalancing in ranging markets

## Advanced Configuration

### Exit Methods Integration
```yaml
exits:
  - roiStopLoss:
      percentage: 0.35%
  - roiTakeProfit:
      percentage: 0.7%
  - protectiveStopLoss:
      activationRatio: 0.5%
      stopLossRatio: 0.2%
      placeStopOrder: false
```

### Multiple Trailing Stops
```yaml
# Conservative setup
trailingActivationRatio: [0.002, 0.005, 0.02]
trailingCallbackRate: [0.0005, 0.001, 0.005]

# Aggressive setup
trailingActivationRatio: [0.0005, 0.001, 0.005]
trailingCallbackRate: [0.0001, 0.0002, 0.001]
```

## Performance Optimization

### For High-Frequency Trading
- Use 1s intervals with minimal smoothing
- Enable smart order cancellation
- Optimize `pendingMinInterval` for market conditions
- Use limit orders for better pricing

### For Trend Following
- Use longer intervals (5m+) with more smoothing
- Enable rebalancing for trend adaptation
- Use wider trailing stop levels
- Focus on trend strength indicators

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/driftBTC.yaml)
- [Technical Indicators Guide](../../doc/topics/indicators.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
