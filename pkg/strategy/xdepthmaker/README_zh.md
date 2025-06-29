# 跨交易所深度做市策略 (XDepthMaker)

## 概述

XDepthMaker策略是一個先進的跨交易所做市策略，通過基於源交易所訂單簿深度放置多層訂單來提供深度流動性。與使用固定價差的傳統做市商不同，XDepthMaker根據累積深度需求和源交易所的訂單簿結構動態計算訂單價格和數量。

## 工作原理

1. **深度分析**：監控源（對沖）交易所的訂單簿深度
2. **動態定價**：基於累積深度需求計算訂單價格
3. **分層訂單**：放置具有深度縮放數量的多個訂單層
4. **倉位對沖**：在源交易所自動對沖已成交的訂單
5. **自適應深度**：當WebSocket深度不足時回退到RESTful API
6. **風險管理**：實施全面的倉位和餘額控制

## 主要特性

- **基於深度的做市**：訂單基於所需深度而非固定價差放置
- **多層架構**：可配置的訂單層數和深度縮放
- **智能深度計算**：使用累積深度確定最優定價
- **自適應數據源**：在WebSocket和REST API之間切換獲取深度數據
- **多種對沖策略**：市價、BBO對手方和BBO排隊對沖選項
- **快速層更新**：分離的快速和完整補充間隔
- **全面監控**：內建Prometheus指標用於深度和價差分析
- **交易恢復**：自動交易恢復機制處理遺漏的交易

## 策略架構

### 核心組件

1. **深度分析器**：分析源交易所訂單簿深度
2. **訂單生成器**：基於深度需求創建做市訂單
3. **對沖引擎**：使用多種策略管理倉位對沖
4. **報價工作器**：處理快速層更新和完整補充
5. **指標收集器**：監控價差比率和深度指標

### 對沖策略

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

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 做市交易所的交易對符號 |
| `hedgeSymbol` | 字符串 | 否 | 對沖交易所的交易對符號（默認為symbol） |
| `makerExchange` | 字符串 | 是 | 做市交易所會話名稱 |
| `hedgeExchange` | 字符串 | 是 | 對沖交易所會話名稱 |

### 更新間隔
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `fastLayerUpdateInterval` | 持續時間 | 否 | 快速層更新間隔（默認：5s） |
| `fullReplenishInterval` | 持續時間 | 否 | 完整補充間隔（默認：10m） |
| `hedgeInterval` | 持續時間 | 否 | 對沖執行間隔（默認：3s） |

### 保證金配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `margin` | 百分比 | 否 | 雙邊默認保證金（默認：0.3%） |
| `bidMargin` | 百分比 | 否 | 特定買單保證金（覆蓋margin） |
| `askMargin` | 百分比 | 否 | 特定賣單保證金（覆蓋margin） |

### 層配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `numLayers` | 整數 | 否 | 每側的訂單層數（默認：1） |
| `numOfFastLayers` | 整數 | 否 | 快速更新層數（默認：5） |
| `pips` | 數字 | 是 | 層間價格增量乘數 |

### 深度縮放
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `depthScale` | 對象 | 是 | 深度縮放配置 |
| `depthScale.byLayer` | 對象 | 否 | 基於層的深度縮放 |
| `depthScale.byLayer.linear` | 對象 | 否 | 線性縮放配置 |
| `depthScale.byLayer.exp` | 對象 | 否 | 指數縮放配置 |

### 數量配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quantity` | 數字 | 否 | 第一層的固定數量 |
| `quantityScale` | 對象 | 否 | 基於層的數量縮放 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `maxExposurePosition` | 數字 | 否 | 最大未對沖倉位 |
| `hedgeMaxOrderQuantity` | 數字 | 否 | 最大對沖訂單數量 |
| `stopHedgeQuoteBalance` | 數字 | 否 | 低於此報價餘額停止對沖 |
| `stopHedgeBaseBalance` | 數字 | 否 | 低於此基礎餘額停止對沖 |
| `disableHedge` | 布爾值 | 否 | 禁用所有對沖 |

### 對沖策略
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `hedgeStrategy` | 字符串 | 否 | 對沖策略類型（默認："market"） |

### 高級功能
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `recoverTrade` | 布爾值 | 否 | 啟用交易恢復 |
| `recoverTradeScanPeriod` | 持續時間 | 否 | 交易恢復掃描期間 |
| `notifyTrade` | 布爾值 | 否 | 啟用交易通知 |
| `skipCleanUpOpenOrders` | 布爾值 | 否 | 啟動時跳過清理未完成訂單 |
| `priceImpactRatio` | 百分比 | 否 | BBO監控的價格影響比率 |

## 配置示例

```yaml
crossExchangeStrategies:
  - xdepthmaker:
      symbol: BTCUSDT
      makerExchange: max          # 在MAX做市
      hedgeExchange: binance      # 在幣安對沖
      
      # 更新間隔
      fastLayerUpdateInterval: 5s
      fullReplenishInterval: 10m
      hedgeInterval: 3s
      
      # 保證金配置
      margin: 0.4%                # 0.4%默認保證金
      bidMargin: 0.35%            # 稍緊的買單保證金
      askMargin: 0.45%            # 稍寬的賣單保證金
      
      # 層配置
      numLayers: 30               # 每側30層
      numOfFastLayers: 5          # 快速更新前5層
      pips: 10                    # 層間10倍tick大小
      
      # 深度縮放配置
      depthScale:
        byLayer:
          linear:
            domain: [1, 30]       # 第1到30層
            range: [50, 20000]    # 深度從$50到$20,000
      
      # 風險管理
      maxExposurePosition: 0.1    # 最大0.1 BTC未對沖
      hedgeMaxOrderQuantity: 0.5  # 每筆對沖訂單最大0.5 BTC
      stopHedgeQuoteBalance: 1000 # 低於$1,000停止對沖
      stopHedgeBaseBalance: 0.01  # 低於0.01 BTC停止對沖
      
      # 對沖策略
      hedgeStrategy: "bbo-counter-party-1"  # 使用BBO對手方第1級
      
      # 高級功能
      recoverTrade: true
      recoverTradeScanPeriod: 30m
      notifyTrade: true
      priceImpactRatio: 0.1%
      
      # 利潤修復器（可選）
      profitFixer:
        tradesSince: "2024-01-01T00:00:00Z"
```

## 策略邏輯

### 基於深度的訂單生成

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) generateMakerOrders(
	sourceBook *types.StreamOrderBook,
	maxLayer int,
	availableBase, availableQuote fixedpoint.Value,
) ([]types.SubmitOrder, error) {
	// 計算每層所需深度
	requiredDepthFloat, err := s.DepthScale.Scale(i)
	if err != nil {
		return nil, errors.Wrapf(err, "depthScale scale error")
	}
	
	// 累積深度需求
	requiredDepth := fixedpoint.NewFromFloat(requiredDepthFloat)
	accumulatedDepth = accumulatedDepth.Add(requiredDepth)
	
	// 找到滿足深度需求的價格水平
	index := sideBook.IndexByQuoteVolumeDepth(accumulatedDepth)
	depthPrice := pvs.AverageDepthPriceByQuote(accumulatedDepth, 0)
```
</augment_code_snippet>

### 自適應深度數據源

<augment_code_snippet path="pkg/strategy/xdepthmaker/strategy.go" mode="EXCERPT">
```go
// 檢查WebSocket深度是否足夠
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

### 對沖執行策略

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

## 深度縮放配置

### 線性縮放
```yaml
depthScale:
  byLayer:
    linear:
      domain: [1, 20]     # 第1到20層
      range: [100, 5000]  # 深度從$100到$5,000
```

### 指數縮放
```yaml
depthScale:
  byLayer:
    exp:
      domain: [1, 20]     # 第1到20層
      range: [1, 3]       # 指數縮放從1倍到3倍
```

## 對沖策略選項

### 1. 市價對沖 (`market`)
- **描述**：使用市價訂單立即執行
- **優點**：保證執行，最小滑點風險
- **缺點**：較高交易成本，潛在市場影響

### 2. BBO對手方 (`bbo-counter-party-N`)
- **描述**：在對手方第N級放置限價訂單
- **級別**：1、3、5（更深級別獲得更好價格）
- **優點**：更好的執行價格，減少市場影響
- **缺點**：執行風險，潛在部分成交

### 3. BBO排隊 (`bbo-queue-1`)
- **描述**：在最佳買賣價放置訂單加入排隊
- **優點**：最佳可能價格，做市商回扣
- **缺點**：排隊位置風險，執行較慢

## 風險管理功能

### 1. 倉位風險敞口控制
- 監控未對沖倉位與 `maxExposurePosition` 的對比
- 當風險敞口限制被超過時自動停止放置訂單
- 分別跟踪基礎和報價貨幣風險敞口

### 2. 餘額保護
- 在兩個交易所維持最小餘額
- 防止超出可用資金的過度交易
- 基礎和報價貨幣的可配置餘額閾值

### 3. 對沖訂單限制
- 通過 `hedgeMaxOrderQuantity` 限制單個對沖訂單規模
- 防止過度的單筆訂單市場影響
- 幫助管理執行風險

### 4. 交易恢復
- 通過REST API自動掃描遺漏的交易
- 可配置的掃描期間和重疊緩衝
- 確保準確的倉位和盈虧跟踪

## 監控和指標

策略提供全面的Prometheus指標：

- **價差指標**：市場價差比率和趨勢
- **深度指標**：按價格範圍的USD訂單簿深度
- **價格水平指標**：範圍內的價格水平數量
- **倉位指標**：風險敞口水平和對沖比率
- **性能指標**：訂單放置成功率

### 關鍵指標
- `bbgo_xdepthmaker_market_spread_ratio`：當前市場價差比率
- `bbgo_xdepthmaker_depth_in_usd`：按方向和價格範圍的USD市場深度
- `bbgo_xdepthmaker_price_level_count`：範圍內的價格水平數量

## 常見用例

### 1. 深度流動性提供
```yaml
numLayers: 50
depthScale:
  byLayer:
    linear:
      domain: [1, 50]
      range: [100, 50000]
hedgeStrategy: "market"
```

### 2. 保守做市
```yaml
numLayers: 10
margin: 0.5%
maxExposurePosition: 0.05
hedgeStrategy: "bbo-counter-party-3"
```

### 3. 高頻更新
```yaml
fastLayerUpdateInterval: 1s
numOfFastLayers: 10
fullReplenishInterval: 5m
hedgeStrategy: "bbo-queue-1"
```

## 最佳實踐

1. **保守開始**：從較少的層數和更寬的保證金開始
2. **監控深度需求**：確保源交易所有足夠的深度
3. **餘額管理**：在兩個交易所維持足夠的餘額
4. **對沖策略選擇**：為市場條件選擇適當的對沖策略
5. **定期監控**：監控指標並根據需要調整參數
6. **交易恢復**：啟用交易恢復以確保準確跟踪

## 故障排除

### 常見問題

**深度不足**
- 增加 `depthScale` 範圍值
- 減少層數
- 檢查源交易所流動性

**對沖失敗**
- 驗證對沖交易所連接性
- 檢查可用餘額
- 檢查對沖策略選擇

**訂單放置問題**
- 檢查保證金設置
- 驗證最小訂單規模
- 檢查餘額可用性

**性能問題**
- 優化更新間隔
- 減少層數
- 監控網絡延遲

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [跨交易所交易指南](../../doc/topics/cross-exchange.md)
- [配置示例](../../../config/xdepthmaker.yaml)
- [XMaker策略](../xmaker/README.md) - 替代跨交易所方法
