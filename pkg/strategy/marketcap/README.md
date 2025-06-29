# Market Cap Strategy

## Overview

The Market Cap strategy is a portfolio rebalancing strategy that automatically adjusts your cryptocurrency holdings based on market capitalization data from CoinMarketCap. This strategy helps maintain a diversified portfolio that reflects the relative market values of different cryptocurrencies.

## How It Works

1. **Data Collection**: The strategy fetches real-time market capitalization data from CoinMarketCap API
2. **Weight Calculation**: It calculates target portfolio weights based on the relative market caps of your selected cryptocurrencies
3. **Portfolio Rebalancing**: The strategy automatically places buy/sell orders to align your actual portfolio with the target weights
4. **Continuous Monitoring**: It runs continuously, rebalancing at specified intervals

## Key Features

- **Market Cap Based Allocation**: Automatically allocates funds based on cryptocurrency market capitalizations
- **Configurable Rebalancing**: Set custom intervals for portfolio rebalancing
- **Threshold Control**: Only rebalance when the difference exceeds a specified threshold
- **Risk Management**: Built-in maximum order amount limits
- **Dry Run Mode**: Test the strategy without placing real orders
- **Multiple Order Types**: Support for LIMIT, LIMIT_MAKER, and MARKET orders

## Prerequisites

### API Key Setup
You must obtain a CoinMarketCap API key and set it as an environment variable:

```bash
export COINMARKETCAP_API_KEY="your_api_key_here"
```

Get your free API key from: https://coinmarketcap.com/api/

### Exchange Configuration
Ensure your exchange session is properly configured in your BBGO configuration file.

## Configuration Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `interval` | string | Yes | Rebalancing interval (e.g., "5m", "1h", "1d") |
| `quoteCurrency` | string | Yes | Base currency for portfolio (e.g., "USDT", "TWD") |
| `quoteCurrencyWeight` | percentage | Yes | Weight allocation for quote currency |
| `baseCurrencies` | array | Yes | List of cryptocurrencies to include in portfolio |
| `threshold` | percentage | Yes | Minimum weight difference to trigger rebalancing |
| `maxAmount` | number | No | Maximum amount per order in quote currency |
| `queryInterval` | string | No | Interval to fetch market cap data (default: "1h") |
| `orderType` | string | No | Order type: LIMIT, LIMIT_MAKER, MARKET (default: LIMIT_MAKER) |
| `dryRun` | boolean | No | Enable dry run mode (default: false) |

## Configuration Example

```yaml
exchangeStrategies:
  - on: binance
    marketcap:
      interval: 1h                    # Rebalance every hour
      quoteCurrency: USDT            # Use USDT as base currency
      quoteCurrencyWeight: 10%       # Keep 10% in USDT
      baseCurrencies:                # Portfolio cryptocurrencies
        - BTC
        - ETH
        - BNB
        - ADA
        - DOT
      threshold: 2%                  # Rebalance when difference > 2%
      maxAmount: 1000               # Max $1000 per order
      queryInterval: 30m            # Update market cap data every 30 minutes
      orderType: LIMIT_MAKER        # Use maker orders for better fees
      dryRun: false                 # Execute real trades
```

## Strategy Logic

### Weight Calculation
1. Fetch market capitalization data for all `baseCurrencies`
2. Calculate relative weights based on market cap proportions
3. Scale weights by `(1 - quoteCurrencyWeight)` to reserve space for quote currency
4. Add quote currency weight to the target allocation

### Rebalancing Decision
For each currency in the portfolio:
1. Calculate current weight: `current_value / total_portfolio_value`
2. Compare with target weight from market cap data
3. If `|current_weight - target_weight| > threshold`, trigger rebalancing
4. Calculate required buy/sell quantity to reach target weight

### Order Execution
1. Cancel any existing active orders
2. Generate new orders based on weight differences
3. Apply `maxAmount` limits if specified
4. Submit orders to the exchange (unless in dry run mode)

## Risk Management

### Built-in Protections
- **Threshold Control**: Prevents excessive trading on minor weight changes
- **Maximum Order Size**: Limits exposure per individual trade
- **Dry Run Mode**: Allows strategy testing without financial risk
- **Order Type Selection**: Choose appropriate order types for your risk tolerance

### Recommended Settings
- **Conservative**: `threshold: 5%`, `maxAmount: 500`, `orderType: LIMIT_MAKER`
- **Moderate**: `threshold: 2%`, `maxAmount: 1000`, `orderType: LIMIT_MAKER`
- **Aggressive**: `threshold: 1%`, `maxAmount: 2000`, `orderType: MARKET`

## Monitoring and Logging

The strategy provides detailed logging for:
- Market cap data updates
- Weight calculations and comparisons
- Order generation and execution
- Rebalancing decisions and thresholds

Monitor logs to ensure the strategy is working as expected:
```bash
bbgo run --config your_config.yaml --verbose
```

## Common Use Cases

### 1. Top 10 Crypto Portfolio
Maintain a portfolio of the top 10 cryptocurrencies by market cap:
```yaml
baseCurrencies: [BTC, ETH, BNB, XRP, ADA, DOGE, MATIC, SOL, DOT, LTC]
quoteCurrencyWeight: 5%
threshold: 3%
```

### 2. Conservative DeFi Portfolio
Focus on established DeFi tokens with higher cash allocation:
```yaml
baseCurrencies: [ETH, BNB, UNI, AAVE, COMP]
quoteCurrencyWeight: 20%
threshold: 5%
```

### 3. Aggressive Growth Portfolio
Target emerging cryptocurrencies with frequent rebalancing:
```yaml
baseCurrencies: [ETH, SOL, AVAX, NEAR, FTM, ATOM]
quoteCurrencyWeight: 5%
threshold: 1%
interval: 30m
```

## Troubleshooting

### Common Issues

**API Key Errors**
- Ensure `COINMARKETCAP_API_KEY` environment variable is set
- Verify API key is valid and has sufficient quota
- Check API key permissions

**Insufficient Balance**
- Ensure adequate balance in quote currency for rebalancing
- Consider reducing `maxAmount` or increasing `threshold`

**Order Failures**
- Check exchange-specific minimum order sizes
- Verify trading pairs are available on your exchange
- Ensure sufficient balance for trading fees

**Market Data Issues**
- Verify cryptocurrency symbols match CoinMarketCap listings
- Check network connectivity
- Monitor API rate limits

### Debug Mode
Enable verbose logging for troubleshooting:
```bash
bbgo run --config your_config.yaml --debug
```

## Performance Considerations

- **API Limits**: CoinMarketCap free tier has rate limits; adjust `queryInterval` accordingly
- **Trading Fees**: Frequent rebalancing incurs trading fees; balance `threshold` vs. optimization
- **Market Impact**: Large orders may impact prices; use appropriate `maxAmount` settings
- **Latency**: Consider exchange latency when setting short rebalancing intervals

## See Also

- [BBGO Strategy Development Guide](../../doc/topics/developing-strategy.md)
- [Rebalance Strategy](../rebalance/README.md) - Alternative rebalancing approach
- [Configuration Examples](../../../config/marketcap.yaml)
- [CoinMarketCap API Documentation](https://coinmarketcap.com/api/documentation/v1/)
