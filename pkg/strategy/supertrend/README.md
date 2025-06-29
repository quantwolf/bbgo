# SuperTrend Strategy

## Overview

The SuperTrend strategy is a sophisticated trend-following strategy that combines the SuperTrend indicator with additional confirmation signals to identify and trade market trends. It uses Average True Range (ATR) to calculate dynamic support and resistance levels, enhanced by Double DEMA (Double Exponential Moving Average) and Linear Regression filters to reduce false signals and improve trade quality.

## How It Works

1. **SuperTrend Calculation**: Uses ATR and a multiplier to create dynamic support/resistance bands
2. **Trend Identification**: Determines bullish/bearish trends based on price position relative to SuperTrend bands
3. **Signal Confirmation**: Uses Double DEMA and Linear Regression as additional confirmation filters
4. **Position Management**: Opens positions when all signals align and closes based on various exit conditions
5. **Risk Management**: Integrates multiple exit methods including stop-loss, take-profit, and trailing stops
6. **Performance Tracking**: Comprehensive profit statistics and visualization capabilities

## Key Features

- **SuperTrend Indicator**: Dynamic trend identification using ATR-based bands
- **Multi-Signal Confirmation**: Double DEMA and Linear Regression filters for signal quality
- **Flexible Position Sizing**: Supports both fixed quantity and leverage-based sizing
- **Advanced Exit Methods**: Multiple exit strategies including reversed signal exits
- **Noise Filtering**: DEMA crossover detection to filter out market noise
- **Trend Confirmation**: Linear regression slope analysis for trend validation
- **Visualization Support**: Built-in PnL chart generation for performance analysis
- **Comprehensive Risk Management**: Multiple stop-loss and take-profit mechanisms

## Strategy Architecture

### SuperTrend Calculation

<augment_code_snippet path="pkg/strategy/supertrend/strategy.go" mode="EXCERPT">
```go
// SuperTrend calculation using ATR and multiplier
func (s *Strategy) calculateSuperTrend(kline types.KLine) {
    // Get ATR value
    atr := s.atr.Last(0)
    
    // Calculate basic upper and lower bands
    hl2 := (kline.High.Add(kline.Low)).Div(fixedpoint.NewFromFloat(2.0))
    upperBand := hl2.Add(fixedpoint.NewFromFloat(s.SupertrendMultiplier * atr))
    lowerBand := hl2.Sub(fixedpoint.NewFromFloat(s.SupertrendMultiplier * atr))
    
    // Determine trend direction
    if kline.Close.Compare(s.superTrend.Last(0)) > 0 {
        s.currentTrend = types.DirectionUp
    } else {
        s.currentTrend = types.DirectionDown
    }
}
```
</augment_code_snippet>

### Double DEMA Signal

<augment_code_snippet path="pkg/strategy/supertrend/double_dema.go" mode="EXCERPT">
```go
// getDemaSignal get current DEMA signal
func (dd *DoubleDema) getDemaSignal(openPrice float64, closePrice float64) types.Direction {
    var demaSignal types.Direction = types.DirectionNone

    // Bullish breakout: close above both DEMAs but open was not
    if closePrice > dd.fastDEMA.Last(0) && closePrice > dd.slowDEMA.Last(0) && 
       !(openPrice > dd.fastDEMA.Last(0) && openPrice > dd.slowDEMA.Last(0)) {
        demaSignal = types.DirectionUp
    // Bearish breakdown: close below both DEMAs but open was not
    } else if closePrice < dd.fastDEMA.Last(0) && closePrice < dd.slowDEMA.Last(0) && 
              !(openPrice < dd.fastDEMA.Last(0) && openPrice < dd.slowDEMA.Last(0)) {
        demaSignal = types.DirectionDown
    }

    return demaSignal
}
```
</augment_code_snippet>

### Linear Regression Trend

<augment_code_snippet path="pkg/strategy/supertrend/linreg.go" mode="EXCERPT">
```go
// GetSignal get linear regression signal
func (lr *LinReg) GetSignal() types.Direction {
    var lrSignal types.Direction = types.DirectionNone

    switch {
    case lr.Last(0) > 0:  // Positive slope = uptrend
        lrSignal = types.DirectionUp
    case lr.Last(0) < 0:  // Negative slope = downtrend
        lrSignal = types.DirectionDown
    }

    return lrSignal
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `interval` | interval | Yes | Time interval for analysis (e.g., "1m", "5m") |
| `window` | int | Yes | ATR window for SuperTrend calculation |
| `supertrendMultiplier` | float | Yes | Multiplier for ATR in SuperTrend bands |

### Position Sizing
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quantity` | number | No | Fixed quantity per trade |
| `leverage` | float | No | Leverage multiplier for position sizing |

### Signal Filters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fastDEMAWindow` | int | No | Fast DEMA period (default: 144) |
| `slowDEMAWindow` | int | No | Slow DEMA period (default: 169) |
| `linearRegression.interval` | interval | No | Linear regression analysis interval |
| `linearRegression.window` | int | No | Linear regression window period |

### Exit Conditions
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `takeProfitAtrMultiplier` | float | No | Take profit based on ATR multiple |
| `stopLossByTriggeringK` | boolean | No | Set stop loss to triggering candle low |
| `stopByReversedSupertrend` | boolean | No | Exit on reversed SuperTrend signal |
| `stopByReversedDema` | boolean | No | Exit on reversed DEMA signal |
| `stopByReversedLinGre` | boolean | No | Exit on reversed Linear Regression signal |

### Visualization
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `drawGraph` | boolean | No | Enable PnL chart generation |
| `graphPNLPath` | string | No | Path for PnL chart output |
| `graphCumPNLPath` | string | No | Path for cumulative PnL chart output |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  supertrend:
    symbol: BTCUSDT
    
    # Basic SuperTrend settings
    interval: 1m
    window: 220                    # ATR window
    supertrendMultiplier: 10       # ATR multiplier for bands
    
    # Position sizing
    leverage: 1.0                  # Use 1x leverage
    # quantity: 0.5                # Or fixed 0.5 BTC quantity
    
    # DEMA filter settings
    fastDEMAWindow: 28             # Fast DEMA period
    slowDEMAWindow: 170            # Slow DEMA period
    
    # Linear regression confirmation
    linearRegression:
      interval: 1m
      window: 18                   # 18-period linear regression
    
    # Exit method settings
    takeProfitAtrMultiplier: 0     # Disabled (0)
    stopLossByTriggeringK: false   # Don't use triggering candle
    stopByReversedSupertrend: false # Don't exit on reversed signals
    stopByReversedDema: false
    stopByReversedLinGre: false
    
    # Visualization
    drawGraph: true
    graphPNLPath: "./pnl.png"
    graphCumPNLPath: "./cumpnl.png"
    
    # Advanced exit methods
    exits:
    - roiStopLoss:
        percentage: 2%             # 2% stop loss
    - trailingStop:
        callbackRate: 2%           # 2% trailing stop
        minProfit: 10%             # Activate after 10% profit
        interval: 1m
        side: both
        closePosition: 100%
```

## SuperTrend Indicator Explained

### What is SuperTrend?
SuperTrend is a trend-following indicator that uses Average True Range (ATR) to calculate dynamic support and resistance levels. It provides clear buy and sell signals based on price position relative to the trend bands.

### Calculation Method
1. **ATR Calculation**: `ATR = Average True Range over specified window`
2. **Basic Bands**: 
   - `Upper Band = (High + Low) / 2 + (Multiplier × ATR)`
   - `Lower Band = (High + Low) / 2 - (Multiplier × ATR)`
3. **Trend Determination**:
   - **Uptrend**: Price above SuperTrend line (uses Lower Band)
   - **Downtrend**: Price below SuperTrend line (uses Upper Band)

### Signal Generation
- **Buy Signal**: Price crosses above SuperTrend line
- **Sell Signal**: Price crosses below SuperTrend line
- **Trend Continuation**: Price remains on same side of SuperTrend line

## Multi-Signal Confirmation System

### 1. SuperTrend (Primary Signal)
- **Purpose**: Main trend identification
- **Strength**: Clear trend direction with minimal lag
- **Weakness**: Can generate false signals in choppy markets

### 2. Double DEMA Filter
- **Purpose**: Noise reduction and breakout confirmation
- **Logic**: Requires price to break above/below both fast and slow DEMA
- **Benefit**: Filters out minor price fluctuations

### 3. Linear Regression Confirmation
- **Purpose**: Trend strength validation
- **Method**: Analyzes slope of linear regression line
- **Signal**: Positive slope = uptrend, Negative slope = downtrend

## Trading Logic

### Entry Conditions

**Long Entry:**
1. SuperTrend indicates uptrend (price above SuperTrend line)
2. DEMA filter confirms bullish breakout (optional)
3. Linear regression shows positive slope (optional)
4. All enabled filters must align

**Short Entry:**
1. SuperTrend indicates downtrend (price below SuperTrend line)
2. DEMA filter confirms bearish breakdown (optional)
3. Linear regression shows negative slope (optional)
4. All enabled filters must align

### Exit Conditions

**Standard Exits:**
- Reversed SuperTrend signal
- Reversed DEMA signal
- Reversed Linear Regression signal
- ATR-based take profit
- Stop loss at triggering candle low

**Advanced Exits (via exit methods):**
- ROI-based stop loss
- Trailing stops
- Higher high/lower low patterns
- Time-based exits

## Position Sizing

### Fixed Quantity Mode
```yaml
quantity: 0.5  # Fixed 0.5 BTC per trade
```
- Uses the same quantity for all trades
- Simple and predictable position sizing
- Good for consistent risk per trade

### Leverage Mode
```yaml
leverage: 1.0  # 1x leverage
```
- Calculates position size based on account net value
- `Position Size = Account Value × Leverage / Current Price`
- Automatically adjusts to account balance changes

## Risk Management Features

### Built-in Protection
- **ATR-based Stops**: Dynamic stop levels based on volatility
- **Trend Reversal Exits**: Automatic exit when trend changes
- **Multiple Confirmation**: Reduces false signal risk

### Configurable Exits
- **Stop Loss**: Percentage-based or ATR-based
- **Take Profit**: ATR multiple or percentage-based
- **Trailing Stops**: Dynamic profit protection
- **Pattern-based Exits**: Higher high/lower low detection

## Performance Optimization

### Parameter Tuning

**ATR Window:**
- **Shorter (50-100)**: More responsive, more signals
- **Medium (150-250)**: Balanced approach
- **Longer (300+)**: Smoother, fewer false signals

**SuperTrend Multiplier:**
- **Lower (3-7)**: More sensitive, more trades
- **Medium (8-12)**: Balanced sensitivity
- **Higher (15+)**: Less sensitive, stronger trends only

**DEMA Windows:**
- **Fast DEMA**: 20-50 periods for responsiveness
- **Slow DEMA**: 100-200 periods for stability

## Common Use Cases

### 1. Conservative Trend Following
```yaml
window: 300
supertrendMultiplier: 15
fastDEMAWindow: 50
slowDEMAWindow: 200
leverage: 0.5
```

### 2. Aggressive Trend Trading
```yaml
window: 100
supertrendMultiplier: 5
fastDEMAWindow: 20
slowDEMAWindow: 100
leverage: 2.0
```

### 3. Balanced Approach
```yaml
window: 220
supertrendMultiplier: 10
fastDEMAWindow: 28
slowDEMAWindow: 170
leverage: 1.0
```

## Best Practices

1. **Market Selection**: Works best in trending markets
2. **Timeframe**: Higher timeframes generally more reliable
3. **Confirmation**: Use multiple signals for better accuracy
4. **Risk Management**: Always use stop losses and position sizing
5. **Backtesting**: Test parameters on historical data
6. **Market Conditions**: Adjust sensitivity based on volatility

## Limitations

1. **Sideways Markets**: Can generate many false signals in ranging markets
2. **Lag**: Trend-following nature means some delay in signals
3. **Whipsaws**: Rapid trend changes can cause multiple losses
4. **Parameter Sensitivity**: Performance varies significantly with parameter choices
5. **Market Gaps**: Cannot protect against overnight gaps

## Troubleshooting

### Common Issues

**Too Many False Signals**
- Increase SuperTrend multiplier
- Use longer ATR window
- Enable DEMA and Linear Regression filters

**Missing Trend Moves**
- Decrease SuperTrend multiplier
- Use shorter ATR window
- Reduce filter requirements

**Poor Risk/Reward**
- Adjust exit methods
- Use trailing stops
- Optimize take profit levels

**Excessive Drawdown**
- Reduce leverage
- Tighten stop losses
- Use more conservative parameters

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/supertrend.yaml)
- [SuperTrend Indicator Guide](../../doc/topics/supertrend-indicator.md)
- [Trend Following Strategies](../../doc/topics/trend-following.md)
