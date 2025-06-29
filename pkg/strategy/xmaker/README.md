# Cross-Exchange Market Maker (XMaker) Strategy

## Overview

The XMaker strategy is an advanced cross-exchange market-making strategy that provides liquidity on one exchange (maker exchange) while hedging positions on another exchange (source exchange). This strategy captures arbitrage opportunities and spreads between different exchanges while maintaining a neutral position through sophisticated hedging mechanisms.

## How It Works

1. **Price Discovery**: Monitors order book and price data from the source exchange
2. **Quote Generation**: Creates bid/ask quotes on the maker exchange with configurable margins
3. **Order Placement**: Places multiple layers of limit orders on the maker exchange
4. **Position Hedging**: Automatically hedges filled orders on the source exchange
5. **Signal Integration**: Uses multiple signal sources to adjust margins and trading behavior
6. **Risk Management**: Implements comprehensive risk controls and circuit breakers

## Key Features

- **Cross-Exchange Arbitrage**: Captures price differences between exchanges
- **Multi-Layer Market Making**: Places multiple order layers with configurable scaling
- **Advanced Hedging**: Direct hedging, delayed hedging, and synthetic hedging options
- **Signal-Based Margin Adjustment**: Dynamic margin adjustment based on market signals
- **Bollinger Band Integration**: Trend-based margin adjustments using Bollinger Bands
- **Spread Making**: Intelligent spread-making orders for position management
- **Circuit Breaker Protection**: Automatic trading halt on excessive losses
- **Comprehensive Monitoring**: Built-in metrics and performance tracking

## Strategy Architecture

### Core Components

1. **Quote Engine**: Generates bid/ask prices with margins and fees
2. **Order Manager**: Handles multi-layer order placement and cancellation
3. **Hedge Engine**: Manages position hedging across exchanges
4. **Signal Processor**: Aggregates multiple market signals
5. **Risk Controller**: Monitors and controls trading risks

### Signal Types

<augment_code_snippet path="pkg/strategy/xmaker/signal.go" mode="EXCERPT">
```go
type SignalConfig struct {
	Weight                   float64                         `json:"weight"`
	BollingerBandTrendSignal *BollingerBandTrendSignal       `json:"bollingerBandTrend,omitempty"`
	OrderBookBestPriceSignal *OrderBookBestPriceVolumeSignal `json:"orderBookBestPrice,omitempty"`
	DepthRatioSignal         *DepthRatioSignal               `json:"depthRatio,omitempty"`
	KLineShapeSignal         *KLineShapeSignal               `json:"klineShape,omitempty"`
	TradeVolumeWindowSignal  *TradeVolumeWindowSignal        `json:"tradeVolumeWindow,omitempty"`
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | string | Yes | Trading pair symbol |
| `sourceExchange` | string | Yes | Source exchange session name for hedging |
| `makerExchange` | string | Yes | Maker exchange session name for market making |
| `updateInterval` | duration | No | Quote update interval (default: "1s") |
| `hedgeInterval` | duration | No | Hedge execution interval (default: "10s") |

### Margin Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `margin` | percentage | Yes | Default margin for both sides |
| `bidMargin` | percentage | No | Specific bid margin (overrides margin) |
| `askMargin` | percentage | No | Specific ask margin (overrides margin) |
| `minMargin` | percentage | No | Minimum margin protection |

### Order Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quantity` | number | Yes | Base quantity for first layer |
| `quantityMultiplier` | number | No | Multiplier for subsequent layers |
| `quantityScale` | object | No | Layer-based quantity scaling |
| `numLayers` | int | Yes | Number of order layers per side |
| `pips` | number | Yes | Price increment between layers |
| `makerOnly` | boolean | No | Use maker-only orders |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `maxExposurePosition` | number | No | Maximum unhedged position |
| `maxHedgeAccountLeverage` | number | No | Maximum leverage for hedge account |
| `maxHedgeQuoteQuantityPerOrder` | number | No | Maximum hedge order size |
| `minMarginLevel` | number | No | Minimum margin level for margin trading |
| `stopHedgeQuoteBalance` | number | No | Stop hedging below this quote balance |
| `stopHedgeBaseBalance` | number | No | Stop hedging below this base balance |
| `disableHedge` | boolean | No | Disable all hedging |

### Signal Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `enableSignalMargin` | boolean | No | Enable signal-based margin adjustment |
| `signals` | array | No | List of signal configurations |
| `signalReverseSideMargin` | object | No | Margin adjustment for reverse signals |
| `signalTrendSideMarginDiscount` | object | No | Margin discount for trend signals |

### Advanced Features
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `enableDelayHedge` | boolean | No | Enable delayed hedging |
| `maxHedgeDelayDuration` | duration | No | Maximum hedge delay |
| `delayHedgeSignalThreshold` | number | No | Signal threshold for delay |
| `enableBollBandMargin` | boolean | No | Enable Bollinger Band margin |
| `enableArbitrage` | boolean | No | Enable arbitrage mode |
| `useDepthPrice` | boolean | No | Use depth-weighted pricing |

## Configuration Example

```yaml
crossExchangeStrategies:
  - xmaker:
      symbol: BTCUSDT
      sourceExchange: binance    # Hedge on Binance
      makerExchange: max         # Make market on MAX
      
      # Update intervals
      updateInterval: 1s
      hedgeInterval: 10s
      
      # Margin configuration
      margin: 0.4%               # 0.4% default margin
      bidMargin: 0.35%           # Slightly tighter bid margin
      askMargin: 0.45%           # Slightly wider ask margin
      minMargin: 0.1%            # Minimum margin protection
      
      # Order configuration
      quantity: 0.001            # 0.001 BTC per layer
      quantityMultiplier: 1.5    # Increase quantity by 1.5x per layer
      numLayers: 3               # 3 layers on each side
      pips: 10                   # $10 between layers
      makerOnly: true            # Use maker-only orders
      
      # Risk management
      maxExposurePosition: 0.1   # Max 0.1 BTC unhedged
      maxHedgeAccountLeverage: 3 # Max 3x leverage
      minMarginLevel: 1.5        # Maintain 1.5 margin level
      
      # Signal-based margin adjustment
      enableSignalMargin: true
      signals:
        - weight: 1.0
          bollingerBandTrend:
            interval: 5m
            window: 20
            bandWidth: 2.0
        - weight: 0.5
          depthRatio:
            depthQuantity: 1.0
            averageDepthPeriod: 5
      
      # Signal margin configuration
      signalTrendSideMarginDiscount:
        enabled: true
        threshold: 0.5
        scale:
          linear:
            domain: [0.5, 2.0]
            range: [0, 0.2]
      
      # Delayed hedging
      enableDelayHedge: true
      maxHedgeDelayDuration: 30s
      delayHedgeSignalThreshold: 1.0
      
      # Circuit breaker
      circuitBreaker:
        enabled: true
        maximumConsecutiveTotalLoss: 50.0
        maximumConsecutiveLossTimes: 5
        haltDuration: 30m
```

## Strategy Logic

### Quote Generation Process

<augment_code_snippet path="pkg/strategy/xmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) updateQuote(ctx context.Context) error {
	// Cancel existing orders
	if err := s.activeMakerOrders.GracefulCancel(ctx, s.makerSession.Exchange); err != nil {
		return nil
	}
	
	// Get best bid/ask from source
	bestBid, bestAsk, hasPrice := s.sourceBook.BestBidAndAsk()
	if !hasPrice {
		return nil
	}
	
	// Apply margins and generate quotes
	var quote = &Quote{
		BestBidPrice: bestBidPrice,
		BestAskPrice: bestAskPrice,
		BidMargin:    s.BidMargin,
		AskMargin:    s.AskMargin,
		BidLayerPips: s.Pips,
		AskLayerPips: s.Pips,
	}
```
</augment_code_snippet>

### Hedging Mechanism

<augment_code_snippet path="pkg/strategy/xmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) hedge(ctx context.Context, uncoveredPosition fixedpoint.Value) {
	if uncoveredPosition.IsZero() {
		return
	}
	
	// Calculate hedge delta (opposite of uncovered position)
	hedgeDelta := uncoveredPosition.Neg()
	side := positionToSide(hedgeDelta)
	
	// Check if hedge can be delayed
	if s.canDelayHedge(side, hedgeDelta) {
		return
	}
	
	// Execute hedge order
	if _, err := s.directHedge(ctx, hedgeDelta); err != nil {
		log.WithError(err).Errorf("unable to hedge position")
	}
}
```
</augment_code_snippet>

## Signal Integration

### Bollinger Band Signal
- **Purpose**: Detect trend direction and adjust margins accordingly
- **Logic**: Increase margin on trend side, decrease on counter-trend side
- **Configuration**: Interval, window, and band width parameters

### Depth Ratio Signal
- **Purpose**: Analyze order book depth imbalance
- **Logic**: Adjust margins based on bid/ask depth ratio
- **Configuration**: Depth quantity and averaging period

### Trade Volume Signal
- **Purpose**: Monitor trading volume patterns
- **Logic**: Adjust behavior based on volume trends
- **Configuration**: Window size and volume thresholds

## Risk Management Features

### 1. Position Exposure Control
- Monitors unhedged position against `maxExposurePosition`
- Disables order placement when exposure limits exceeded
- Separate controls for different position directions

### 2. Margin Level Protection
- Monitors margin account health on source exchange
- Prevents hedging when margin level too low
- Calculates available debt quota for safe hedging

### 3. Circuit Breaker
- Automatic trading halt on consecutive losses
- Configurable loss thresholds and halt duration
- Prevents catastrophic losses during market stress

### 4. Balance Protection
- Maintains minimum balances on both exchanges
- Prevents over-trading beyond available funds
- Configurable balance thresholds

## Advanced Features

### Delayed Hedging
- Delays hedge execution when signals are strong
- Allows capturing additional profit from favorable moves
- Configurable delay duration and signal thresholds

### Spread Making
- Places intelligent spread-making orders
- Helps close positions at favorable prices
- Configurable profit targets and order lifespans

### Synthetic Hedging
- Alternative hedging using synthetic instruments
- Useful when direct hedging is not available
- Configurable synthetic pair relationships

### Arbitrage Mode
- Detects and executes arbitrage opportunities
- Places IOC orders when price differences exceed margins
- Automatic profit capture from price discrepancies

## Monitoring and Metrics

The strategy provides comprehensive metrics:

- **Price Metrics**: Source/maker prices, spreads, margins
- **Position Metrics**: Exposure, hedge ratios, PnL
- **Order Metrics**: Fill rates, order placement success
- **Signal Metrics**: Aggregated signals, margin adjustments
- **Performance Metrics**: Profit/loss, Sharpe ratio, drawdown

## Common Use Cases

### 1. Conservative Cross-Exchange Market Making
```yaml
margin: 0.5%
numLayers: 2
quantity: 0.001
maxExposurePosition: 0.05
enableDelayHedge: false
```

### 2. Aggressive Arbitrage Trading
```yaml
margin: 0.2%
numLayers: 5
enableArbitrage: true
enableDelayHedge: true
maxHedgeDelayDuration: 10s
```

### 3. Signal-Driven Adaptive Trading
```yaml
enableSignalMargin: true
enableBollBandMargin: true
signals: [multiple signal configs]
signalTrendSideMarginDiscount: [config]
```

## Best Practices

1. **Start Conservative**: Begin with wider margins and smaller quantities
2. **Monitor Latency**: Ensure low latency between exchanges
3. **Balance Management**: Maintain adequate balances on both exchanges
4. **Signal Tuning**: Carefully tune signal parameters for your market
5. **Risk Limits**: Set appropriate exposure and loss limits
6. **Regular Monitoring**: Monitor performance and adjust parameters

## Troubleshooting

### Common Issues

**Orders Not Filled**
- Check margin settings (too wide margins)
- Verify market liquidity and competition
- Review order placement timing

**Hedge Failures**
- Check source exchange connectivity
- Verify available balances for hedging
- Review margin level requirements

**Signal Issues**
- Validate signal data sources
- Check signal calculation parameters
- Monitor signal aggregation weights

**Performance Issues**
- Analyze spread capture vs. costs
- Review hedge timing and slippage
- Monitor inventory turnover rates

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Cross-Exchange Trading Guide](../../doc/topics/cross-exchange.md)
- [Configuration Examples](../../../config/xmaker.yaml)
- [Risk Management Best Practices](../../doc/topics/risk-management.md)
