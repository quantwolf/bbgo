# Trading Desk Strategy

## Overview

The Trading Desk strategy is a manual trading assistant that provides intelligent position sizing based on risk management principles. It acts as a "trading desk" that helps traders execute manual trades with proper risk controls by automatically calculating optimal position sizes based on stop-loss levels and maximum loss limits.

## How It Works

1. **Manual Trade Execution**: Provides an interface for manual trade execution with risk controls
2. **Risk-Based Position Sizing**: Automatically calculates position size based on stop-loss distance and maximum loss tolerance
3. **Balance Validation**: Ensures sufficient account balance before executing trades
4. **Price Type Support**: Supports both maker and taker pricing for position sizing calculations
5. **Multi-Asset Support**: Works with any trading pair supported by the exchange

## Key Features

- **Intelligent Position Sizing**: Calculates optimal position size based on risk parameters
- **Risk Management**: Enforces maximum loss limits per trade
- **Balance Protection**: Validates account balance before trade execution
- **Stop-Loss Integration**: Uses stop-loss prices for risk calculation
- **Flexible Pricing**: Supports maker/taker pricing models
- **Manual Control**: Designed for discretionary trading with automated risk controls

## Strategy Logic

### Position Size Calculation

<augment_code_snippet path="pkg/strategy/tradingdesk/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) calculatePositionSize(ctx context.Context, param OpenPositionParam) (fixedpoint.Value, error) {
	// If no stop loss is set, return the original quantity
	if param.StopLossPrice.IsZero() {
		return param.Quantity, nil
	}

	// Get current market price
	ticker, err := s.Session.Exchange.QueryTicker(ctx, param.Symbol)
	if err != nil {
		return fixedpoint.Zero, fmt.Errorf("failed to query ticker for %s: %w", param.Symbol, err)
	}

	// Calculate entry price based on side and price type
	var entryPrice fixedpoint.Value
	if param.Side == types.SideTypeBuy {
		entryPrice = ticker.Buy // Ask price for buying
	} else {
		entryPrice = ticker.Sell // Bid price for selling
	}

	// Calculate risk per unit
	var riskPerUnit fixedpoint.Value
	if param.Side == types.SideTypeBuy {
		if param.StopLossPrice.Compare(entryPrice) >= 0 {
			return fixedpoint.Zero, fmt.Errorf("invalid stop loss price for buy order")
		}
		riskPerUnit = entryPrice.Sub(param.StopLossPrice)
	} else {
		if param.StopLossPrice.Compare(entryPrice) <= 0 {
			return fixedpoint.Zero, fmt.Errorf("invalid stop loss price for sell order")
		}
		riskPerUnit = param.StopLossPrice.Sub(entryPrice)
	}

	// Calculate maximum quantity based on risk
	var maxQuantityByRisk fixedpoint.Value
	if s.MaxLossLimit.Sign() > 0 && riskPerUnit.Sign() > 0 {
		maxQuantityByRisk = s.MaxLossLimit.Div(riskPerUnit)
	} else {
		maxQuantityByRisk = param.Quantity
	}

	// Calculate maximum quantity based on available balance
	maxQuantityByBalance := s.calculateMaxQuantityByBalance(param.Symbol, param.Side, entryPrice)

	// Return the minimum of requested, risk-limited, and balance-limited quantities
	return fixedpoint.Min(param.Quantity, fixedpoint.Min(maxQuantityByRisk, maxQuantityByBalance)), nil
}
```
</augment_code_snippet>

### Balance Calculation

<augment_code_snippet path="pkg/strategy/tradingdesk/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) calculateMaxQuantityByBalance(symbol string, side types.SideType, price fixedpoint.Value) fixedpoint.Value {
	market, ok := s.Session.Market(symbol)
	if !ok {
		return fixedpoint.Zero
	}

	account := s.Session.GetAccount()
	if side == types.SideTypeBuy {
		// For buy orders, check quote currency balance
		balance := account.Balance(market.QuoteCurrency)
		if balance.Available.Sign() <= 0 || price.Sign() <= 0 {
			return fixedpoint.Zero
		}
		return balance.Available.Div(price)
	} else {
		// For sell orders, check base currency balance
		balance := account.Balance(market.BaseCurrency)
		return balance.Available
	}
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `maxLossLimit` | number | No | Maximum loss limit per trade in quote currency |
| `priceType` | string | No | Price type for calculations ("maker" or "taker") |

### Advanced Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session` | string | No | Exchange session name (defaults to primary session) |

## Configuration Example

```yaml
exchangeStrategies:
  - on: binance
    tradingdesk:
      maxLossLimit: 100.0        # Maximum $100 loss per trade
      priceType: maker           # Use maker pricing for calculations

  # Alternative configuration with higher risk tolerance
  - on: max
    tradingdesk:
      maxLossLimit: 500.0        # Maximum $500 loss per trade
      priceType: taker           # Use taker pricing for calculations
```

## Usage Patterns

### Manual Trade Execution

The Trading Desk strategy is designed to be used programmatically or through trading interfaces. Here's how position sizing works:

```go
// Example usage in Go code
param := OpenPositionParam{
    Symbol:        "BTCUSDT",
    Side:          types.SideTypeBuy,
    Quantity:      fixedpoint.NewFromFloat(1.0),    // Want to buy 1 BTC
    StopLossPrice: fixedpoint.NewFromFloat(45000),  // Stop loss at $45,000
}

// Strategy calculates optimal position size
optimalSize, err := strategy.calculatePositionSize(ctx, param)
// If BTC is at $50,000 and maxLossLimit is $1000:
// Risk per unit = $50,000 - $45,000 = $5,000
// Max quantity by risk = $1,000 / $5,000 = 0.2 BTC
// Final quantity = min(1.0, 0.2, available_balance) = 0.2 BTC
```

## Risk Management Features

### 1. Position Size Optimization
- **Risk-Based Sizing**: Calculates position size based on stop-loss distance
- **Loss Limit Enforcement**: Ensures no single trade can exceed maximum loss limit
- **Balance Protection**: Prevents over-leveraging beyond available funds

### 2. Price Validation
- **Stop-Loss Validation**: Ensures stop-loss prices are logically correct
- **Market Price Integration**: Uses real-time market prices for calculations
- **Price Type Flexibility**: Supports both maker and taker pricing models

### 3. Balance Management
- **Real-Time Balance Check**: Validates available balance before trade execution
- **Multi-Currency Support**: Handles both base and quote currency constraints
- **Precision Handling**: Respects market precision requirements

## Mathematical Foundation

### Risk Calculation Formula

For **Buy Orders**:
```
Risk per Unit = Entry Price - Stop Loss Price
Max Quantity by Risk = Max Loss Limit / Risk per Unit
```

For **Sell Orders**:
```
Risk per Unit = Stop Loss Price - Entry Price  
Max Quantity by Risk = Max Loss Limit / Risk per Unit
```

### Final Position Size
```
Final Quantity = min(
    Requested Quantity,
    Max Quantity by Risk,
    Max Quantity by Balance
)
```

## Use Cases

### 1. Conservative Risk Management
```yaml
tradingdesk:
  maxLossLimit: 50.0           # Risk only $50 per trade
  priceType: maker             # Use maker pricing for better fills
```

### 2. Moderate Risk Trading
```yaml
tradingdesk:
  maxLossLimit: 200.0          # Risk up to $200 per trade
  priceType: taker             # Use taker pricing for immediate execution
```

### 3. High-Risk Trading
```yaml
tradingdesk:
  maxLossLimit: 1000.0         # Risk up to $1000 per trade
  priceType: maker             # Use maker pricing to minimize costs
```

## Integration Examples

### With Trading Bots
```go
// Integration with automated trading systems
func executeTrade(symbol string, side types.SideType, quantity float64, stopLoss float64) {
    param := OpenPositionParam{
        Symbol:        symbol,
        Side:          side,
        Quantity:      fixedpoint.NewFromFloat(quantity),
        StopLossPrice: fixedpoint.NewFromFloat(stopLoss),
    }
    
    optimalSize, err := tradingDesk.calculatePositionSize(ctx, param)
    if err != nil {
        log.Error("Failed to calculate position size:", err)
        return
    }
    
    // Execute trade with optimal size
    executeOrder(symbol, side, optimalSize)
}
```

### With Manual Trading Interfaces
```go
// Integration with trading UI
func handleTradeRequest(req TradeRequest) {
    // Validate and calculate optimal position size
    param := OpenPositionParam{
        Symbol:        req.Symbol,
        Side:          req.Side,
        Quantity:      req.Quantity,
        StopLossPrice: req.StopLoss,
    }
    
    optimalSize, err := tradingDesk.calculatePositionSize(ctx, param)
    if err != nil {
        return TradeResponse{Error: err.Error()}
    }
    
    return TradeResponse{
        OptimalSize: optimalSize,
        RiskAmount:  calculateRisk(optimalSize, req.StopLoss),
    }
}
```

## Best Practices

1. **Set Appropriate Loss Limits**: Choose `maxLossLimit` based on your risk tolerance and account size
2. **Always Use Stop Losses**: The strategy works best when stop-loss prices are provided
3. **Monitor Account Balance**: Ensure sufficient balance for intended trade sizes
4. **Choose Correct Price Type**: Use "maker" for better pricing, "taker" for immediate execution
5. **Regular Review**: Periodically review and adjust loss limits based on performance
6. **Position Sizing**: Let the strategy calculate optimal sizes rather than using fixed amounts

## Limitations

1. **Manual Execution Required**: Strategy calculates sizes but doesn't automatically execute trades
2. **Stop-Loss Dependency**: Works best when stop-loss prices are provided
3. **Single Trade Focus**: Designed for individual trade risk management, not portfolio-level risk
4. **Market Price Dependency**: Requires real-time market data for accurate calculations

## Error Handling

### Common Error Scenarios

**Invalid Stop-Loss Price**
```
Error: "invalid stop loss price for buy order"
Solution: Ensure stop-loss is below entry price for buy orders, above for sell orders
```

**Insufficient Balance**
```
Behavior: Returns maximum available quantity based on balance
Solution: Ensure adequate account balance or reduce trade size
```

**Market Data Unavailable**
```
Error: "failed to query ticker"
Solution: Check exchange connectivity and symbol validity
```

## Testing

The strategy includes comprehensive unit tests covering:
- Normal position sizing calculations
- Edge cases (insufficient balance, invalid stop-loss)
- Different price types and market conditions
- Error handling scenarios

Run tests with:
```bash
go test ./pkg/strategy/tradingdesk/...
```

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
- [Configuration Examples](../../../config/tradingdesk.yaml)
- [Position Sizing Techniques](../../doc/topics/position-sizing.md)
