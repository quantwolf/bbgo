# Factor Zoo Strategy

## Overview

The Factor Zoo strategy is an advanced quantitative trading strategy that implements a multi-factor model using machine learning techniques. It combines multiple financial factors (price-volume divergence, momentum, mean reversion, drift, and volume momentum) with logistic regression to predict future price movements and make trading decisions. This strategy represents a sophisticated approach to algorithmic trading based on academic research in quantitative finance.

## How It Works

1. **Factor Calculation**: Computes multiple financial factors from market data
2. **Feature Engineering**: Transforms raw market data into predictive features
3. **Machine Learning**: Uses logistic regression to model the relationship between factors and returns
4. **Signal Generation**: Generates binary trading signals (buy/sell) based on model predictions
5. **Position Management**: Executes trades based on predicted probabilities and rolling averages
6. **Risk Control**: Integrates with exit methods for comprehensive risk management

## Key Features

- **Multi-Factor Model**: Combines 5 different financial factors for comprehensive market analysis
- **Machine Learning Integration**: Uses logistic regression for predictive modeling
- **Real-Time Adaptation**: Continuously updates model with new market data
- **Binary Classification**: Converts continuous predictions to actionable trading signals
- **Rolling Prediction**: Uses time-series rolling averages for signal smoothing
- **Exit Method Integration**: Compatible with various risk management strategies
- **Academic Foundation**: Based on established quantitative finance research

## Strategy Architecture

### Factor Components

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
type Linear struct {
    // Xs (input), factors & indicators
    divergence *factorzoo.PVD   // price volume divergence
    reversion  *factorzoo.PMR   // price mean reversion
    momentum   *factorzoo.MOM   // price momentum from paper, alpha 101
    drift      *indicator.Drift // GBM (Geometric Brownian Motion)
    volume     *factorzoo.VMOM  // quarterly volume momentum

    // Y (output), internal rate of return
    irr *factorzoo.RR
}
```
</augment_code_snippet>

### Machine Learning Pipeline

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
// Prepare feature matrix (X) and target vector (Y)
a := []floats.Slice{
    s.divergence.Values[len(s.divergence.Values)-s.Window-2 : len(s.divergence.Values)-2],
    s.reversion.Values[len(s.reversion.Values)-s.Window-2 : len(s.reversion.Values)-2],
    s.drift.Values[len(s.drift.Values)-s.Window-2 : len(s.drift.Values)-2],
    s.momentum.Values[len(s.momentum.Values)-s.Window-2 : len(s.momentum.Values)-2],
    s.volume.Values[len(s.volume.Values)-s.Window-2 : len(s.volume.Values)-2],
}

// Binary target: convert returns to 0/1 classification
b := []floats.Slice{filter(s.irr.Values[len(s.irr.Values)-s.Window-1:len(s.irr.Values)-1], binary)}

// Train logistic regression model
model := types.LogisticRegression(x, y[0], s.Window, 8000, 0.0001)

// Make prediction with current factor values
input := []float64{
    s.divergence.Last(0),
    s.reversion.Last(0),
    s.drift.Last(0),
    s.momentum.Last(0),
    s.volume.Last(0),
}
pred := model.Predict(input)
```
</augment_code_snippet>

### Trading Logic

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
// Use rolling average of predictions for signal smoothing
predLst.Update(pred)

// Trading decisions based on prediction vs. rolling mean
if pred > predLst.Mean() {
    if position.IsShort() {
        s.ClosePosition(ctx, one)
        s.placeMarketOrder(ctx, types.SideTypeBuy, qty, symbol)
    } else if position.IsClosed() {
        s.placeMarketOrder(ctx, types.SideTypeBuy, qty, symbol)
    }
} else if pred < predLst.Mean() {
    if position.IsLong() {
        s.ClosePosition(ctx, one)
        s.placeMarketOrder(ctx, types.SideTypeSell, qty, symbol)
    } else if position.IsClosed() {
        s.placeMarketOrder(ctx, types.SideTypeSell, qty, symbol)
    }
}
```
</augment_code_snippet>

## Financial Factors Explained

### 1. Price Volume Divergence (PVD)

<augment_code_snippet path="pkg/strategy/factorzoo/factors/price_volume_divergence.go" mode="EXCERPT">
```go
// Measures divergence between price and volume
// Negative correlation indicates divergence
func (inc *PVD) Update(price float64, volume float64) {
    inc.Prices.Update(price)
    inc.Volumes.Update(volume)
    if inc.Prices.Length() >= inc.Window && inc.Volumes.Length() >= inc.Window {
        divergence := -types.Correlation(inc.Prices, inc.Volumes, inc.Window)
        inc.Values.Push(divergence)
    }
}
```
</augment_code_snippet>

**Theory**: When price and volume move in opposite directions, it often signals potential reversals. High negative correlation suggests strong divergence.

### 2. Price Mean Reversion (PMR)

<augment_code_snippet path="pkg/strategy/factorzoo/factors/price_mean_reversion.go" mode="EXCERPT">
```go
// Measures tendency of price to revert to moving average
func (inc *PMR) Update(price float64) {
    inc.SMA.Update(price)
    if inc.SMA.Length() >= inc.Window {
        reversion := inc.SMA.Last(0) / price  // SMA/Price ratio
        inc.Values.Push(reversion)
    }
}
```
</augment_code_snippet>

**Theory**: Assumes prices tend to revert to their moving average. Values > 1 suggest price is below average (potential buy), values < 1 suggest price is above average (potential sell).

### 3. Momentum (MOM)

<augment_code_snippet path="pkg/strategy/factorzoo/factors/momentum.go" mode="EXCERPT">
```go
// Gap jump momentum - measures opening price gaps
func (inc *MOM) Update(open, close float64) {
    inc.opens.Update(open)
    inc.closes.Update(close)
    if inc.opens.Length() >= inc.Window && inc.closes.Length() >= inc.Window {
        gap := inc.opens.Last(0)/inc.closes.Index(1) - 1  // Gap ratio
        inc.Values.Push(gap)
    }
}
```
</augment_code_snippet>

**Theory**: Large gaps between current open and previous close indicate strong momentum. Positive gaps suggest bullish momentum, negative gaps suggest bearish momentum.

### 4. Drift (Geometric Brownian Motion)

Uses the standard BBGO Drift indicator to measure the underlying trend direction and strength in price movements.

### 5. Volume Momentum (VMOM)

Measures momentum in trading volume over quarterly periods, indicating institutional interest and market participation changes.

### 6. Return Rate (RR) - Target Variable

<augment_code_snippet path="pkg/strategy/factorzoo/factors/return_rate.go" mode="EXCERPT">
```go
// Simple return rate calculation
func (inc *RR) Update(price float64) {
    inc.prices.Update(price)
    irr := inc.prices.Last(0)/inc.prices.Index(1) - 1  // Period return
    inc.Values.Push(irr)
}
```
</augment_code_snippet>

**Purpose**: Serves as the target variable (Y) for the machine learning model, converted to binary classification (positive return = 1, negative return = 0).

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `linear.interval` | interval | Yes | Time interval for analysis (e.g., "1d", "4h") |
| `linear.window` | int | Yes | Lookback window for factor calculation and model training |
| `linear.quantity` | number | Yes | Fixed quantity for each trade |

### Advanced Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `linear.marketOrder` | boolean | No | Enable market order execution (default: true) |
| `linear.stopEMARange` | number | No | EMA-based stop loss range |
| `linear.stopEMA` | object | No | EMA configuration for stop loss |

### Exit Methods
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `exits` | array | No | Array of exit method configurations |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  factorzoo:
    symbol: BTCUSDT
    
    # Linear regression model configuration
    linear:
      interval: 1d              # Daily analysis
      window: 5                 # 5-day lookback window
      quantity: 0.01            # 0.01 BTC per trade
      marketOrder: true         # Use market orders
      
    # Risk management
    exits:
    - trailingStop:
        callbackRate: 1%        # 1% trailing stop
        activationRatio: 1%     # Activate after 1% profit
        closePosition: 100%     # Close entire position
        minProfit: 15%          # Minimum 15% profit before activation
        interval: 1m            # Check every minute
        side: buy               # For long positions
    - trailingStop:
        callbackRate: 1%
        activationRatio: 1%
        closePosition: 100%
        minProfit: 15%
        interval: 1m
        side: sell              # For short positions
```

## Factor Configuration Details

### Default Factor Windows
- **PVD (Price Volume Divergence)**: 60 periods
- **PMR (Price Mean Reversion)**: 60 periods  
- **MOM (Momentum)**: 1 period (gap detection)
- **Drift**: 7 periods
- **VMOM (Volume Momentum)**: 90 periods
- **RR (Return Rate)**: 2 periods

### Model Parameters
- **Training Iterations**: 8000
- **Learning Rate**: 0.0001
- **Classification Threshold**: Rolling mean of predictions

## Machine Learning Details

### Logistic Regression Model
The strategy uses logistic regression to model the probability of positive returns:

```
P(Return > 0) = 1 / (1 + e^(-(β₀ + β₁×PVD + β₂×PMR + β₃×MOM + β₄×Drift + β₅×VMOM)))
```

### Feature Engineering
1. **Normalization**: Factors are calculated as standardized values
2. **Windowing**: Uses rolling windows to capture temporal patterns
3. **Binary Target**: Converts continuous returns to binary classification
4. **Lag Structure**: Uses lagged factor values to predict future returns

### Model Training
- **Online Learning**: Model retrains on each new data point
- **Rolling Window**: Uses fixed window size for training data
- **Gradient Descent**: Optimizes model parameters iteratively

## Trading Logic

### Signal Generation
1. **Factor Calculation**: Compute all 5 factors for current period
2. **Model Prediction**: Generate probability prediction (0-1 scale)
3. **Signal Smoothing**: Compare prediction to rolling average
4. **Decision Making**: Trade based on prediction vs. threshold

### Position Management
- **Long Signal**: Prediction > Rolling Mean → Buy
- **Short Signal**: Prediction < Rolling Mean → Sell
- **Position Reversal**: Automatically closes opposite positions
- **Market Orders**: Uses market orders for immediate execution

### Risk Management
- **Exit Methods**: Integrates with trailing stops and other exit strategies
- **Position Sizing**: Fixed quantity per trade
- **Order Cancellation**: Gracefully cancels orders before new trades

## Performance Considerations

### Computational Complexity
- **Factor Calculation**: O(n) for each factor per period
- **Model Training**: O(n×m×i) where n=samples, m=features, i=iterations
- **Memory Usage**: Stores rolling windows of factor values

### Market Conditions
- **Trending Markets**: Momentum and drift factors perform well
- **Sideways Markets**: Mean reversion factors provide value
- **High Volatility**: Volume divergence factors capture regime changes
- **Low Liquidity**: May require larger windows for stable signals

## Common Use Cases

### 1. Daily Systematic Trading
```yaml
linear:
  interval: 1d
  window: 10
  quantity: 0.1
```

### 2. Intraday Factor Trading
```yaml
linear:
  interval: 4h
  window: 20
  quantity: 0.05
```

### 3. Conservative Long-Term
```yaml
linear:
  interval: 1w
  window: 4
  quantity: 0.2
exits:
- roiStopLoss:
    percentage: 0.1
```

## Best Practices

1. **Window Selection**: Larger windows for stability, smaller for responsiveness
2. **Factor Validation**: Monitor individual factor performance
3. **Model Monitoring**: Track prediction accuracy over time
4. **Risk Management**: Always use exit methods for downside protection
5. **Market Regime**: Consider different parameters for different market conditions
6. **Backtesting**: Extensive historical testing before live deployment

## Limitations

1. **Model Assumptions**: Assumes linear relationships between factors and returns
2. **Overfitting Risk**: Small windows may lead to overfitting
3. **Factor Stability**: Factor effectiveness may change over time
4. **Market Regime Changes**: Model may not adapt quickly to new regimes
5. **Transaction Costs**: High-frequency retraining may increase costs
6. **Data Requirements**: Requires sufficient historical data for training

## Troubleshooting

### Common Issues

**Poor Prediction Accuracy**
- Increase window size for more stable training
- Check factor correlation and multicollinearity
- Validate data quality and completeness

**Excessive Trading**
- Increase prediction threshold sensitivity
- Use longer intervals for signal generation
- Add minimum holding period constraints

**Factor Calculation Errors**
- Verify sufficient data for factor windows
- Check for missing or invalid price/volume data
- Monitor factor value ranges for anomalies

**Model Training Failures**
- Reduce learning rate for convergence
- Increase iteration count for complex patterns
- Check for numerical stability issues

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/factorzoo.yaml)
- [Quantitative Factor Models](../../doc/topics/factor-models.md)
- [Machine Learning in Trading](../../doc/topics/ml-trading.md)
