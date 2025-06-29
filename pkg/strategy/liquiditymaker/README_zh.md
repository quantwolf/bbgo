# 流動性做市策略

## 概述

流動性做市策略是一個複雜的做市策略，旨在通過在當前市場價格周圍放置多層買賣訂單來為訂單簿提供流動性。該策略旨在從買賣價差中獲利，同時通過各種保護機制維持平衡倉位和管理風險。

## 工作原理

1. **流動性提供**：在買賣雙方放置多層限價訂單
2. **動態定價**：使用中間價或最後交易價作為訂單放置的參考
3. **倉位管理**：根據當前倉位和風險敞口自動調整訂單
4. **風險控制**：實施各種止損機制和利潤保護功能
5. **持續再平衡**：以可配置的間隔更新訂單

## 主要特性

- **多層訂單放置**：創建可配置層數的流動性深度
- **可擴展訂單規模**：使用指數或線性縮放來調整訂單數量
- **倉位感知交易**：根據當前倉位風險敞口調整策略
- **利潤保護**：確保平倉訂單的最低利潤率
- **基於EMA的控制**：可選的EMA過濾器用於價格偏差保護
- **全面指標**：內建Prometheus指標用於監控
- **靈活配置**：針對不同市場條件的高度可定制參數

## 策略組件

### 1. 流動性訂單
- **目的**：提供市場流動性並捕獲價差
- **放置**：圍繞中間價的多層訂單，具有可配置的價差
- **縮放**：跨層的指數或線性數量縮放
- **更新頻率**：由 `liquidityUpdateInterval` 控制

### 2. 調整訂單
- **目的**：以利潤保護方式平倉現有倉位
- **觸發**：當倉位不是塵埃（顯著規模）時激活
- **保護**：使用 `profitProtectedPrice` 函數確保最低利潤
- **更新頻率**：由 `adjustmentUpdateInterval` 控制

## 配置參數

### 核心設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `liquidityUpdateInterval` | 間隔 | 否 | 更新流動性訂單的間隔（默認："1h"） |
| `adjustmentUpdateInterval` | 間隔 | 否 | 更新調整訂單的間隔（默認："5m"） |

### 流動性配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `numOfLiquidityLayers` | 整數 | 是 | 每側的訂單層數 |
| `askLiquidityAmount` | 數字 | 是 | 賣單的總流動性金額 |
| `bidLiquidityAmount` | 數字 | 是 | 買單的總流動性金額 |
| `liquidityPriceRange` | 百分比 | 是 | 流動性訂單的價格範圍 |
| `askLiquidityPriceRange` | 百分比 | 否 | 特定賣單價格範圍（覆蓋liquidityPriceRange） |
| `bidLiquidityPriceRange` | 百分比 | 否 | 特定買單價格範圍（覆蓋liquidityPriceRange） |

### 縮放配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `liquidityScale` | 對象 | 是 | 訂單數量的縮放函數 |
| `liquidityScale.exp` | 對象 | 否 | 指數縮放配置 |
| `liquidityScale.linear` | 對象 | 否 | 線性縮放配置 |

### 價格和價差設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `spread` | 百分比 | 是 | 從中間價的最小價差 |
| `useLastTradePrice` | 布爾值 | 否 | 使用最後交易價而不是中間價 |
| `maxPrice` | 數字 | 否 | 訂單的最大允許價格 |
| `minPrice` | 數字 | 否 | 訂單的最小允許價格 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `maxPositionExposure` | 數字 | 否 | 允許的最大倉位規模 |
| `minProfit` | 百分比 | 否 | 調整訂單的最小利潤率 |
| `stopBidPrice` | 數字 | 否 | 停止放置買單的價格水平 |
| `stopAskPrice` | 數字 | 否 | 停止放置賣單的價格水平 |
| `useProtectedPriceRange` | 布爾值 | 否 | 為流動性訂單啟用利潤保護 |

### 高級功能
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `adjustmentOrderMaxQuantity` | 數字 | 否 | 調整訂單的最大數量 |
| `adjustmentOrderPriceType` | 字符串 | 否 | 調整訂單的價格類型（默認："MAKER"） |

### EMA控制
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `midPriceEMA` | 對象 | 否 | 中間價EMA偏差保護 |
| `midPriceEMA.enabled` | 布爾值 | 否 | 啟用中間價EMA保護 |
| `midPriceEMA.interval` | 間隔 | 否 | EMA計算間隔 |
| `midPriceEMA.window` | 整數 | 否 | EMA窗口大小 |
| `midPriceEMA.maxBiasRatio` | 百分比 | 否 | 最大允許偏差比率 |
| `stopEMA` | 對象 | 否 | 停止EMA配置 |
| `stopEMA.enabled` | 布爾值 | 否 | 啟用停止EMA |
| `stopEMA.interval` | 間隔 | 否 | 停止EMA間隔 |
| `stopEMA.window` | 整數 | 否 | 停止EMA窗口 |

## 配置示例

```yaml
exchangeStrategies:
  - on: binance
    liquiditymaker:
      symbol: BTCUSDT
      
      # 更新間隔
      liquidityUpdateInterval: 1h
      adjustmentUpdateInterval: 5m
      
      # 流動性配置
      numOfLiquidityLayers: 20
      askLiquidityAmount: 10000.0    # 價值$10,000的賣單流動性
      bidLiquidityAmount: 10000.0    # 價值$10,000的買單流動性
      liquidityPriceRange: 2%        # 訂單的2%價格範圍
      
      # 價差和定價
      spread: 0.1%                   # 0.1%最小價差
      useLastTradePrice: true        # 使用最後交易價作為參考
      
      # 訂單縮放（指數）
      liquidityScale:
        exp:
          domain: [1, 20]            # 層數1到20
          range: [1, 4]              # 從1倍縮放到4倍
      
      # 風險管理
      maxPositionExposure: 1.0       # 最大1 BTC倉位
      minProfit: 0.05%               # 調整時最小0.05%利潤
      useProtectedPriceRange: true   # 啟用利潤保護
      
      # 停止水平
      stopBidPrice: 25000            # 在$25,000以下停止買單
      stopAskPrice: 45000            # 在$45,000以上停止賣單
      
      # EMA保護
      midPriceEMA:
        enabled: true
        interval: 5m
        window: 20
        maxBiasRatio: 1%             # 最大1%EMA偏差
```

## 策略邏輯

### 流動性訂單放置

<augment_code_snippet path="pkg/strategy/liquiditymaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) placeLiquidityOrders(ctx context.Context) {
	// 取消現有訂單
	err := s.liquidityOrderBook.GracefulCancel(ctx, s.Session.Exchange)
	
	// 計算中間價和價差
	midPrice := ticker.Sell.Add(ticker.Buy).Div(fixedpoint.Two)
	sideSpread := s.Spread.Div(fixedpoint.Two)
	
	ask1Price := midPrice.Mul(fixedpoint.One.Add(sideSpread))
	bid1Price := midPrice.Mul(fixedpoint.One.Sub(sideSpread))
```
</augment_code_snippet>

### 訂單生成

<augment_code_snippet path="pkg/strategy/liquiditymaker/generator.go" mode="EXCERPT">
```go
func (g *LiquidityOrderGenerator) Generate(
	side types.SideType, totalAmount, startPrice, endPrice fixedpoint.Value, numLayers int, scale bbgo.Scale,
) (orders []types.SubmitOrder) {
	// 計算層價差和價格
	layerSpread := endPrice.Sub(startPrice).Div(fixedpoint.NewFromInt(int64(numLayers - 1)))
	
	// 使用縮放生成訂單
	for i := 0; i < numLayers; i++ {
		layerPrice := startPrice.Add(layerSpread.Mul(fi))
		layerScale := scale.Call(float64(i + 1))
		quantity := fixedpoint.NewFromFloat(factor * layerScale)
```
</augment_code_snippet>

### 利潤保護

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

## 風險管理功能

### 1. 倉位風險敞口控制
- 監控總倉位規模與 `maxPositionExposure` 的對比
- 當風險敞口限制被超過時禁用訂單放置
- 對多頭和空頭倉位分別控制

### 2. 停止價格機制
- `stopBidPrice`：防止在指定價格以上放置買單
- `stopAskPrice`：防止在指定價格以下放置賣單
- `stopEMA`：基於EMA指標的動態停止水平

### 3. 利潤保護
- 確保調整訂單維持最低利潤率
- 在利潤計算中考慮交易手續費
- 通過 `useProtectedPriceRange` 為流動性訂單提供可選保護

### 4. EMA偏差保護
- 防止在極端價格波動期間放置訂單
- 將當前價格與EMA平滑價格進行比較
- 可配置的偏差比率閾值

## 監控和指標

策略提供全面的Prometheus指標：

- **價格指標**：中間價、買賣價、價差
- **流動性指標**：訂單風險敞口、流動性金額
- **倉位指標**：當前倉位、風險敞口水平
- **性能指標**：訂單放置狀態、偏差比率

### 關鍵指標
- `liqmaker_spread`：當前市場價差
- `liqmaker_mid_price`：計算的中間價
- `liqmaker_open_order_bid_exposure_in_usd`：總買單風險敞口
- `liqmaker_open_order_ask_exposure_in_usd`：總賣單風險敞口
- `liqmaker_order_placement_status`：按方向的訂單放置狀態

## 常見用例

### 1. 高頻做市
```yaml
liquidityUpdateInterval: 1m
adjustmentUpdateInterval: 30s
numOfLiquidityLayers: 50
spread: 0.05%
```

### 2. 保守流動性提供
```yaml
liquidityUpdateInterval: 1h
adjustmentUpdateInterval: 5m
numOfLiquidityLayers: 10
spread: 0.2%
maxPositionExposure: 0.5
```

### 3. 波動市場適應
```yaml
midPriceEMA:
  enabled: true
  maxBiasRatio: 2%
stopEMA:
  enabled: true
useProtectedPriceRange: true
```

## 最佳實踐

1. **保守開始**：從更寬的價差和較小的流動性金額開始
2. **監控指標**：使用Prometheus指標跟踪性能
3. **適應波動性**：在高波動期間增加價差
4. **倉位管理**：設置適當的 `maxPositionExposure` 限制
5. **手續費優化**：使用掛單來最小化交易手續費
6. **回測**：在實盤部署前徹底測試配置

## 故障排除

### 常見問題

**訂單未放置**
- 檢查餘額可用性
- 驗證最小訂單規模要求
- 檢查停止價格配置

**過度倉位風險敞口**
- 調整 `maxPositionExposure` 設置
- 檢查調整訂單頻率
- 檢查利潤保護設置

**性能不佳**
- 分析價差與市場波動性的關係
- 檢查訂單層分佈
- 監控成交率和庫存周轉

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/liquiditymaker.yaml)
- [做市最佳實踐](../../doc/topics/market-making.md)
