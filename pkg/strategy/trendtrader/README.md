# Trend Trader Strategy

## Overview

The Trend Trader strategy is a trend-following strategy that identifies and trades breakouts from dynamic trend lines. It uses pivot high and pivot low points to construct converging trend lines (support and resistance) and executes trades when price breaks through these trend lines, indicating potential trend continuation or reversal.

## How It Works

1. **Pivot Point Detection**: Identifies pivot highs and pivot lows using configurable window periods
2. **Trend Line Construction**: Builds dynamic support and resistance trend lines from pivot points
3. **Convergence Analysis**: Detects when support and resistance lines are converging (forming triangular patterns)
4. **Breakout Detection**: Monitors for price breakouts above resistance or below support
5. **Position Management**: Opens long positions on upward breakouts and short positions on downward breakouts
6. **Exit Strategy**: Uses configurable exit methods including trailing stops for profit protection

## Key Features

- **Dynamic Trend Lines**: Automatically constructs trend lines from pivot points
- **Convergence Detection**: Identifies converging trend patterns for higher probability setups
- **Breakout Trading**: Executes trades on confirmed trend line breakouts
- **Flexible Exit Methods**: Supports multiple exit strategies including trailing stops
- **Position Reversal**: Automatically closes opposite positions when new signals occur
- **Market Order Execution**: Uses market orders for immediate execution on breakouts
- **Configurable Parameters**: Adjustable pivot windows and quantity settings

## Strategy Logic

### Pivot Point Analysis

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
s.pivotHigh = standardIndicator.PivotHigh(types.IntervalWindow{
	Interval: s.Interval,
	Window:   int(3. * s.PivotRightWindow), RightWindow: &s.PivotRightWindow})

s.pivotLow = standardIndicator.PivotLow(types.IntervalWindow{
	Interval: s.Interval,
	Window:   int(3. * s.PivotRightWindow), RightWindow: &s.PivotRightWindow})
```
</augment_code_snippet>

### Trend Line Construction

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
// Calculate resistance slope from pivot highs
if line(resistancePrices.Index(2), resistancePrices.Index(1), resistancePrices.Index(0)) < 0 {
	resistanceSlope1 = (resistancePrices.Index(1) - resistancePrices.Index(2)) / resistanceDuration.Index(1)
	resistanceSlope2 = (resistancePrices.Index(0) - resistancePrices.Index(1)) / resistanceDuration.Index(0)
	resistanceSlope = (resistanceSlope1 + resistanceSlope2) / 2.
}

// Calculate support slope from pivot lows
if line(supportPrices.Index(2), supportPrices.Index(1), supportPrices.Index(0)) > 0 {
	supportSlope1 = (supportPrices.Index(1) - supportPrices.Index(2)) / supportDuration.Index(1)
	supportSlope2 = (supportPrices.Index(0) - supportPrices.Index(1)) / supportDuration.Index(0)
	supportSlope = (supportSlope1 + supportSlope2) / 2.
}
```
</augment_code_snippet>

### Breakout Detection and Trading

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
if converge(resistanceSlope, supportSlope) {
	// Calculate current trend line levels
	currentResistance := resistanceSlope*pivotHighDurationCounter + resistancePrices.Last(0)
	currentSupport := supportSlope*pivotLowDurationCounter + supportPrices.Last(0)
	
	// Upward breakout - go long
	if kline.High.Float64() > currentResistance {
		if position.IsShort() {
			s.orderExecutor.ClosePosition(context.Background(), one)
		}
		if position.IsDust(kline.Close) || position.IsClosed() {
			s.placeOrder(context.Background(), types.SideTypeBuy, s.Quantity, symbol)
		}
	
	// Downward breakout - go short
	} else if kline.Low.Float64() < currentSupport {
		if position.IsLong() {
			s.orderExecutor.ClosePosition(context.Background(), one)
		}
		if position.IsDust(kline.Close) || position.IsClosed() {
			s.placeOrder(context.Background(), types.SideTypeSell, s.Quantity, symbol)
		}
	}
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |

### Trend Line Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `trendLine` | object | Yes | Trend line detection settings |
| `trendLine.interval` | interval | Yes | Time interval for analysis (e.g., "30m", "1h") |
| `trendLine.pivotRightWindow` | int | Yes | Right window size for pivot detection |
| `trendLine.quantity` | number | Yes | Fixed quantity for each trade |
| `trendLine.marketOrder` | boolean | No | Enable market order execution (default: true) |

### Exit Methods
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `exits` | array | No | List of exit method configurations |
| `exits[].trailingStop` | object | No | Trailing stop configuration |
| `exits[].trailingStop.callbackRate` | percentage | No | Callback rate for trailing stop |
| `exits[].trailingStop.activationRatio` | percentage | No | Activation ratio for trailing stop |
| `exits[].trailingStop.closePosition` | percentage | No | Percentage of position to close |
| `exits[].trailingStop.minProfit` | percentage | No | Minimum profit before activation |
| `exits[].trailingStop.interval` | interval | No | Update interval for trailing stop |
| `exits[].trailingStop.side` | string | No | Side for trailing stop ("buy" or "sell") |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  trendtrader:
    symbol: BTCUSDT
    
    # Trend line configuration
    trendLine:
      interval: 30m              # 30-minute analysis interval
      pivotRightWindow: 40       # 40-period right window for pivots
      quantity: 0.001            # 0.001 BTC per trade
      marketOrder: true          # Use market orders for execution
    
    # Exit strategies
    exits:
      # Trailing stop for long positions
      - trailingStop:
          callbackRate: 1%       # 1% callback rate
          activationRatio: 1%    # Activate after 1% profit
          closePosition: 100%    # Close entire position
          minProfit: 15%         # Minimum 15% profit required
          interval: 1m           # Update every minute
          side: buy              # For long positions
      
      # Trailing stop for short positions
      - trailingStop:
          callbackRate: 1%       # 1% callback rate
          activationRatio: 1%    # Activate after 1% profit
          closePosition: 100%    # Close entire position
          minProfit: 15%         # Minimum 15% profit required
          interval: 1m           # Update every minute
          side: sell             # For short positions
```

## Strategy Components

### 1. Pivot Point Detection
- **Purpose**: Identify significant price turning points
- **Method**: Uses configurable right window to confirm pivot validity
- **Output**: Series of pivot highs and pivot lows for trend line construction

### 2. Trend Line Construction
- **Resistance Lines**: Built from descending pivot highs
- **Support Lines**: Built from ascending pivot lows
- **Slope Calculation**: Uses time-weighted average of recent pivot slopes

### 3. Convergence Detection
- **Condition**: Support slope > Resistance slope
- **Significance**: Indicates converging trend lines forming triangular patterns
- **Trading Implication**: Higher probability breakout setups

### 4. Breakout Execution
- **Upward Breakout**: Price high exceeds current resistance level
- **Downward Breakout**: Price low falls below current support level
- **Execution**: Market orders for immediate position entry

## Mathematical Foundation

### Line Direction Function
```go
func line(p1, p2, p3 float64) int64 {
	if p1 >= p2 && p2 >= p3 {
		return -1  // Descending line
	} else if p1 <= p2 && p2 <= p3 {
		return +1  // Ascending line
	}
	return 0       // No clear direction
}
```

### Convergence Function
```go
func converge(mr, ms float64) bool {
	return ms > mr  // Support slope > Resistance slope
}
```

### Current Trend Line Level
```
Current Resistance = Resistance Slope × Time Since Last Pivot + Last Pivot High
Current Support = Support Slope × Time Since Last Pivot + Last Pivot Low
```

## Risk Management Features

### 1. Position Reversal
- Automatically closes opposite positions when new signals occur
- Prevents conflicting positions in volatile markets
- Ensures clean position management

### 2. Dust Position Handling
- Checks for dust positions before opening new trades
- Prevents unnecessary small position accumulation
- Maintains clean position sizing

### 3. Exit Method Integration
- Supports multiple exit strategies simultaneously
- Trailing stops for profit protection
- Configurable profit targets and stop losses

## Common Use Cases

### 1. Swing Trading Setup
```yaml
trendLine:
  interval: 4h
  pivotRightWindow: 20
  quantity: 0.01
exits:
  - trailingStop:
      callbackRate: 2%
      minProfit: 10%
```

### 2. Day Trading Setup
```yaml
trendLine:
  interval: 15m
  pivotRightWindow: 10
  quantity: 0.005
exits:
  - trailingStop:
      callbackRate: 0.5%
      minProfit: 5%
```

### 3. Conservative Setup
```yaml
trendLine:
  interval: 1d
  pivotRightWindow: 50
  quantity: 0.002
exits:
  - trailingStop:
      callbackRate: 3%
      minProfit: 20%
```

## Best Practices

1. **Timeframe Selection**: Choose intervals that match your trading style
2. **Pivot Window Tuning**: Larger windows for more significant pivots, smaller for responsiveness
3. **Quantity Management**: Start with smaller quantities to test strategy performance
4. **Exit Strategy**: Always configure appropriate exit methods for risk management
5. **Market Conditions**: Works best in trending markets with clear breakout patterns
6. **Backtesting**: Test different parameter combinations before live trading

## Limitations

1. **Sideways Markets**: May generate false signals in ranging markets
2. **Whipsaw Risk**: Rapid reversals can cause multiple small losses
3. **Lag**: Pivot-based approach has inherent lag in signal generation
4. **Market Orders**: May experience slippage during volatile periods

## Troubleshooting

### Common Issues

**No Trades Generated**
- Check if pivot windows are appropriate for timeframe
- Verify convergence conditions are being met
- Ensure sufficient price movement for breakouts

**Excessive Trading**
- Increase pivot right window for more significant pivots
- Add minimum profit requirements
- Consider longer timeframes

**Poor Performance**
- Review exit strategy settings
- Adjust trailing stop parameters
- Consider market conditions and volatility

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/trendtrader.yaml)
- [Supertrend Strategy](../supertrend/README.md) - Alternative trend-following approach
- [Exit Methods Documentation](../../doc/topics/exit-methods.md)
