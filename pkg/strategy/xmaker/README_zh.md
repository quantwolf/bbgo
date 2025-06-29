# 跨交易所做市策略 (XMaker)

## 概述

XMaker策略是一個先進的跨交易所做市策略，在一個交易所（做市交易所）提供流動性，同時在另一個交易所（源交易所）對沖倉位。該策略通過複雜的對沖機制捕獲不同交易所之間的套利機會和價差，同時保持中性倉位。

## 工作原理

1. **價格發現**：監控源交易所的訂單簿和價格數據
2. **報價生成**：在做市交易所創建具有可配置保證金的買賣報價
3. **訂單放置**：在做市交易所放置多層限價訂單
4. **倉位對沖**：在源交易所自動對沖已成交的訂單
5. **信號整合**：使用多個信號源調整保證金和交易行為
6. **風險管理**：實施全面的風險控制和熔斷機制

## 主要特性

- **跨交易所套利**：捕獲交易所之間的價格差異
- **多層做市**：放置具有可配置縮放的多個訂單層
- **高級對沖**：直接對沖、延遲對沖和合成對沖選項
- **基於信號的保證金調整**：基於市場信號的動態保證金調整
- **布林帶整合**：使用布林帶的趨勢基礎保證金調整
- **價差做市**：用於倉位管理的智能價差做市訂單
- **熔斷保護**：過度虧損時自動停止交易
- **全面監控**：內建指標和性能跟踪

## 策略架構

### 核心組件

1. **報價引擎**：生成帶有保證金和手續費的買賣價格
2. **訂單管理器**：處理多層訂單放置和取消
3. **對沖引擎**：管理跨交易所的倉位對沖
4. **信號處理器**：聚合多個市場信號
5. **風險控制器**：監控和控制交易風險

### 信號類型

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

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號 |
| `sourceExchange` | 字符串 | 是 | 用於對沖的源交易所會話名稱 |
| `makerExchange` | 字符串 | 是 | 用於做市的做市交易所會話名稱 |
| `updateInterval` | 持續時間 | 否 | 報價更新間隔（默認："1s"） |
| `hedgeInterval` | 持續時間 | 否 | 對沖執行間隔（默認："10s"） |

### 保證金配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `margin` | 百分比 | 是 | 雙邊默認保證金 |
| `bidMargin` | 百分比 | 否 | 特定買單保證金（覆蓋margin） |
| `askMargin` | 百分比 | 否 | 特定賣單保證金（覆蓋margin） |
| `minMargin` | 百分比 | 否 | 最小保證金保護 |

### 訂單配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quantity` | 數字 | 是 | 第一層的基礎數量 |
| `quantityMultiplier` | 數字 | 否 | 後續層的乘數 |
| `quantityScale` | 對象 | 否 | 基於層的數量縮放 |
| `numLayers` | 整數 | 是 | 每側的訂單層數 |
| `pips` | 數字 | 是 | 層間價格增量 |
| `makerOnly` | 布爾值 | 否 | 使用僅掛單訂單 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `maxExposurePosition` | 數字 | 否 | 最大未對沖倉位 |
| `maxHedgeAccountLeverage` | 數字 | 否 | 對沖賬戶最大槓桿 |
| `maxHedgeQuoteQuantityPerOrder` | 數字 | 否 | 最大對沖訂單規模 |
| `minMarginLevel` | 數字 | 否 | 保證金交易的最小保證金水平 |
| `stopHedgeQuoteBalance` | 數字 | 否 | 低於此報價餘額停止對沖 |
| `stopHedgeBaseBalance` | 數字 | 否 | 低於此基礎餘額停止對沖 |
| `disableHedge` | 布爾值 | 否 | 禁用所有對沖 |

### 信號配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `enableSignalMargin` | 布爾值 | 否 | 啟用基於信號的保證金調整 |
| `signals` | 數組 | 否 | 信號配置列表 |
| `signalReverseSideMargin` | 對象 | 否 | 反向信號的保證金調整 |
| `signalTrendSideMarginDiscount` | 對象 | 否 | 趨勢信號的保證金折扣 |

### 高級功能
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `enableDelayHedge` | 布爾值 | 否 | 啟用延遲對沖 |
| `maxHedgeDelayDuration` | 持續時間 | 否 | 最大對沖延遲 |
| `delayHedgeSignalThreshold` | 數字 | 否 | 延遲的信號閾值 |
| `enableBollBandMargin` | 布爾值 | 否 | 啟用布林帶保證金 |
| `enableArbitrage` | 布爾值 | 否 | 啟用套利模式 |
| `useDepthPrice` | 布爾值 | 否 | 使用深度加權定價 |

## 配置示例

```yaml
crossExchangeStrategies:
  - xmaker:
      symbol: BTCUSDT
      sourceExchange: binance    # 在幣安對沖
      makerExchange: max         # 在MAX做市
      
      # 更新間隔
      updateInterval: 1s
      hedgeInterval: 10s
      
      # 保證金配置
      margin: 0.4%               # 0.4%默認保證金
      bidMargin: 0.35%           # 稍緊的買單保證金
      askMargin: 0.45%           # 稍寬的賣單保證金
      minMargin: 0.1%            # 最小保證金保護
      
      # 訂單配置
      quantity: 0.001            # 每層0.001 BTC
      quantityMultiplier: 1.5    # 每層數量增加1.5倍
      numLayers: 3               # 每側3層
      pips: 10                   # 層間$10
      makerOnly: true            # 使用僅掛單訂單
      
      # 風險管理
      maxExposurePosition: 0.1   # 最大0.1 BTC未對沖
      maxHedgeAccountLeverage: 3 # 最大3倍槓桿
      minMarginLevel: 1.5        # 維持1.5保證金水平
      
      # 基於信號的保證金調整
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
      
      # 信號保證金配置
      signalTrendSideMarginDiscount:
        enabled: true
        threshold: 0.5
        scale:
          linear:
            domain: [0.5, 2.0]
            range: [0, 0.2]
      
      # 延遲對沖
      enableDelayHedge: true
      maxHedgeDelayDuration: 30s
      delayHedgeSignalThreshold: 1.0
      
      # 熔斷器
      circuitBreaker:
        enabled: true
        maximumConsecutiveTotalLoss: 50.0
        maximumConsecutiveLossTimes: 5
        haltDuration: 30m
```

## 策略邏輯

### 報價生成過程

<augment_code_snippet path="pkg/strategy/xmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) updateQuote(ctx context.Context) error {
	// 取消現有訂單
	if err := s.activeMakerOrders.GracefulCancel(ctx, s.makerSession.Exchange); err != nil {
		return nil
	}
	
	// 從源獲取最佳買賣價
	bestBid, bestAsk, hasPrice := s.sourceBook.BestBidAndAsk()
	if !hasPrice {
		return nil
	}
	
	// 應用保證金並生成報價
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

### 對沖機制

<augment_code_snippet path="pkg/strategy/xmaker/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) hedge(ctx context.Context, uncoveredPosition fixedpoint.Value) {
	if uncoveredPosition.IsZero() {
		return
	}
	
	// 計算對沖增量（未覆蓋倉位的相反）
	hedgeDelta := uncoveredPosition.Neg()
	side := positionToSide(hedgeDelta)
	
	// 檢查是否可以延遲對沖
	if s.canDelayHedge(side, hedgeDelta) {
		return
	}
	
	// 執行對沖訂單
	if _, err := s.directHedge(ctx, hedgeDelta); err != nil {
		log.WithError(err).Errorf("unable to hedge position")
	}
}
```
</augment_code_snippet>

## 信號整合

### 布林帶信號
- **目的**：檢測趨勢方向並相應調整保證金
- **邏輯**：在趨勢側增加保證金，在反趨勢側減少
- **配置**：間隔、窗口和帶寬參數

### 深度比率信號
- **目的**：分析訂單簿深度不平衡
- **邏輯**：基於買賣深度比率調整保證金
- **配置**：深度數量和平均期間

### 交易量信號
- **目的**：監控交易量模式
- **邏輯**：基於成交量趨勢調整行為
- **配置**：窗口大小和成交量閾值

## 風險管理功能

### 1. 倉位風險敞口控制
- 監控未對沖倉位與 `maxExposurePosition` 的對比
- 當風險敞口限制被超過時禁用訂單放置
- 對不同倉位方向分別控制

### 2. 保證金水平保護
- 監控源交易所保證金賬戶健康狀況
- 當保證金水平過低時防止對沖
- 計算安全對沖的可用債務配額

### 3. 熔斷器
- 連續虧損時自動停止交易
- 可配置的虧損閾值和停止持續時間
- 防止市場壓力期間的災難性損失

### 4. 餘額保護
- 在兩個交易所維持最小餘額
- 防止超出可用資金的過度交易
- 可配置的餘額閾值

## 高級功能

### 延遲對沖
- 當信號強烈時延遲對沖執行
- 允許從有利走勢中捕獲額外利潤
- 可配置的延遲持續時間和信號閾值

### 價差做市
- 放置智能價差做市訂單
- 幫助以有利價格平倉
- 可配置的利潤目標和訂單壽命

### 合成對沖
- 使用合成工具的替代對沖
- 當直接對沖不可用時有用
- 可配置的合成對關係

### 套利模式
- 檢測並執行套利機會
- 當價格差異超過保證金時放置IOC訂單
- 從價格差異中自動捕獲利潤

## 監控和指標

策略提供全面的指標：

- **價格指標**：源/做市價格、價差、保證金
- **倉位指標**：風險敞口、對沖比率、盈虧
- **訂單指標**：成交率、訂單放置成功率
- **信號指標**：聚合信號、保證金調整
- **性能指標**：盈虧、夏普比率、回撤

## 常見用例

### 1. 保守跨交易所做市
```yaml
margin: 0.5%
numLayers: 2
quantity: 0.001
maxExposurePosition: 0.05
enableDelayHedge: false
```

### 2. 激進套利交易
```yaml
margin: 0.2%
numLayers: 5
enableArbitrage: true
enableDelayHedge: true
maxHedgeDelayDuration: 10s
```

### 3. 信號驅動自適應交易
```yaml
enableSignalMargin: true
enableBollBandMargin: true
signals: [多個信號配置]
signalTrendSideMarginDiscount: [配置]
```

## 最佳實踐

1. **保守開始**：從更寬的保證金和較小的數量開始
2. **監控延遲**：確保交易所之間的低延遲
3. **餘額管理**：在兩個交易所維持足夠的餘額
4. **信號調優**：為您的市場仔細調整信號參數
5. **風險限制**：設置適當的風險敞口和虧損限制
6. **定期監控**：監控性能並調整參數

## 故障排除

### 常見問題

**訂單未成交**
- 檢查保證金設置（保證金過寬）
- 驗證市場流動性和競爭
- 檢查訂單放置時機

**對沖失敗**
- 檢查源交易所連接性
- 驗證對沖的可用餘額
- 檢查保證金水平要求

**信號問題**
- 驗證信號數據源
- 檢查信號計算參數
- 監控信號聚合權重

**性能問題**
- 分析價差捕獲與成本
- 檢查對沖時機和滑點
- 監控庫存周轉率

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [跨交易所交易指南](../../doc/topics/cross-exchange.md)
- [配置示例](../../../config/xmaker.yaml)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
