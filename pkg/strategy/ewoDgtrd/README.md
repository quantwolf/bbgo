# Elliott Wave Oscillator (EWO) Digital Trading Strategy

## Overview

The EWO Digital Trading (ewoDgtrd) strategy is a sophisticated technical analysis strategy that combines the Elliott Wave Oscillator with multiple complementary indicators to identify high-probability trading opportunities. It uses a multi-layered approach including CCI Stochastic, moving averages, ATR-based stop losses, and optional Heikin-Ashi candlesticks for enhanced signal quality.

## How It Works

1. **Elliott Wave Oscillator (EWO)**: Calculates the difference between fast (5-period) and slow (34-period) moving averages as a percentage
2. **Signal Line**: Creates a smoothed signal line from the EWO using configurable window size
3. **CCI Stochastic Filter**: Uses a custom CCI-Stochastic indicator to filter trade signals
4. **Trend Confirmation**: Employs multiple moving averages and candlestick patterns for trend validation
5. **Dynamic Stop Loss**: Implements ATR-based stop losses and take profit mechanisms
6. **Heikin-Ashi Support**: Optional Heikin-Ashi candlesticks for smoother price action analysis

## Key Features

- **Multi-Indicator Confluence**: Combines EWO, CCI Stochastic, moving averages, and ATR
- **Flexible Moving Averages**: Supports EMA, SMA, or Volume-Weighted EMA (VWEMA)
- **Heikin-Ashi Integration**: Optional Heikin-Ashi candlesticks for noise reduction
- **Advanced Stop Loss**: Multiple stop loss mechanisms including ATR-based and percentage-based
- **Dynamic Take Profit**: Peak/bottom tracking with ATR-based profit taking
- **Signal Filtering**: EWO change rate filters to avoid false signals during low volatility
- **Comprehensive Reporting**: Detailed performance analytics and trade categorization
- **Position Management**: Intelligent position sizing using all available balance

## Strategy Logic

### Elliott Wave Oscillator Calculation

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
// EWO = (MA5 / MA34 - 1) * 100
s.ewo = s.ma5.Div(s.ma34).Minus(1.0).Mul(100.)
s.ewoHistogram = s.ma5.Minus(s.ma34)

// Signal line is a moving average of EWO
windowSignal := types.IntervalWindow{Interval: s.Interval, Window: s.SignalWindow}
if s.UseEma {
    sig := &indicator.EWMA{IntervalWindow: windowSignal}
    // ... signal calculation
    s.ewoSignal = sig
}
```
</augment_code_snippet>

### CCI Stochastic Filter

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
type CCISTOCH struct {
    cci        *indicator.CCI
    stoch      *indicator.STOCH
    ma         *indicator.SMA
    filterHigh float64
    filterLow  float64
}

func (inc *CCISTOCH) BuySignal() bool {
    hasGrey := false
    for i := 0; i < inc.ma.Values.Length(); i++ {
        v := inc.ma.Index(i)
        if v > inc.filterHigh {
            return false
        } else if v >= inc.filterLow && v <= inc.filterHigh {
            hasGrey = true
            continue
        } else if v < inc.filterLow {
            return hasGrey
        }
    }
    return false
}
```
</augment_code_snippet>

### Entry Signal Generation

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
longSignal := types.CrossOver(s.ewo, s.ewoSignal)
shortSignal := types.CrossUnder(s.ewo, s.ewoSignal)

// Trend confirmation
bull := clozes.Last(0) > opens.Last(0)
breakThrough := clozes.Last(0) > s.ma5.Last(0) && clozes.Last(0) > s.ma34.Last(0)
breakDown := clozes.Last(0) < s.ma5.Last(0) && clozes.Last(0) < s.ma34.Last(0)

// Final signal with filters
IsBull := bull && breakThrough && s.ccis.BuySignal() && 
          s.ewoChangeRate < s.EwoChangeFilterHigh && s.ewoChangeRate > s.EwoChangeFilterLow
IsBear := !bull && breakDown && s.ccis.SellSignal() && 
          s.ewoChangeRate < s.EwoChangeFilterHigh && s.ewoChangeRate > s.EwoChangeFilterLow

// Entry conditions
if (longSignal.Index(1) && !shortSignal.Last() && IsBull) || lastPrice.Float64() <= buyLine {
    // Place buy order
}
if (shortSignal.Index(1) && !longSignal.Last() && IsBear) || lastPrice.Float64() >= sellLine {
    // Place sell order
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `interval` | interval | Yes | Time interval for analysis (e.g., "15m", "1h") |
| `stoploss` | percentage | Yes | Stop loss percentage from entry price |

### Moving Average Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `useEma` | boolean | No | Use Exponential Moving Average (default: false) |
| `useSma` | boolean | No | Use Simple Moving Average when EMA is false (default: false) |
| `sigWin` | int | Yes | Signal window size for EWO signal line |

### CCI Stochastic Filter
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cciStochFilterHigh` | number | Yes | High filter threshold for CCI Stochastic (e.g., 80) |
| `cciStochFilterLow` | number | Yes | Low filter threshold for CCI Stochastic (e.g., 20) |

### EWO Change Rate Filter
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ewoChangeFilterHigh` | number | Yes | High filter for EWO change rate (1.0 = no filter) |
| `ewoChangeFilterLow` | number | Yes | Low filter for EWO change rate (0.0 = no filter) |

### Advanced Features
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `useHeikinAshi` | boolean | No | Use Heikin-Ashi candlesticks (default: false) |
| `disableShortStop` | boolean | No | Disable stop loss for short positions (default: false) |
| `disableLongStop` | boolean | No | Disable stop loss for long positions (default: false) |
| `record` | boolean | No | Enable detailed trade recording (default: false) |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  ewo_dgtrd:
    symbol: BTCUSDT
    
    # Time interval for analysis
    interval: 15m
    
    # Moving average settings
    useEma: false              # Use EMA (false = use VWEMA)
    useSma: false              # Use SMA when useEma is false
    sigWin: 5                  # Signal line window size
    
    # Risk management
    stoploss: 2%               # 2% stop loss from entry
    disableShortStop: false    # Enable stop loss for shorts
    disableLongStop: false     # Enable stop loss for longs
    
    # Heikin-Ashi candlesticks
    useHeikinAshi: true        # Use Heikin-Ashi for smoother signals
    
    # CCI Stochastic filter
    cciStochFilterHigh: 80     # Upper threshold
    cciStochFilterLow: 20      # Lower threshold
    
    # EWO change rate filter
    ewoChangeFilterHigh: 1.0   # Upper bound (1.0 = no filter)
    ewoChangeFilterLow: 0.0    # Lower bound (0.0 = no filter)
    
    # Debugging and analysis
    record: false              # Disable detailed logging
```

## Strategy Components

### 1. Elliott Wave Oscillator (EWO)
- **Fast MA**: 5-period moving average
- **Slow MA**: 34-period moving average  
- **Calculation**: `(Fast MA / Slow MA - 1) × 100`
- **Signal Line**: Smoothed EWO using configurable window

### 2. CCI Stochastic Indicator
- **CCI**: 28-period Commodity Channel Index
- **Stochastic**: Applied to CCI values
- **Smoothing**: 3-period SMA of Stochastic %D
- **Filtering**: High/low thresholds for signal validation

### 3. Moving Average Types
- **EMA**: Exponential Moving Average (responsive)
- **SMA**: Simple Moving Average (stable)
- **VWEMA**: Volume-Weighted EMA (default, volume-aware)

### 4. Heikin-Ashi Implementation

<augment_code_snippet path="pkg/strategy/ewoDgtrd/heikinashi.go" mode="EXCERPT">
```go
func (inc *HeikinAshi) Update(kline types.KLine) {
    open := kline.Open.Float64()
    cloze := kline.Close.Float64()
    high := kline.High.Float64()
    low := kline.Low.Float64()
    
    newClose := (open + high + low + cloze) / 4.
    newOpen := (inc.Open.Last(0) + inc.Close.Last(0)) / 2.
    
    inc.Close.Update(newClose)
    inc.Open.Update(newOpen)
    inc.High.Update(math.Max(math.Max(high, newOpen), newClose))
    inc.Low.Update(math.Min(math.Min(low, newOpen), newClose))
    inc.Volume.Update(kline.Volume.Float64())
}
```
</augment_code_snippet>

## Entry and Exit Rules

### Entry Conditions

**Long Entry**:
1. EWO crosses above signal line (previous bar) AND no cross under (current bar)
2. Bullish candlestick (Close > Open)
3. Price above both MA5 and MA34
4. CCI Stochastic buy signal
5. EWO change rate within filter bounds
6. OR price touches MA34 - ATR × 3

**Short Entry**:
1. EWO crosses below signal line (previous bar) AND no cross over (current bar)
2. Bearish candlestick (Close < Open)
3. Price below both MA5 and MA34
4. CCI Stochastic sell signal
5. EWO change rate within filter bounds
6. OR price touches MA34 + ATR × 3

### Exit Conditions

**Take Profit**:
1. EWO pivot high/low signals (opposite direction)
2. Price reaches MA34 ± ATR × 2
3. Peak/bottom tracking with ATR-based exits
4. Price moves favorably from peak/bottom by ATR

**Stop Loss**:
1. Percentage-based stop loss from entry price
2. ATR-based stop loss (entry ± ATR)
3. Can be disabled separately for long/short positions

## Risk Management Features

### 1. Multiple Stop Loss Mechanisms
- **Percentage Stop**: Fixed percentage from entry price
- **ATR Stop**: Dynamic stop based on market volatility
- **Selective Disable**: Can disable stops for long or short positions

### 2. Position Sizing
- **Full Balance**: Uses entire available balance for each trade
- **Balance Validation**: Ensures sufficient funds before order placement
- **Market Minimums**: Respects exchange minimum quantity and notional requirements

### 3. Peak/Bottom Tracking
- **Dynamic Tracking**: Continuously updates peak (long) and bottom (short) prices
- **ATR-Based Exits**: Exits when price moves against position by ATR amount
- **Profit Protection**: Locks in profits when favorable moves exceed thresholds

## Performance Analytics

The strategy provides comprehensive trade analytics:

### Trade Categories
- **Peak/Bottom with ATR**: Exits based on peak/bottom tracking
- **CCI Stochastic**: Exits based on CCI Stochastic signals
- **Long/Short Signals**: Exits based on EWO pivot signals
- **MA34 and ATR×2**: Exits based on moving average levels
- **Active Orders**: Exits from new opposing signals
- **Entry Stop Loss**: Exits from percentage/ATR stop losses

### Reporting Metrics
- **Win Rate**: Percentage of profitable trades
- **Average PnL**: Average profit/loss by exit category
- **Trade Counts**: Number of trades per exit type
- **Performance Summary**: Detailed breakdown at strategy shutdown

## Common Use Cases

### 1. Trend Following Setup
```yaml
interval: 1h
useEma: true
sigWin: 10
stoploss: 3%
useHeikinAshi: true
```

### 2. Scalping Setup
```yaml
interval: 5m
useEma: false
useSma: true
sigWin: 3
stoploss: 1%
useHeikinAshi: false
```

### 3. Conservative Setup
```yaml
interval: 4h
useEma: false
sigWin: 8
stoploss: 5%
cciStochFilterHigh: 70
cciStochFilterLow: 30
```

## Best Practices

1. **Timeframe Selection**: Higher timeframes (1h+) for trend following, lower (5m-15m) for scalping
2. **Heikin-Ashi Usage**: Enable for smoother signals in volatile markets
3. **Filter Tuning**: Adjust CCI Stochastic filters based on market conditions
4. **Stop Loss Management**: Consider disabling stops in strong trending markets
5. **EWO Change Filters**: Use to avoid trading during low volatility periods
6. **Signal Window**: Smaller windows for faster signals, larger for stability

## Limitations

1. **Whipsaw Risk**: May generate false signals in sideways markets
2. **Lag**: Multiple indicators create inherent signal lag
3. **Full Position**: Uses entire balance, limiting risk management flexibility
4. **Complexity**: Many parameters require careful tuning
5. **Market Dependency**: Performance varies significantly across different market conditions

## Troubleshooting

### Common Issues

**No Trades Generated**
- Check if all filter conditions are being met
- Verify CCI Stochastic thresholds are appropriate
- Ensure EWO change rate filters aren't too restrictive

**Excessive Losses**
- Increase stop loss percentage
- Tighten CCI Stochastic filters
- Consider enabling Heikin-Ashi for smoother signals

**Poor Signal Quality**
- Adjust signal window size
- Modify EWO change rate filters
- Try different moving average types

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/ewo_dgtrd.yaml)
- [Elliott Wave Theory](../../doc/topics/elliott-wave.md)
- [Technical Indicators Guide](../../doc/topics/indicators.md)
