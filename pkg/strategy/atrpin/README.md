# ATR Pin Strategy

## Overview

The ATR Pin strategy is a market-making strategy that uses the Average True Range (ATR) indicator to dynamically place buy and sell orders around the current market price. It "pins" orders at calculated distances from the market price based on volatility, creating a spread that adapts to market conditions. The strategy aims to profit from price oscillations while managing risk through volatility-based positioning.

## How It Works

1. **ATR Calculation**: Calculates the Average True Range over a specified window to measure market volatility
2. **Price Range Determination**: Multiplies ATR by a configurable multiplier to determine order placement distance
3. **Dynamic Order Placement**: Places buy orders below and sell orders above the current market price
4. **Position Management**: Automatically places take-profit orders when holding positions
5. **Risk Protection**: Implements minimum price range protection and dust quantity filtering
6. **Continuous Adjustment**: Cancels and replaces orders on each interval to maintain optimal positioning

## Key Features

- **Volatility-Adaptive Positioning**: Uses ATR to adjust order placement based on market volatility
- **Dynamic Market Making**: Continuously places buy and sell orders around market price
- **Automatic Take Profit**: Places immediate take-profit orders when holding positions
- **Risk Protection**: Multiple safeguards including minimum price range and balance validation
- **Dust Quantity Filtering**: Prevents placement of orders below exchange minimums
- **Flexible Position Sizing**: Supports both fixed quantity and percentage-based sizing
- **Balance-Based Take Profit**: Optional feature to manage positions based on expected balance

## Strategy Logic

### ATR-Based Price Range Calculation

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// Calculate ATR and apply multiplier
lastAtr := atr.Last(0)

// Protection: ensure ATR is at least the current candle range
if lastAtr <= k.High.Sub(k.Low).Float64() {
    lastAtr = k.High.Sub(k.Low).Float64()
}

priceRange := fixedpoint.NewFromFloat(lastAtr * s.Multiplier)

// Apply minimum price range protection
priceRange = fixedpoint.Max(priceRange, k.Close.Mul(s.MinPriceRange))
```
</augment_code_snippet>

### Dynamic Order Placement

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// Calculate bid and ask prices based on current ticker and price range
bidPrice := fixedpoint.Max(ticker.Buy.Sub(priceRange), s.Market.TickSize)
askPrice := ticker.Sell.Add(priceRange)

// Calculate quantities for each order
bidQuantity := s.QuantityOrAmount.CalculateQuantity(bidPrice)
askQuantity := s.QuantityOrAmount.CalculateQuantity(askPrice)
```
</augment_code_snippet>

### Position Management and Take Profit

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// Check if we have a position that needs take profit
position := s.Strategy.OrderExecutor.Position()
base := position.GetBase()

// Optional: use expected base balance for position calculation
if s.TakeProfitByExpectedBaseBalance {
    base = baseBalance.Available.Sub(s.ExpectedBaseBalance)
}

// Determine take profit side and price
side := types.SideTypeSell
takerPrice := ticker.Buy
if base.Sign() < 0 {
    side = types.SideTypeBuy
    takerPrice = ticker.Sell
}

// Place take profit order if position is not dust
positionQuantity := base.Abs()
if !s.Market.IsDustQuantity(positionQuantity, takerPrice) {
    // Submit take profit order at current market price
    orderForms = append(orderForms, types.SubmitOrder{
        Symbol:      s.Symbol,
        Type:        types.OrderTypeLimit,
        Side:        side,
        Price:       takerPrice,
        Quantity:    positionQuantity,
        Tag:         "takeProfit",
    })
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `interval` | interval | Yes | Time interval for ATR calculation (e.g., "5m", "1h") |
| `window` | int | Yes | ATR calculation window (number of periods) |
| `multiplier` | float | Yes | Multiplier applied to ATR for price range calculation |
| `minPriceRange` | percentage | Yes | Minimum price range as percentage of current price |

### Position Sizing
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quantity` | number | No | Fixed quantity for each order |
| `amount` | number | No | Fixed quote amount for each order |

### Advanced Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `takeProfitByExpectedBaseBalance` | boolean | No | Use expected base balance for position calculation |
| `expectedBaseBalance` | number | No | Expected base currency balance for position management |

## Configuration Example

```yaml
exchangeStrategies:
- on: binance
  atrpin:
    symbol: BTCUSDT
    
    # ATR calculation settings
    interval: 5m              # 5-minute intervals for ATR
    window: 14                # 14-period ATR window
    multiplier: 10.0          # 10x ATR multiplier for price range
    minPriceRange: 0.5%       # Minimum 0.5% price range protection
    
    # Position sizing
    amount: 100               # $100 per order
    
    # Advanced position management (optional)
    takeProfitByExpectedBaseBalance: false
    expectedBaseBalance: 0.0
```

## ATR (Average True Range) Explained

### What is ATR?
ATR measures market volatility by calculating the average of true ranges over a specified period. The true range is the maximum of:
- Current High - Current Low
- |Current High - Previous Close|
- |Current Low - Previous Close|

### How ATR Pin Uses ATR
1. **Volatility Measurement**: ATR quantifies current market volatility
2. **Dynamic Spacing**: Higher ATR = wider order spacing, Lower ATR = tighter spacing
3. **Risk Adaptation**: Automatically adjusts to changing market conditions
4. **Profit Optimization**: Balances capture rate vs. profit per trade

### ATR Multiplier Impact
- **Low Multiplier (1-5)**: Tight spreads, higher fill rate, lower profit per trade
- **Medium Multiplier (5-15)**: Balanced approach for most market conditions
- **High Multiplier (15+)**: Wide spreads, lower fill rate, higher profit per trade

## Strategy Workflow

### 1. Market Data Processing
- Subscribes to K-line data for ATR calculation
- Subscribes to 1-minute data for responsive order management
- Calculates ATR using specified window and interval

### 2. Order Management Cycle
1. **Cancel Existing Orders**: Gracefully cancels all open orders
2. **Update Account**: Refreshes balance information
3. **Calculate ATR**: Gets latest ATR value with protection
4. **Determine Price Range**: Applies multiplier and minimum range protection
5. **Query Market**: Gets current bid/ask prices
6. **Check Position**: Evaluates current position for take profit

### 3. Order Placement Logic
- **Take Profit Priority**: If position exists, place take profit order first
- **Market Making Orders**: If no position or position is dust, place bid/ask orders
- **Balance Validation**: Ensures sufficient balance for each order
- **Dust Filtering**: Prevents orders below exchange minimums

### 4. Risk Management
- **Minimum Price Range**: Protects against extremely low volatility
- **Balance Checks**: Validates available balance before order placement
- **Dust Quantity Protection**: Filters out orders below minimum thresholds
- **Tick Size Compliance**: Ensures bid price meets minimum tick requirements

## Price Range Protection

### ATR Protection
```go
// Ensure ATR is at least the current candle range
if lastAtr <= k.High.Sub(k.Low).Float64() {
    lastAtr = k.High.Sub(k.Low).Float64()
}
```

### Minimum Range Protection
```go
// Apply minimum price range (e.g., 0.5% of current price)
priceRange = fixedpoint.Max(priceRange, k.Close.Mul(s.MinPriceRange))
```

### Tick Size Protection
```go
// Ensure bid price is at least one tick size
bidPrice := fixedpoint.Max(ticker.Buy.Sub(priceRange), s.Market.TickSize)
```

## Position Management Modes

### Standard Mode
Uses the strategy's internal position tracking:
```yaml
takeProfitByExpectedBaseBalance: false
```

### Expected Balance Mode
Uses expected base balance for position calculation:
```yaml
takeProfitByExpectedBaseBalance: true
expectedBaseBalance: 1.0  # Expected 1.0 BTC balance
```

This mode is useful for handling missing trades or external position changes.

## Performance Optimization

### Interval Selection
- **1m**: Very responsive, higher transaction costs
- **5m**: Good balance of responsiveness and efficiency
- **15m**: Lower frequency, suitable for less volatile markets
- **1h**: Long-term positioning, minimal transaction costs

### Window Size Tuning
- **Small Window (7-10)**: More responsive to recent volatility changes
- **Medium Window (14-21)**: Standard ATR calculation, balanced approach
- **Large Window (30+)**: Smoother ATR, less sensitive to short-term spikes

### Multiplier Optimization
- **Market Conditions**: Adjust based on current volatility regime
- **Spread Competition**: Consider exchange's typical spreads
- **Risk Tolerance**: Higher multipliers for more conservative approach

## Common Use Cases

### 1. High-Frequency Market Making
```yaml
interval: 1m
window: 7
multiplier: 5.0
minPriceRange: 0.1%
amount: 50
```

### 2. Medium-Frequency Balanced
```yaml
interval: 5m
window: 14
multiplier: 10.0
minPriceRange: 0.5%
amount: 100
```

### 3. Conservative Long-Term
```yaml
interval: 1h
window: 21
multiplier: 20.0
minPriceRange: 1.0%
amount: 200
```

## Best Practices

1. **ATR Period Selection**: Use 14 periods as a starting point, adjust based on market characteristics
2. **Multiplier Tuning**: Start with 10x, increase for more conservative approach
3. **Minimum Range Protection**: Set to 0.5-1% to handle low volatility periods
4. **Position Sizing**: Use amount-based sizing for consistent risk exposure
5. **Market Selection**: Works best in liquid markets with consistent volatility
6. **Monitoring**: Regularly review fill rates and profitability

## Limitations

1. **Trending Markets**: May accumulate positions in strong trends
2. **Low Volatility**: Reduced profit opportunities during quiet periods
3. **High Volatility**: Wide spreads may reduce fill rates
4. **Market Gaps**: Cannot protect against overnight gaps or sudden moves
5. **Exchange Fees**: High-frequency trading may incur significant fees
6. **Slippage**: Take profit orders use limit orders which may not fill immediately

## Troubleshooting

### Common Issues

**No Orders Placed**
- Check minimum price range settings
- Verify sufficient account balance
- Ensure ATR calculation has enough data

**Orders Not Filling**
- Reduce ATR multiplier for tighter spreads
- Check market liquidity and typical spreads
- Verify price range calculations

**Excessive Position Accumulation**
- Enable take profit by expected balance
- Reduce position sizing
- Implement additional exit strategies

**ATR Calculation Errors**
- Ensure sufficient historical data
- Check interval and window settings
- Verify market data connectivity

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/atrpin.yaml)
- [ATR Indicator Guide](../../doc/topics/atr-indicator.md)
- [Market Making Strategies](../../doc/topics/market-making.md)
