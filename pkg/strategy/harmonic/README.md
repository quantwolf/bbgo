# Harmonic Strategy

## Overview

The Harmonic strategy is an advanced technical analysis strategy that combines harmonic pattern recognition with Hidden Markov Model (HMM) signal filtering. It uses the SHARK harmonic pattern indicator to identify potential reversal points and employs HMM to denoise trading signals, providing a sophisticated approach to pattern-based trading with statistical signal processing.

## How It Works

1. **SHARK Pattern Detection**: Identifies SHARK harmonic patterns in price action using Fibonacci ratios
2. **Signal Generation**: Calculates long and short scores based on pattern completion
3. **HMM Filtering**: Applies Hidden Markov Model to denoise and filter trading signals
4. **State Management**: Uses a state machine approach with long (1), neutral (0), and short (-1) states
5. **Position Management**: Opens and closes positions based on filtered signals and state transitions
6. **Performance Analytics**: Comprehensive profit tracking and visualization capabilities

## Key Features

- **SHARK Harmonic Pattern Recognition**: Advanced pattern detection using Fibonacci retracements
- **Hidden Markov Model Filtering**: Statistical signal denoising for improved accuracy
- **Multi-State Trading System**: Long, neutral, and short state management
- **Real-Time Pattern Scoring**: Continuous evaluation of bullish and bearish patterns
- **Comprehensive Analytics**: Detailed profit reporting and performance visualization
- **Interactive Commands**: Real-time PnL and cumulative profit chart generation
- **Backtesting Support**: Full backtesting capabilities with graph generation
- **Exit Method Integration**: Compatible with various exit strategies

## Strategy Logic

### SHARK Pattern Detection

<augment_code_snippet path="pkg/strategy/harmonic/shark.go" mode="EXCERPT">
```go
func (inc SHARK) SharkLong(highs, lows floats.Slice, p float64, lookback int) float64 {
    score := 0.
    for x := 5; x < lookback; x++ {
        if lows.Index(x-1) > lows.Index(x) && lows.Index(x) < lows.Index(x+1) {
            X := lows.Index(x)
            for a := 4; a < x; a++ {
                if highs.Index(a-1) < highs.Index(a) && highs.Index(a) > highs.Index(a+1) {
                    A := highs.Index(a)
                    XA := math.Abs(X - A)
                    hB := A - 0.382*XA  // 38.2% Fibonacci retracement
                    lB := A - 0.618*XA  // 61.8% Fibonacci retracement
                    // Pattern validation continues...
                }
            }
        }
    }
    return score
}
```
</augment_code_snippet>

### Hidden Markov Model Implementation

<augment_code_snippet path="pkg/strategy/harmonic/strategy.go" mode="EXCERPT">
```go
func hmm(y_t []float64, x_t []float64, l int) float64 {
    // HMM implementation for signal filtering
    // States: -1 (short), 0 (neutral), 1 (long)
    
    for n := 2; n <= len(x_t); n++ {
        for j := -1; j <= 1; j++ {
            // Calculate transition probabilities
            for i := -1; i <= 1; i++ {
                transitProb := transitProbability(i, j)
                observeProb := observeDistribution(y_t[n-1], float64(j))
                // Update alpha values for each state
            }
        }
    }
    
    // Return most probable state
    if maximum[0] == long {
        return 1
    } else if maximum[0] == short {
        return -1
    }
    return 0
}
```
</augment_code_snippet>

### Trading Signal Logic

<augment_code_snippet path="pkg/strategy/harmonic/strategy.go" mode="EXCERPT">
```go
s.session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {
    log.Infof("shark score: %f, current price: %f", s.shark.Last(0), kline.Close.Float64())
    
    nextState := hmm(s.shark.Array(s.Window), states.Array(s.Window), s.Window)
    states.Update(nextState)
    log.Infof("Denoised signal via HMM: %f", states.Last(0))
    
    if states.Length() < s.Window {
        return
    }
    
    // Close position when signal becomes neutral
    if s.Position.IsOpened(kline.Close) && states.Mean(5) == 0 {
        s.orderExecutor.ClosePosition(ctx, fixedpoint.One)
    }
    
    // Open long position
    if states.Mean(5) == 1 && direction != 1 {
        s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
            Symbol:   s.Symbol,
            Side:     types.SideTypeBuy,
            Quantity: s.Quantity,
            Type:     types.OrderTypeMarket,
            Tag:      "sharkLong",
        })
    }
    
    // Open short position
    if states.Mean(5) == -1 && direction != -1 {
        s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
            Symbol:   s.Symbol,
            Side:     types.SideTypeSell,
            Quantity: s.Quantity,
            Type:     types.OrderTypeMarket,
            Tag:      "sharkShort",
        })
    }
}))
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `interval` | interval | Yes | Time interval for analysis (e.g., "1s", "1m") |
| `window` | int | Yes | Lookback window for pattern detection |
| `quantity` | number | Yes | Fixed quantity for each trade |

### Visualization Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `drawGraph` | boolean | No | Enable graph generation during backtesting |
| `graphPNLPath` | string | No | Path for PnL percentage graph output |
| `graphCumPNLPath` | string | No | Path for cumulative PnL graph output |

### Exit Methods
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `exits` | array | No | Array of exit method configurations |

### Accumulated Profit Report
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `accumulatedProfitReport.accumulatedProfitMAWindow` | int | No | SMA window for accumulated profit (default: 60) |
| `accumulatedProfitReport.intervalWindow` | int | No | Interval window in days (default: 7) |
| `accumulatedProfitReport.numberOfInterval` | int | No | Number of intervals to output (default: 1) |
| `accumulatedProfitReport.tsvReportPath` | string | No | Path for TSV report output |
| `accumulatedProfitReport.accumulatedDailyProfitWindow` | int | No | Daily profit accumulation window (default: 7) |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  harmonic:
    symbol: BTCUSDT
    
    # Time settings
    interval: 1s              # 1-second analysis interval
    window: 60                # 60-period lookback window
    
    # Position sizing
    quantity: 0.005           # 0.005 BTC per trade
    
    # Visualization (for backtesting)
    drawGraph: true           # Enable graph generation
    graphPNLPath: "./pnl.png"
    graphCumPNLPath: "./cumpnl.png"
    
    # Exit methods
    exits:
    - roiStopLoss:
        percentage: 0.02      # 2% stop loss
    - roiTakeProfit:
        percentage: 0.05      # 5% take profit
    
    # Accumulated profit reporting
    accumulatedProfitReport:
      accumulatedProfitMAWindow: 60
      intervalWindow: 7
      numberOfInterval: 4
      tsvReportPath: "./profit_report.tsv"
      accumulatedDailyProfitWindow: 7
```

## SHARK Harmonic Pattern

### Pattern Structure
The SHARK pattern is a 5-point harmonic pattern (X-A-B-C-D) with specific Fibonacci ratios:

**Bullish SHARK Pattern:**
- **Point X**: Initial low
- **Point A**: High after X
- **Point B**: Retracement of XA (38.2% - 61.8%)
- **Point C**: Extension of AB (113% - 161.8%)
- **Point D**: Retracement of XC (88.6% - 113%) and BC (161.8% - 224%)

**Bearish SHARK Pattern:**
- **Point X**: Initial high
- **Point A**: Low after X
- **Point B**: Retracement of XA (38.2% - 61.8%)
- **Point C**: Extension of AB (113% - 161.8%)
- **Point D**: Extension of XC (88.6% - 113%) and BC (161.8% - 224%)

### Fibonacci Ratios Used
- **B Point**: 0.382 - 0.618 of XA
- **C Point**: 1.13 - 1.618 of AB
- **D Point**: 0.886 - 1.13 of XC and 1.618 - 2.24 of BC

## Hidden Markov Model

### State Definition
- **State 1**: Long/Bullish signal
- **State 0**: Neutral signal
- **State -1**: Short/Bearish signal

### Transition Probabilities
- **Same State**: 0.99 (high probability of staying in current state)
- **State Change**: 0.01 (low probability of state transition)

### Observation Model
The observation distribution evaluates the consistency between the SHARK indicator value and the current state:
- **Consistent**: Returns 1.0 (high probability)
- **Inconsistent**: Returns 0.0 (low probability)

## Trading Logic

### Entry Conditions

**Long Entry:**
1. HMM filtered state mean over 5 periods equals 1
2. Current position is not already long
3. SHARK long score indicates bullish pattern completion

**Short Entry:**
1. HMM filtered state mean over 5 periods equals -1
2. Current position is not already short
3. SHARK short score indicates bearish pattern completion

### Exit Conditions

**Neutral Signal Exit:**
1. HMM filtered state mean over 5 periods equals 0
2. Position is currently open
3. Closes 100% of the position

**Exit Method Integration:**
- Compatible with ROI-based stop loss and take profit
- Supports trailing stops and other exit strategies
- Can combine multiple exit methods

## Performance Analytics

### Real-Time Monitoring
- **SHARK Score Tracking**: Continuous monitoring of pattern scores
- **HMM State Logging**: Real-time state transition logging
- **Position Tracking**: Detailed position and profit monitoring

### Accumulated Profit Report
- **Daily Profit Calculation**: Tracks daily profit/loss
- **Moving Average Analysis**: SMA of accumulated profits
- **Win Rate Tracking**: Daily win ratio statistics
- **Profit Factor Analysis**: Risk-adjusted performance metrics
- **Trade Count Monitoring**: Number of trades per period

### Visualization Features
- **PnL Percentage Charts**: Trade-by-trade profit percentage
- **Cumulative PnL Charts**: Total portfolio value over time
- **Interactive Commands**: `/pnl` and `/cumpnl` for real-time charts
- **TSV Report Export**: Detailed performance data export

## Interactive Commands

### Real-Time Chart Generation
```
/pnl      - Generate PnL percentage chart
/cumpnl   - Generate cumulative PnL chart
```

These commands generate real-time performance charts that can be shared via messaging platforms.

## Common Use Cases

### 1. High-Frequency Pattern Trading
```yaml
interval: 1s
window: 30
quantity: 0.001
drawGraph: false
```

### 2. Medium-Term Pattern Recognition
```yaml
interval: 5m
window: 100
quantity: 0.01
exits:
- roiStopLoss:
    percentage: 0.03
- roiTakeProfit:
    percentage: 0.08
```

### 3. Conservative Pattern Trading
```yaml
interval: 15m
window: 200
quantity: 0.005
exits:
- roiStopLoss:
    percentage: 0.02
- roiTakeProfit:
    percentage: 0.06
- trailingStop:
    callbackRate: 0.01
```

## Best Practices

1. **Window Size Selection**: Larger windows for more reliable patterns, smaller for responsiveness
2. **Interval Optimization**: Match interval to trading style and market volatility
3. **Risk Management**: Always use exit methods for risk control
4. **Pattern Validation**: Monitor SHARK scores to understand pattern strength
5. **HMM Tuning**: Consider adjusting transition probabilities for different markets
6. **Backtesting**: Use graph generation to visualize strategy performance

## Limitations

1. **Pattern Dependency**: Relies on harmonic pattern formation which may be rare
2. **Computational Intensity**: Complex pattern detection requires significant processing
3. **Market Conditions**: Performance varies significantly across different market regimes
4. **Lag**: Pattern completion detection introduces inherent lag
5. **False Signals**: HMM filtering reduces but doesn't eliminate false signals

## Troubleshooting

### Common Issues

**No Trading Signals**
- Check if window size allows sufficient pattern detection
- Verify SHARK scores are being generated
- Ensure HMM states are transitioning properly

**Excessive Trading**
- Increase window size for more stable signals
- Adjust HMM transition probabilities
- Add additional exit methods for risk control

**Poor Pattern Recognition**
- Verify price data quality and completeness
- Check Fibonacci ratio calculations
- Monitor SHARK score generation logs

**HMM State Issues**
- Review observation distribution logic
- Check transition probability settings
- Verify state array initialization

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/harmonic.yaml)
- [Harmonic Pattern Trading Guide](../../doc/topics/harmonic-patterns.md)
- [Hidden Markov Models in Trading](../../doc/topics/hmm-trading.md)
