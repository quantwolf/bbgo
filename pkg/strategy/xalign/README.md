# XAlign Strategy

## Overview

The XAlign strategy is a sophisticated cross-exchange balance alignment strategy that automatically maintains target balance allocations across multiple exchanges. It monitors balance deviations from expected levels and executes trades to realign portfolios when discrepancies exceed tolerance thresholds for sustained periods. This strategy is particularly useful for maintaining consistent portfolio allocation across different exchanges or implementing cross-exchange arbitrage opportunities.

## How It Works

1. **Balance Monitoring**: Continuously monitors account balances across multiple exchanges
2. **Deviation Detection**: Uses statistical analysis to detect when balances deviate from expected targets
3. **Sustained Deviation Tracking**: Tracks how long deviations persist before triggering actions
4. **Intelligent Order Placement**: Places buy/sell orders to realign balances toward target allocations
5. **Quote Currency Selection**: Intelligently selects optimal quote currencies for different trade directions
6. **Alert System**: Sends notifications when large discrepancies are detected or corrected

## Key Features

- **Cross-Exchange Balance Management**: Monitors and aligns balances across multiple exchanges simultaneously
- **Deviation Detection System**: Advanced statistical analysis to identify significant balance deviations
- **Sustained Deviation Tracking**: Only acts on deviations that persist for configurable time periods
- **Intelligent Quote Currency Selection**: Different quote currencies for buy vs. sell operations
- **Large Amount Alerts**: Slack notifications for significant balance discrepancies
- **Flexible Tolerance Settings**: Configurable tolerance ranges and time thresholds
- **Dry Run Support**: Test mode for strategy validation without actual trading
- **Order Type Selection**: Support for both maker and taker orders

## Strategy Architecture

### Deviation Detection System

<augment_code_snippet path="pkg/strategy/xalign/detector/deviation.go" mode="EXCERPT">
```go
type DeviationDetector[T any] struct {
    mu            sync.Mutex
    expectedValue T             // Expected value for comparison
    tolerance     float64       // Tolerance percentage (e.g., 0.01 for 1%)
    duration      time.Duration // Time limit for sustained deviation

    toFloat64Amount func(T) (float64, error) // Function to convert T to float64
    records         []Record[T]              // Tracks deviation records
}

func (d *DeviationDetector[T]) AddRecord(at time.Time, value T) (bool, time.Duration) {
    // Calculate deviation percentage
    deviationPercentage := math.Abs((current - expected) / expected)
    
    // Reset records if deviation is within tolerance
    if deviationPercentage <= d.tolerance {
        d.records = nil
        return false, 0
    }
    
    // Track sustained deviations
    d.records = append(d.records, record)
    return d.ShouldFix()
}
```
</augment_code_snippet>

### Balance Alignment Logic

<augment_code_snippet path="pkg/strategy/xalign/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) align(ctx context.Context, sessions map[string]*bbgo.ExchangeSession) {
    for currency, expectedBalance := range s.ExpectedBalances {
        // Calculate total balance across all sessions
        totalBalance := s.calculateTotalBalance(currency, sessions)
        
        // Check for deviation using detector
        shouldFix, sustainedDuration := s.detectors[currency].AddRecord(time.Now(), totalBalance)
        
        if shouldFix {
            delta := totalBalance.Sub(expectedBalance)
            s.executeBalanceAlignment(ctx, currency, delta, sustainedDuration, sessions)
        }
    }
}
```
</augment_code_snippet>

### Alert System

<augment_code_snippet path="pkg/strategy/xalign/alert_cbd.go" mode="EXCERPT">
```go
type CriticalBalanceDiscrepancyAlert struct {
    SlackAlert *slackalert.SlackAlert
    
    Warning bool
    BaseCurrency      string
    Delta             fixedpoint.Value
    SustainedDuration time.Duration
    
    QuoteCurrency string
    AlertAmount   fixedpoint.Value
    
    Side     types.SideType
    Price    fixedpoint.Value
    Quantity fixedpoint.Value
    Amount   fixedpoint.Value
}
```
</augment_code_snippet>

## Configuration Parameters

### Basic Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sessions` | array | Yes | List of exchange sessions to monitor |
| `interval` | duration | Yes | Balance check interval (e.g., "1m", "30s") |
| `for` | duration | Yes | Sustained deviation duration threshold |
| `expectedBalances` | object | Yes | Target balance for each currency |

### Trading Settings
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `quoteCurrencies.buy` | array | Yes | Preferred quote currencies for buy orders |
| `quoteCurrencies.sell` | array | Yes | Preferred quote currencies for sell orders |
| `useTakerOrder` | boolean | No | Use taker orders instead of maker orders |
| `balanceToleranceRange` | percentage | Yes | Tolerance range for balance deviations |

### Risk Management
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `maxAmounts` | object | Yes | Maximum trade amounts per quote currency |
| `dryRun` | boolean | No | Enable dry run mode (no actual trades) |

### Alert Configuration
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `largeAmountAlert.quoteCurrency` | string | No | Quote currency for alert calculations |
| `largeAmountAlert.amount` | number | No | Threshold for large amount alerts |
| `largeAmountAlert.slack.channel` | string | No | Slack channel for alerts |
| `largeAmountAlert.slack.mentions` | array | No | Users/groups to mention in alerts |

## Configuration Example

```yaml
crossExchangeStrategies:
- xalign:
    # Monitoring settings
    interval: 1m                    # Check balances every minute
    for: 5m                         # Act on deviations sustained for 5+ minutes
    
    # Exchange sessions to monitor
    sessions:
    - binance
    - max
    - okex
    
    # Target balance allocations
    expectedBalances:
      BTC: 1.0                      # Maintain 1.0 BTC total across exchanges
      ETH: 10.0                     # Maintain 10.0 ETH total across exchanges
      USDT: 50000                   # Maintain $50,000 USDT total
    
    # Quote currency preferences
    quoteCurrencies:
      buy: [USDT, USDC, BUSD]       # Prefer USDT for buy orders
      sell: [USDT, USDC]            # Prefer USDT for sell orders
    
    # Risk management
    balanceToleranceRange: 2%       # 2% tolerance before triggering
    maxAmounts:
      USDT: 1000                    # Max $1000 per trade
      USDC: 1000                    # Max $1000 USDC per trade
      BUSD: 1000                    # Max $1000 BUSD per trade
    
    # Order execution
    useTakerOrder: false            # Use maker orders for better fees
    dryRun: false                   # Live trading mode
    
    # Alert system
    largeAmountAlert:
      quoteCurrency: USDT
      amount: 5000                  # Alert for trades > $5000
      slack:
        channel: "trading-alerts"
        mentions:
        - '<@USER_ID>'              # Mention specific user
        - '<!subteam^TEAM_ID>'      # Mention team
```

## Balance Alignment Logic

### Deviation Detection Process
1. **Continuous Monitoring**: Checks balances at specified intervals
2. **Deviation Calculation**: `deviation = |current_balance - expected_balance| / expected_balance`
3. **Tolerance Check**: Compares deviation against `balanceToleranceRange`
4. **Sustained Tracking**: Records deviations that exceed tolerance
5. **Action Trigger**: Acts when deviation persists for `for` duration

### Trade Execution Logic
1. **Delta Calculation**: Determines how much to buy or sell
2. **Quote Currency Selection**: Chooses optimal quote currency based on trade direction
3. **Market Selection**: Finds exchange with best liquidity/pricing
4. **Order Placement**: Executes trade with appropriate order type
5. **Confirmation**: Verifies trade execution and updates tracking

## Quote Currency Selection

### Buy Orders
```yaml
quoteCurrencies:
  buy: [USDT, USDC, BUSD]
```
- Strategy tries USDT first, then USDC, then BUSD
- Selects first available pair with sufficient liquidity
- Considers balance availability on target exchange

### Sell Orders
```yaml
quoteCurrencies:
  sell: [USDT, USDC]
```
- Strategy tries USDT first, then USDC
- Optimizes for best price and liquidity
- Ensures sufficient quote currency balance for settlement

## Risk Management Features

### Balance Tolerance
- **Percentage-based**: Uses relative tolerance (e.g., 2% of expected balance)
- **Sustained Duration**: Only acts on persistent deviations
- **Maximum Amounts**: Caps individual trade sizes per currency

### Order Management
- **Maker Orders**: Default to maker orders for better fee rates
- **Taker Orders**: Optional taker orders for immediate execution
- **Market Selection**: Chooses optimal exchange for each trade

### Alert System
- **Large Amount Alerts**: Notifications for significant trades
- **Deviation Warnings**: Early warnings before action thresholds
- **Slack Integration**: Real-time notifications with user mentions

## Use Cases

### 1. Cross-Exchange Portfolio Balancing
```yaml
expectedBalances:
  BTC: 2.0
  ETH: 20.0
  USDT: 100000
balanceToleranceRange: 1%
for: 10m
```
**Purpose**: Maintain consistent portfolio allocation across exchanges
**Benefit**: Reduces concentration risk and optimizes capital efficiency

### 2. Arbitrage Opportunity Preparation
```yaml
expectedBalances:
  BTC: 0.5
  USDT: 25000
balanceToleranceRange: 0.5%
for: 2m
useTakerOrder: true
```
**Purpose**: Quickly realign balances for arbitrage opportunities
**Benefit**: Faster execution with taker orders for time-sensitive opportunities

### 3. Conservative Balance Management
```yaml
expectedBalances:
  BTC: 1.0
  USDT: 50000
balanceToleranceRange: 5%
for: 30m
maxAmounts:
  USDT: 500
```
**Purpose**: Gradual balance adjustments with conservative limits
**Benefit**: Minimizes market impact and trading costs

## Performance Optimization

### Monitoring Frequency
- **High Frequency (30s-1m)**: For active arbitrage strategies
- **Medium Frequency (5m-15m)**: For general portfolio management
- **Low Frequency (1h+)**: For long-term allocation maintenance

### Tolerance Tuning
- **Tight Tolerance (0.5-1%)**: More frequent rebalancing, higher precision
- **Medium Tolerance (2-3%)**: Balanced approach for most use cases
- **Loose Tolerance (5%+)**: Minimal rebalancing, lower costs

## Best Practices

1. **Start with Dry Run**: Test configuration before live trading
2. **Conservative Limits**: Begin with small maximum amounts
3. **Monitor Alerts**: Set up proper Slack notifications
4. **Regular Review**: Periodically assess expected balance targets
5. **Fee Consideration**: Use maker orders when possible to reduce costs
6. **Exchange Selection**: Ensure good liquidity on all monitored exchanges

## Limitations

1. **Market Impact**: Large rebalancing trades may affect market prices
2. **Exchange Limits**: Subject to individual exchange trading limits
3. **Network Latency**: Cross-exchange coordination may have delays
4. **Fee Costs**: Frequent rebalancing can accumulate significant fees
5. **Market Conditions**: Extreme volatility may trigger excessive trading

## Troubleshooting

### Common Issues

**No Rebalancing Actions**
- Check if deviations exceed tolerance threshold
- Verify sustained duration requirements are met
- Ensure sufficient balances for trading

**Excessive Trading**
- Increase tolerance range
- Extend sustained duration requirement
- Review expected balance targets

**Alert Spam**
- Adjust large amount alert thresholds
- Review tolerance settings
- Check for exchange connectivity issues

**Order Failures**
- Verify exchange API permissions
- Check minimum order requirements
- Ensure sufficient quote currency balances

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Configuration Examples](../../../config/xalign.yaml)
- [Cross-Exchange Trading Guide](../../doc/topics/cross-exchange.md)
- [Balance Management Best Practices](../../doc/topics/balance-management.md)
