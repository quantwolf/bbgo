# Grid2 Strategy

## Overview

Grid2 is an advanced grid trading strategy that places buy and sell orders at predetermined price levels to profit from market volatility. It creates a "grid" of orders across a price range, automatically buying low and selling high as the market oscillates. This strategy is particularly effective in sideways or ranging markets where prices fluctuate within a defined range.

## How It Works

1. **Grid Setup**: Creates a grid of buy and sell orders between upper and lower price boundaries
2. **Order Placement**: Places buy orders below current price and sell orders above current price
3. **Profit Capture**: When orders are filled, the strategy places new orders to maintain the grid
4. **Compound Growth**: Optionally reinvests profits to increase position sizes over time
5. **Risk Management**: Includes stop-loss, take-profit, and position management features
6. **Recovery System**: Can recover and rebuild grids after restarts or interruptions

## Key Features

- **Flexible Grid Configuration**: Supports both arithmetic and geometric grid spacing
- **Auto Range Detection**: Automatically determines price ranges from historical data
- **Compound Trading**: Reinvests profits to increase grid order sizes
- **Base Currency Earning**: Option to earn profits in base currency instead of quote
- **Recovery System**: Recovers grid state after restarts or connection issues
- **Profit Tracking**: Comprehensive profit statistics and performance metrics
- **Risk Controls**: Stop-loss, take-profit, and position size management
- **Order Management**: Advanced order lifecycle management with error handling

## Strategy Architecture

### Grid Structure

<augment_code_snippet path="pkg/strategy/grid2/grid.go" mode="EXCERPT">
```go
type Grid struct {
    UpperPrice fixedpoint.Value `json:"upperPrice"`
    LowerPrice fixedpoint.Value `json:"lowerPrice"`
    
    // Size is the number of total grids
    Size fixedpoint.Value `json:"size"`
    
    // TickSize is the price tick size, this is used for truncating price
    TickSize fixedpoint.Value `json:"tickSize"`
    
    // Spread is a immutable number
    Spread fixedpoint.Value `json:"spread"`
    
    // Pins are the pinned grid prices, from low to high
    Pins []Pin `json:"pins"`
}
```
</augment_code_snippet>

### Strategy Configuration

<augment_code_snippet path="pkg/strategy/grid2/strategy.go" mode="EXCERPT">
```go
type Strategy struct {
    Symbol string `json:"symbol"`
    
    // ProfitSpread is the fixed profit spread you want to submit the sell order
    ProfitSpread fixedpoint.Value `json:"profitSpread"`
    
    // GridNum is the grid number, how many orders you want to post on the orderbook
    GridNum int64 `json:"gridNumber"`
    
    UpperPrice fixedpoint.Value `json:"upperPrice"`
    LowerPrice fixedpoint.Value `json:"lowerPrice"`
    
    // Compound option is used for buying more inventory when
    // the profit is made by the filled sell order
    Compound bool `json:"compound"`
    
    // EarnBase option is used for earning profit in base currency
    EarnBase bool `json:"earnBase"`
    
    QuoteInvestment fixedpoint.Value `json:"quoteInvestment"`
    BaseInvestment  fixedpoint.Value `json:"baseInvestment"`
}
```
</augment_code_snippet>

### Grid Calculation

<augment_code_snippet path="pkg/strategy/grid2/grid.go" mode="EXCERPT">
```go
func calculateArithmeticPins(lower, upper, spread, tickSize fixedpoint.Value) []Pin {
    var pins []Pin
    
    // Calculate precision based on tick size
    var ts = tickSize.Float64()
    var prec = int(math.Round(math.Log10(ts) * -1.0))
    
    // Generate pins from lower to upper with spread intervals
    for p := lower; p.Compare(upper.Sub(spread)) <= 0; p = p.Add(spread) {
        price := util.RoundAndTruncatePrice(p, prec)
        pins = append(pins, Pin(price))
    }
    
    // Ensure upper price is included
    upperPrice := util.RoundAndTruncatePrice(upper, prec)
    pins = append(pins, Pin(upperPrice))
    
    return pins
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `gridNumber` | int | Yes | Number of grid levels between upper and lower prices |
| `upperPrice` | number | Yes* | Upper boundary of the grid |
| `lowerPrice` | number | Yes* | Lower boundary of the grid |

*Required unless using `autoRange`

### Investment Settings (Choose One)
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `amount` | number | No | Fixed quote amount per order (e.g., 10 USDT) |
| `quantity` | number | No | Fixed base quantity per order (e.g., 0.001 BTC) |
| `quoteInvestment` | number | No | Total quote currency to invest |
| `baseInvestment` | number | No | Total base currency to invest |

### Advanced Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `autoRange` | duration | No | Auto-detect price range from historical data |
| `profitSpread` | number | No | Custom profit spread (overrides calculated spread) |
| `compound` | boolean | No | Reinvest profits to increase order sizes |
| `earnBase` | boolean | No | Earn profits in base currency instead of quote |
| `triggerPrice` | number | No | Price level to trigger grid activation |
| `stopLossPrice` | number | No | Price level to close grid and liquidate |
| `takeProfitPrice` | number | No | Price level to close grid and take profits |

### Recovery and Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `recoverOrdersWhenStart` | boolean | No | Recover grid orders on restart |
| `keepOrdersWhenShutdown` | boolean | No | Keep orders active when shutting down |
| `clearOpenOrdersWhenStart` | boolean | No | Clear existing orders on start |
| `closeWhenCancelOrder` | boolean | No | Close grid if any order is manually canceled |

## Configuration Examples

### 1. Basic Grid Trading
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 50
    upperPrice: 50000
    lowerPrice: 30000
    quoteInvestment: 10000
    compound: false
```

### 2. Compound Grid with Auto Range
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: ETHUSDT
    gridNumber: 100
    autoRange: 14d              # Use 14-day price range
    quoteInvestment: 5000
    baseInvestment: 2.0
    compound: true              # Reinvest profits
    earnBase: true              # Earn ETH instead of USDT
```

### 3. Conservative Grid with Risk Management
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 30
    upperPrice: 45000
    lowerPrice: 35000
    amount: 100                 # Fixed $100 per order
    stopLossPrice: 32000        # Stop loss at $32,000
    takeProfitPrice: 48000      # Take profit at $48,000
    closeWhenCancelOrder: true  # Close if manually canceled
```

### 4. High-Frequency Grid
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 200
    upperPrice: 42000
    lowerPrice: 38000
    quantity: 0.001             # Fixed 0.001 BTC per order
    compound: true
    profitSpread: 50            # Custom $50 profit spread
```

## Grid Types

### Arithmetic Grid
- **Equal Price Intervals**: Each grid level is separated by the same price difference
- **Calculation**: `spread = (upperPrice - lowerPrice) / (gridNumber - 1)`
- **Best For**: Most market conditions, easier to understand and configure

### Geometric Grid
- **Equal Percentage Intervals**: Each grid level is separated by the same percentage
- **Calculation**: Uses exponential spacing between price levels
- **Best For**: Highly volatile markets or assets with exponential growth patterns

## Investment Modes

### 1. Fixed Amount Mode
```yaml
amount: 100  # $100 USDT per order
```
- Each order uses the same quote currency amount
- Order quantity varies with price level
- Simple and predictable capital allocation

### 2. Fixed Quantity Mode
```yaml
quantity: 0.01  # 0.01 BTC per order
```
- Each order uses the same base currency quantity
- Order value varies with price level
- Useful for accumulating specific amounts

### 3. Investment Mode
```yaml
quoteInvestment: 10000  # $10,000 total investment
baseInvestment: 0.5     # 0.5 BTC existing position
```
- Strategy calculates optimal order sizes
- Distributes investment across all grid levels
- Most capital-efficient approach

## Auto Range Detection

### Usage
```yaml
autoRange: 14d  # Use 14-day price range
```

### How It Works
1. **Historical Analysis**: Analyzes price data for the specified period
2. **Pivot Detection**: Finds pivot highs and lows within the timeframe
3. **Range Calculation**: Sets upperPrice to pivot high, lowerPrice to pivot low
4. **Dynamic Adjustment**: Automatically adapts to market conditions

### Valid Formats
- `7d` - 7 days
- `2w` - 2 weeks  
- `1h` - 1 hour
- `30m` - 30 minutes

## Profit Mechanisms

### Standard Mode (Quote Currency Profits)
- **Buy Low**: Places buy orders below current price
- **Sell High**: Places sell orders above current price
- **Profit**: Earns quote currency (e.g., USDT) from spread capture

### Earn Base Mode
```yaml
earnBase: true
```
- **Accumulation**: Focuses on accumulating base currency (e.g., BTC)
- **Strategy**: Uses profits to buy more base currency
- **Long-term**: Better for long-term asset accumulation

### Compound Mode
```yaml
compound: true
```
- **Reinvestment**: Automatically reinvests profits into larger orders
- **Growth**: Order sizes increase over time as profits accumulate
- **Exponential**: Can lead to exponential growth in favorable conditions

## Risk Management

### Stop Loss
```yaml
stopLossPrice: 30000
```
- Automatically closes grid and liquidates position if price drops below level
- Protects against major market downturns
- Converts all base currency to quote currency

### Take Profit
```yaml
takeProfitPrice: 60000
```
- Automatically closes grid and realizes profits if price rises above level
- Locks in gains during strong uptrends
- Maintains current position allocation

### Position Limits
- **Balance Checks**: Ensures sufficient balance before placing orders
- **Minimum Notional**: Respects exchange minimum order requirements
- **Risk Allocation**: Limits maximum exposure per grid level

## Recovery System

### Order Recovery
```yaml
recoverOrdersWhenStart: true
```
- Scans existing orders on startup
- Rebuilds grid state from active orders
- Continues trading seamlessly after restarts

### Trade Recovery
```yaml
recoverGridByScanningTrades: true
recoverGridWithin: 72h
```
- Analyzes recent trade history
- Reconstructs grid state from filled orders
- Recovers profit statistics and position data

### Error Handling
- **Connection Issues**: Automatically reconnects and recovers state
- **Order Failures**: Retries failed orders with exponential backoff
- **Balance Mismatches**: Adjusts grid to match available balances

## Performance Optimization

### Order Management
- **Batch Operations**: Groups order submissions for efficiency
- **Rate Limiting**: Respects exchange API rate limits
- **Error Recovery**: Handles temporary failures gracefully

### Memory Efficiency
- **Order Caching**: Caches active orders to reduce API calls
- **State Persistence**: Saves critical state to survive restarts
- **Garbage Collection**: Cleans up old data periodically

## Best Practices

1. **Range Selection**: Choose ranges based on historical support/resistance levels
2. **Grid Density**: More grids = more trades but higher fees
3. **Capital Allocation**: Don't invest more than you can afford to lose
4. **Market Conditions**: Works best in ranging/sideways markets
5. **Fee Consideration**: Ensure grid spread > 2 × trading fees
6. **Monitoring**: Regularly check grid performance and adjust if needed

## Common Use Cases

### Sideways Market Trading
- **Scenario**: Price oscillates between support and resistance
- **Strategy**: Wide grid with moderate density
- **Profit**: Consistent profits from range-bound trading

### Volatility Harvesting
- **Scenario**: High volatility with no clear trend
- **Strategy**: Dense grid with compound enabled
- **Profit**: Captures profits from price swings

### Dollar-Cost Averaging
- **Scenario**: Long-term accumulation of assets
- **Strategy**: Grid with earnBase enabled
- **Profit**: Accumulates base currency over time

### Arbitrage Enhancement
- **Scenario**: Enhance returns during low-volatility periods
- **Strategy**: Tight grid with small spreads
- **Profit**: Additional returns from micro-movements

## Limitations

1. **Trending Markets**: Can accumulate losing positions in strong trends
2. **Gap Risk**: Cannot protect against price gaps or sudden moves
3. **Fee Sensitivity**: High trading frequency increases fee costs
4. **Capital Requirements**: Requires significant capital for effective grids
5. **Market Risk**: Subject to overall market direction and volatility

## Troubleshooting

### Common Issues

**Grid Not Starting**
- Check price range validity (upper > lower)
- Verify sufficient balance for minimum orders
- Ensure grid spread > minimum profit threshold

**Orders Not Filling**
- Check if grid prices are competitive
- Verify market liquidity at grid levels
- Consider adjusting grid density or range

**Profit Lower Than Expected**
- Review trading fees vs. grid spread
- Check market volatility and trading frequency
- Consider enabling compound mode

**Recovery Issues**
- Enable order recovery options
- Check trade history permissions
- Verify API key permissions

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/grid2.yaml)
- [Grid Trading Guide](../../doc/topics/grid-trading.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
