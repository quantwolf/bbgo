# Liquidity Maker Strategy

## Overview

The Liquidity Maker strategy is a sophisticated market-making strategy designed to provide liquidity to the order book by placing multiple layers of buy and sell orders around the current market price. This strategy aims to profit from the bid-ask spread while maintaining a balanced position and managing risk through various protective mechanisms.

## How It Works

1. **Liquidity Provision**: Places multiple layers of limit orders on both buy and sell sides
2. **Dynamic Pricing**: Uses mid-price or last trade price as reference for order placement
3. **Position Management**: Automatically adjusts orders based on current position and exposure
4. **Risk Control**: Implements various stop mechanisms and profit protection features
5. **Continuous Rebalancing**: Updates orders at configurable intervals

## Key Features

- **Multi-Layer Order Placement**: Creates liquidity depth with configurable number of layers
- **Scalable Order Sizing**: Uses exponential or linear scaling for order quantities
- **Position-Aware Trading**: Adjusts strategy based on current position exposure
- **Profit Protection**: Ensures minimum profit margins on position-closing orders
- **EMA-Based Controls**: Optional EMA filters for price bias protection
- **Comprehensive Metrics**: Built-in Prometheus metrics for monitoring
- **Flexible Configuration**: Highly customizable parameters for different market conditions

## Strategy Components

### 1. Liquidity Orders
- **Purpose**: Provide market liquidity and capture spread
- **Placement**: Multiple layers around mid-price with configurable spread
- **Scaling**: Exponential or linear quantity scaling across layers
- **Update Frequency**: Controlled by `liquidityUpdateInterval`

### 2. Adjustment Orders
- **Purpose**: Close existing positions with profit protection
- **Trigger**: Activated when position is not dust (significant size)
- **Protection**: Uses `profitProtectedPrice` function to ensure minimum profit
- **Update Frequency**: Controlled by `adjustmentUpdateInterval`

## Configuration Parameters

### Core Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol (e.g., "BTCUSDT") |
| `liquidityUpdateInterval` | interval | No | Interval to update liquidity orders (default: "1h") |
| `adjustmentUpdateInterval` | interval | No | Interval to update adjustment orders (default: "5m") |

### Liquidity Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `numOfLiquidityLayers` | int | Yes | Number of order layers on each side |
| `askLiquidityAmount` | number | Yes | Total liquidity amount for ask orders |
| `bidLiquidityAmount` | number | Yes | Total liquidity amount for bid orders |
| `liquidityPriceRange` | percentage | Yes | Price range for liquidity orders |
| `askLiquidityPriceRange` | percentage | No | Specific ask price range (overrides liquidityPriceRange) |
| `bidLiquidityPriceRange` | percentage | No | Specific bid price range (overrides liquidityPriceRange) |

### Scaling Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `liquidityScale` | object | Yes | Scaling function for order quantities |
| `liquidityScale.exp` | object | No | Exponential scaling configuration |
| `liquidityScale.linear` | object | No | Linear scaling configuration |

### Price and Spread Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `spread` | percentage | Yes | Minimum spread from mid-price |
| `useLastTradePrice` | boolean | No | Use last trade price instead of mid-price |
| `maxPrice` | number | No | Maximum allowed price for orders |
| `minPrice` | number | No | Minimum allowed price for orders |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `maxPositionExposure` | number | No | Maximum position size allowed |
| `minProfit` | percentage | No | Minimum profit margin for adjustment orders |
| `stopBidPrice` | number | No | Price level to stop placing bid orders |
| `stopAskPrice` | number | No | Price level to stop placing ask orders |
| `useProtectedPriceRange` | boolean | No | Enable profit protection for liquidity orders |

### Advanced Features
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `adjustmentOrderMaxQuantity` | number | No | Maximum quantity for adjustment orders |
| `adjustmentOrderPriceType` | string | No | Price type for adjustment orders (default: "MAKER") |

### EMA Controls
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `midPriceEMA` | object | No | Mid-price EMA bias protection |
| `midPriceEMA.enabled` | boolean | No | Enable mid-price EMA protection |
| `midPriceEMA.interval` | interval | No | EMA calculation interval |
| `midPriceEMA.window` | int | No | EMA window size |
| `midPriceEMA.maxBiasRatio` | percentage | No | Maximum allowed bias ratio |
| `stopEMA` | object | No | Stop EMA configuration |
| `stopEMA.enabled` | boolean | No | Enable stop EMA |
| `stopEMA.interval` | interval | No | Stop EMA interval |
| `stopEMA.window` | int | No | Stop EMA window |

## Configuration Example

```yaml
exchangeStrategies:
  - on: binance
    liquiditymaker:
      symbol: BTCUSDT
      
      # Update intervals
      liquidityUpdateInterval: 1h
      adjustmentUpdateInterval: 5m
      
      # Liquidity configuration
      numOfLiquidityLayers: 20
      askLiquidityAmount: 10000.0    # $10,000 worth of ask liquidity
      bidLiquidityAmount: 10000.0    # $10,000 worth of bid liquidity
      liquidityPriceRange: 2%        # 2% price range for orders
      
      # Spread and pricing
      spread: 0.1%                   # 0.1% minimum spread
      useLastTradePrice: true        # Use last trade price as reference
      
      # Order scaling (exponential)
      liquidityScale:
        exp:
          domain: [1, 20]            # Layer numbers 1 to 20
          range: [1, 4]              # Scale from 1x to 4x
      
      # Risk management
      maxPositionExposure: 1.0       # Maximum 1 BTC position
      minProfit: 0.05%               # Minimum 0.05% profit on adjustments
      useProtectedPriceRange: true   # Enable profit protection
      
      # Stop levels
      stopBidPrice: 25000            # Stop bids below $25,000
      stopAskPrice: 45000            # Stop asks above $45,000
      
      # EMA protection
      midPriceEMA:
        enabled: true
        interval: 5m
        window: 20
        maxBiasRatio: 1%             # Max 1% bias from EMA
```

## Strategy Logic

### Liquidity Order Placement

<augment_code_snippet path="pkg/strategy/liquiditymaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) placeLiquidityOrders(ctx context.Context) {
	// Cancel existing orders
	err := s.liquidityOrderBook.GracefulCancel(ctx, s.Session.Exchange)
	
	// Calculate mid-price and spread
	midPrice := ticker.Sell.Add(ticker.Buy).Div(fixedpoint.Two)
	sideSpread := s.Spread.Div(fixedpoint.Two)
	
	ask1Price := midPrice.Mul(fixedpoint.One.Add(sideSpread))
	bid1Price := midPrice.Mul(fixedpoint.One.Sub(sideSpread))
```
</augment_code_snippet>

### Order Generation

<augment_code_snippet path="pkg/strategy/liquiditymaker/generator.go" mode="EXCERPT">
```go
func (g *LiquidityOrderGenerator) Generate(
	side types.SideType, totalAmount, startPrice, endPrice fixedpoint.Value, numLayers int, scale bbgo.Scale,
) (orders []types.SubmitOrder) {
	// Calculate layer spread and prices
	layerSpread := endPrice.Sub(startPrice).Div(fixedpoint.NewFromInt(int64(numLayers - 1)))
	
	// Generate orders with scaling
	for i := 0; i < numLayers; i++ {
		layerPrice := startPrice.Add(layerSpread.Mul(fi))
		layerScale := scale.Call(float64(i + 1))
		quantity := fixedpoint.NewFromFloat(factor * layerScale)
```
</augment_code_snippet>

### Profit Protection

<augment_code_snippet path="pkg/strategy/liquiditymaker/strategy.go" mode="EXCERPT">
```go
func profitProtectedPrice(
	side types.SideType, averageCost, price, feeRate, minProfit fixedpoint.Value,
) fixedpoint.Value {
	switch side {
	case types.SideTypeSell:
		minProfitPrice := averageCost.Add(averageCost.Mul(feeRate.Add(minProfit)))
		return fixedpoint.Max(minProfitPrice, price)
	case types.SideTypeBuy:
		minProfitPrice := averageCost.Sub(averageCost.Mul(feeRate.Add(minProfit)))
		return fixedpoint.Min(minProfitPrice, price)
	}
```
</augment_code_snippet>

## Risk Management Features

### 1. Position Exposure Control
- Monitors total position size against `maxPositionExposure`
- Disables order placement when exposure limits are exceeded
- Separate controls for long and short positions

### 2. Stop Price Mechanisms
- `stopBidPrice`: Prevents bid orders above specified price
- `stopAskPrice`: Prevents ask orders below specified price
- `stopEMA`: Dynamic stop levels based on EMA indicator

### 3. Profit Protection
- Ensures adjustment orders maintain minimum profit margins
- Accounts for trading fees in profit calculations
- Optional protection for liquidity orders via `useProtectedPriceRange`

### 4. EMA Bias Protection
- Prevents order placement during extreme price movements
- Compares current price with EMA-smoothed price
- Configurable bias ratio threshold

## Monitoring and Metrics

The strategy provides comprehensive Prometheus metrics:

- **Price Metrics**: Mid-price, bid/ask prices, spread
- **Liquidity Metrics**: Order exposure, liquidity amounts
- **Position Metrics**: Current position, exposure levels
- **Performance Metrics**: Order placement status, bias ratios

### Key Metrics
- `liqmaker_spread`: Current market spread
- `liqmaker_mid_price`: Calculated mid-price
- `liqmaker_open_order_bid_exposure_in_usd`: Total bid exposure
- `liqmaker_open_order_ask_exposure_in_usd`: Total ask exposure
- `liqmaker_order_placement_status`: Order placement status by side

## Common Use Cases

### 1. High-Frequency Market Making
```yaml
liquidityUpdateInterval: 1m
adjustmentUpdateInterval: 30s
numOfLiquidityLayers: 50
spread: 0.05%
```

### 2. Conservative Liquidity Provision
```yaml
liquidityUpdateInterval: 1h
adjustmentUpdateInterval: 5m
numOfLiquidityLayers: 10
spread: 0.2%
maxPositionExposure: 0.5
```

### 3. Volatile Market Adaptation
```yaml
midPriceEMA:
  enabled: true
  maxBiasRatio: 2%
stopEMA:
  enabled: true
useProtectedPriceRange: true
```

## Best Practices

1. **Start Conservative**: Begin with wider spreads and smaller liquidity amounts
2. **Monitor Metrics**: Use Prometheus metrics to track performance
3. **Adjust for Volatility**: Increase spreads during high volatility periods
4. **Position Management**: Set appropriate `maxPositionExposure` limits
5. **Fee Optimization**: Use maker orders to minimize trading fees
6. **Backtesting**: Test configurations thoroughly before live deployment

## Troubleshooting

### Common Issues

**Orders Not Placed**
- Check balance availability
- Verify minimum order size requirements
- Review stop price configurations

**Excessive Position Exposure**
- Adjust `maxPositionExposure` setting
- Review adjustment order frequency
- Check profit protection settings

**Poor Performance**
- Analyze spread vs. market volatility
- Review order layer distribution
- Monitor fill rates and inventory turnover

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/liquiditymaker.yaml)
- [Market Making Best Practices](../../doc/topics/market-making.md)
