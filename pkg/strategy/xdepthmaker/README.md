# Cross-Exchange Depth Maker (XDepthMaker) Strategy

## Overview

The XDepthMaker strategy is an advanced cross-exchange market-making strategy that provides deep liquidity by placing multiple layers of orders based on the order book depth from a source exchange. Unlike traditional market makers that use fixed spreads, XDepthMaker dynamically calculates order prices and quantities based on the accumulated depth requirements and the source exchange's order book structure.

## How It Works

1. **Depth Analysis**: Monitors order book depth from the source (hedge) exchange
2. **Dynamic Pricing**: Calculates order prices based on accumulated depth requirements
3. **Layer-Based Ordering**: Places multiple order layers with depth-scaled quantities
4. **Position Hedging**: Automatically hedges filled orders on the source exchange
5. **Adaptive Depth**: Falls back to RESTful API when WebSocket depth is insufficient
6. **Risk Management**: Implements comprehensive position and balance controls

## Key Features

- **Depth-Based Market Making**: Orders are placed based on required depth rather than fixed spreads
- **Multi-Layer Architecture**: Configurable number of order layers with depth scaling
- **Intelligent Depth Calculation**: Uses accumulated depth to determine optimal pricing
- **Adaptive Data Sources**: Switches between WebSocket and REST API for depth data
- **Multiple Hedge Strategies**: Market, BBO counter-party, and BBO queue hedging options
- **Fast Layer Updates**: Separate fast and full replenishment intervals
- **Comprehensive Monitoring**: Built-in Prometheus metrics for depth and spread analysis
- **Trade Recovery**: Automatic trade recovery mechanism for missed trades

## Strategy Architecture

### Core Components

1. **Depth Analyzer**: Analyzes source exchange order book depth
2. **Order Generator**: Creates maker orders based on depth requirements
3. **Hedge Engine**: Manages position hedging with multiple strategies
4. **Quote Worker**: Handles fast layer updates and full replenishment
5. **Metrics Collector**: Monitors spread ratios and depth metrics

### Hedge Strategies

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
type HedgeStrategy string

const (
	HedgeStrategyMarket           HedgeStrategy = "market"
	HedgeStrategyBboCounterParty1 HedgeStrategy = "bbo-counter-party-1"
	HedgeStrategyBboCounterParty3 HedgeStrategy = "bbo-counter-party-3"
	HedgeStrategyBboCounterParty5 HedgeStrategy = "bbo-counter-party-5"
	HedgeStrategyBboQueue1        HedgeStrategy = "bbo-queue-1"
)
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol for maker exchange |
| `hedgeSymbol` | string | No | Trading pair symbol for hedge exchange (defaults to symbol) |
| `makerExchange` | string | Yes | Maker exchange session name |
| `hedgeExchange` | string | Yes | Hedge exchange session name |

### Update Intervals
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `fastLayerUpdateInterval` | duration | No | Fast layer update interval (default: 5s) |
| `fullReplenishInterval` | duration | No | Full replenishment interval (default: 10m) |
| `hedgeInterval` | duration | No | Hedge execution interval (default: 3s) |

### Margin Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `margin` | percentage | No | Default margin for both sides (default: 0.3%) |
| `bidMargin` | percentage | No | Specific bid margin (overrides margin) |
| `askMargin` | percentage | No | Specific ask margin (overrides margin) |

### Layer Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `numLayers` | int | No | Number of order layers per side (default: 1) |
| `numOfFastLayers` | int | No | Number of fast update layers (default: 5) |
| `pips` | number | Yes | Price increment multiplier between layers |

### Depth Scaling
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `depthScale` | object | Yes | Depth scaling configuration |
| `depthScale.byLayer` | object | No | Layer-based depth scaling |
| `depthScale.byLayer.linear` | object | No | Linear scaling configuration |
| `depthScale.byLayer.exp` | object | No | Exponential scaling configuration |

### Quantity Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quantity` | number | No | Fixed quantity for first layer |
| `quantityScale` | object | No | Layer-based quantity scaling |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `maxExposurePosition` | number | No | Maximum unhedged position |
| `hedgeMaxOrderQuantity` | number | No | Maximum hedge order quantity |
| `stopHedgeQuoteBalance` | number | No | Stop hedging below this quote balance |
| `stopHedgeBaseBalance` | number | No | Stop hedging below this base balance |
| `disableHedge` | boolean | No | Disable all hedging |

### Hedge Strategy
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `hedgeStrategy` | string | No | Hedge strategy type (default: "market") |

### Advanced Features
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `recoverTrade` | boolean | No | Enable trade recovery |
| `recoverTradeScanPeriod` | duration | No | Trade recovery scan period |
| `notifyTrade` | boolean | No | Enable trade notifications |
| `skipCleanUpOpenOrders` | boolean | No | Skip cleaning up open orders on start |
| `priceImpactRatio` | percentage | No | Price impact ratio for BBO monitoring |

## Configuration Example

```yaml
crossExchangeStrategies:
  - xdepthmaker:
      symbol: BTCUSDT
      makerExchange: max          # Make market on MAX
      hedgeExchange: binance      # Hedge on Binance
      
      # Update intervals
      fastLayerUpdateInterval: 5s
      fullReplenishInterval: 10m
      hedgeInterval: 3s
      
      # Margin configuration
      margin: 0.4%                # 0.4% default margin
      bidMargin: 0.35%            # Slightly tighter bid margin
      askMargin: 0.45%            # Slightly wider ask margin
      
      # Layer configuration
      numLayers: 30               # 30 layers on each side
      numOfFastLayers: 5          # Update first 5 layers quickly
      pips: 10                    # 10x tick size between layers
      
      # Depth scaling configuration
      depthScale:
        byLayer:
          linear:
            domain: [1, 30]       # Layers 1 to 30
            range: [50, 20000]    # Depth from $50 to $20,000
      
      # Risk management
      maxExposurePosition: 0.1    # Max 0.1 BTC unhedged
      hedgeMaxOrderQuantity: 0.5  # Max 0.5 BTC per hedge order
      stopHedgeQuoteBalance: 1000 # Stop hedging below $1,000
      stopHedgeBaseBalance: 0.01  # Stop hedging below 0.01 BTC
      
      # Hedge strategy
      hedgeStrategy: "bbo-counter-party-1"  # Use BBO counter-party level 1
      
      # Advanced features
      recoverTrade: true
      recoverTradeScanPeriod: 30m
      notifyTrade: true
      priceImpactRatio: 0.1%
      
      # Profit fixer (optional)
      profitFixer:
        tradesSince: "2024-01-01T00:00:00Z"
```

## Strategy Logic

### Depth-Based Order Generation

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) generateMakerOrders(
	sourceBook *types.StreamOrderBook,
	maxLayer int,
	availableBase, availableQuote fixedpoint.Value,
) ([]types.SubmitOrder, error) {
	// Calculate required depth for each layer
	requiredDepthFloat, err := s.DepthScale.Scale(i)
	if err != nil {
		return nil, errors.Wrapf(err, "depthScale scale error")
	}
	
	// Accumulate depth requirements
	requiredDepth := fixedpoint.NewFromFloat(requiredDepthFloat)
	accumulatedDepth = accumulatedDepth.Add(requiredDepth)
	
	// Find price level that satisfies depth requirement
	index := sideBook.IndexByQuoteVolumeDepth(accumulatedDepth)
	depthPrice := pvs.AverageDepthPriceByQuote(accumulatedDepth, 0)
```
</augment_code_snippet>

### Adaptive Depth Data Source

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
// Check if WebSocket depth is sufficient
if requireFullDepthRequest {
	s.logger.Warnf("source book depth (%f) from websocket is not enough (< %f), falling back to RESTful api query...",
		actualDepth.Float64(), requiredDepth.Float64())
	
	if depthService, ok := s.hedgeSession.Exchange.(DepthQueryService); ok {
		snapshot, _, err := depthService.QueryDepth(context.Background(), s.HedgeSymbol, 0)
		if err != nil {
			s.logger.WithError(err).Errorf("unable to query source book depth via RESTful API")
		} else {
			dupPricingBook.Load(snapshot)
			s.logger.Infof("source depth snapshot is loaded from RESTful API")
		}
	}
}
```
</augment_code_snippet>

### Hedge Execution Strategies

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) Hedge(ctx context.Context, pos fixedpoint.Value) error {
	switch s.HedgeStrategy {
	case HedgeStrategyMarket:
		return s.executeHedgeMarket(ctx, side, quantity)
	case HedgeStrategyBboCounterParty1:
		return s.executeHedgeBboCounterPartyWithIndex(ctx, side, 1, quantity)
	case HedgeStrategyBboCounterParty3:
		return s.executeHedgeBboCounterPartyWithIndex(ctx, side, 3, quantity)
	case HedgeStrategyBboCounterParty5:
		return s.executeHedgeBboCounterPartyWithIndex(ctx, side, 5, quantity)
	case HedgeStrategyBboQueue1:
		return s.executeHedgeBboQueue1(ctx, side, quantity)
	}
}
```
</augment_code_snippet>

## Depth Scaling Configuration

### Linear Scaling
```yaml
depthScale:
  byLayer:
    linear:
      domain: [1, 20]     # Layer numbers 1 to 20
      range: [100, 5000]  # Depth from $100 to $5,000
```

### Exponential Scaling
```yaml
depthScale:
  byLayer:
    exp:
      domain: [1, 20]     # Layer numbers 1 to 20
      range: [1, 3]       # Exponential scale from 1x to 3x
```

## Hedge Strategy Options

### 1. Market Hedge (`market`)
- **Description**: Uses market orders for immediate execution
- **Pros**: Guaranteed execution, minimal slippage risk
- **Cons**: Higher trading costs, potential market impact

### 2. BBO Counter-Party (`bbo-counter-party-N`)
- **Description**: Places limit orders at Nth level of opposite side
- **Levels**: 1, 3, 5 (deeper levels for better prices)
- **Pros**: Better execution prices, reduced market impact
- **Cons**: Execution risk, potential partial fills

### 3. BBO Queue (`bbo-queue-1`)
- **Description**: Places orders at best bid/ask to join the queue
- **Pros**: Best possible prices, maker rebates
- **Cons**: Queue position risk, slower execution

## Risk Management Features

### 1. Position Exposure Control
- Monitors unhedged position against `maxExposurePosition`
- Automatically stops placing orders when exposure limits exceeded
- Separate tracking for base and quote currency exposures

### 2. Balance Protection
- Maintains minimum balances on both exchanges
- Prevents over-trading beyond available funds
- Configurable balance thresholds for both base and quote currencies

### 3. Hedge Order Limits
- Limits individual hedge order size via `hedgeMaxOrderQuantity`
- Prevents excessive single-order market impact
- Helps manage execution risk

### 4. Trade Recovery
- Automatically scans for missing trades via REST API
- Configurable scan periods and overlap buffers
- Ensures accurate position and PnL tracking

## Monitoring and Metrics

The strategy provides comprehensive Prometheus metrics:

- **Spread Metrics**: Market spread ratios and trends
- **Depth Metrics**: Order book depth in USD by price ranges
- **Price Level Metrics**: Number of price levels within ranges
- **Position Metrics**: Exposure levels and hedge ratios
- **Performance Metrics**: Order placement success rates

### Key Metrics
- `bbgo_xdepthmaker_market_spread_ratio`: Current market spread ratio
- `bbgo_xdepthmaker_depth_in_usd`: Market depth in USD by side and price range
- `bbgo_xdepthmaker_price_level_count`: Number of price levels within range

## Common Use Cases

### 1. Deep Liquidity Provision
```yaml
numLayers: 50
depthScale:
  byLayer:
    linear:
      domain: [1, 50]
      range: [100, 50000]
hedgeStrategy: "market"
```

### 2. Conservative Market Making
```yaml
numLayers: 10
margin: 0.5%
maxExposurePosition: 0.05
hedgeStrategy: "bbo-counter-party-3"
```

### 3. High-Frequency Updates
```yaml
fastLayerUpdateInterval: 1s
numOfFastLayers: 10
fullReplenishInterval: 5m
hedgeStrategy: "bbo-queue-1"
```

## Best Practices

1. **Start Conservative**: Begin with fewer layers and wider margins
2. **Monitor Depth Requirements**: Ensure source exchange has sufficient depth
3. **Balance Management**: Maintain adequate balances on both exchanges
4. **Hedge Strategy Selection**: Choose appropriate hedge strategy for market conditions
5. **Regular Monitoring**: Monitor metrics and adjust parameters as needed
6. **Trade Recovery**: Enable trade recovery for accurate tracking

## Troubleshooting

### Common Issues

**Insufficient Depth**
- Increase `depthScale` range values
- Reduce number of layers
- Check source exchange liquidity

**Hedge Failures**
- Verify hedge exchange connectivity
- Check available balances
- Review hedge strategy selection

**Order Placement Issues**
- Check margin settings
- Verify minimum order sizes
- Review balance availability

**Performance Issues**
- Optimize update intervals
- Reduce number of layers
- Monitor network latency

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Cross-Exchange Trading Guide](../../doc/topics/cross-exchange.md)
- [Configuration Examples](../../../config/xdepthmaker.yaml)
- [XMaker Strategy](../xmaker/README.md) - Alternative cross-exchange approach
