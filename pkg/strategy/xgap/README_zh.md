# XGap 策略（跨交易所價差交易）

## 概述

XGap策略是一個複雜的跨交易所套利策略，通過在交易交易所提供流動性同時監控源交易所的價格變動來利用不同交易所之間的價差。它充當做市商的角色，在中間價水平放置買賣訂單，通過買賣價差獲利，同時通過智能訂單放置和取消來管理倉位風險。

## 工作原理

1. **跨交易所監控**：監控源交易所和交易交易所的訂單簿
2. **價差檢測**：識別交易所間的價差和價差機會
3. **流動性提供**：在計算的中間價放置買賣訂單
4. **倉位管理**：當出現盈利機會時自動調整倉位
5. **風險控制**：實施價差閾值和倉位限制
6. **成交量模擬**：可基於源交易所活動模擬交易量

## 主要特性

- **跨交易所套利**：利用交易所間的價格差異
- **智能做市**：在最優中間價水平放置訂單
- **倉位調整**：自動平倉盈利倉位
- **價差管理**：可配置的最小價差要求
- **成交量模擬**：模仿源交易所的交易模式
- **風險控制**：多重保護措施，包括餘額檢查和塵埃數量過濾
- **靈活定價**：中間價或低於賣價的定價策略選項
- **製造價差功能**：當市場過於緊密時可主動創造價差

## 策略邏輯

### 跨交易所價格監控

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) CrossSubscribe(sessions map[string]*bbgo.ExchangeSession) {
    if len(s.SourceExchange) > 0 && len(s.SourceSymbol) > 0 {
        sourceSession, ok := sessions[s.SourceExchange]
        if !ok {
            panic(fmt.Errorf("source session %s is not defined", s.SourceExchange))
        }
        
        sourceSession.Subscribe(types.KLineChannel, s.SourceSymbol, types.SubscribeOptions{Interval: "1m"})
        sourceSession.Subscribe(types.BookChannel, s.SourceSymbol, types.SubscribeOptions{Depth: types.DepthLevelFull})
    }
    
    tradingSession.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: "1m"})
    tradingSession.Subscribe(types.BookChannel, s.Symbol, types.SubscribeOptions{Depth: types.DepthLevel5})
}
```
</augment_code_snippet>

### 倉位調整邏輯

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func buildAdjustPositionOrder(
    symbol string,
    positionSnapshot *types.Position,
    bestBid, bestAsk types.PriceVolume,
) (ok bool, order types.SubmitOrder) {
    if positionSnapshot.IsClosed() {
        return
    }
    
    var pv types.PriceVolume
    var side types.SideType
    if positionSnapshot.IsShort() && bestAsk.Price.Compare(positionSnapshot.AverageCost) < 0 {
        // 處於空頭倉位且最佳賣價低於平均成本
        pv = bestAsk
        side = types.SideTypeBuy
    } else if positionSnapshot.IsLong() && bestBid.Price.Compare(positionSnapshot.AverageCost) > 0 {
        // 處於多頭倉位且最佳買價高於平均成本
        pv = bestBid
        side = types.SideTypeSell
    }
    
    price := pv.Price
    quantity := fixedpoint.Min(positionSnapshot.Base.Abs(), pv.Volume)
    
    if !pv.IsZero() {
        order = types.SubmitOrder{
            Symbol:      symbol,
            Side:        side,
            Type:        types.OrderTypeLimit,
            Quantity:    quantity,
            Price:       price,
            TimeInForce: types.TimeInForceIOC,
        }
        ok = true
    }
    return
}
```
</augment_code_snippet>

### 訂單放置策略

<augment_code_snippet path="pkg/strategy/xgap/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) placeOrders(ctx context.Context) {
    bestBid, bestAsk, hasPrice := s.tradingBook.BestBidAndAsk()
    
    if hasPrice {
        var spread = bestAsk.Price.Sub(bestBid.Price)
        var spreadPercentage = spread.Div(bestAsk.Price)
        
        // 如果價差過大（>5%），使用源訂單簿價格
        if s.SimulatePrice && s.sourceBook != nil && spreadPercentage.Compare(maxStepPercentageGap) > 0 {
            bestBid, bestAsk, hasPrice = s.sourceBook.BestBidAndAsk()
        }
        
        // 檢查最小價差要求
        if s.MinSpread.Sign() > 0 && spreadPercentage.Compare(s.MinSpread) < 0 {
            if s.MakeSpread.Enabled {
                s.makeSpread(ctx, bestBid, bestAsk)
            }
            return
        }
    }
    
    var midPrice = bestAsk.Price.Add(bestBid.Price).Div(Two)
    var price fixedpoint.Value
    
    if s.SellBelowBestAsk {
        price = bestAsk.Price.Sub(s.tradingMarket.TickSize)
    } else {
        price = adjustPrice(midPrice, s.tradingMarket.PricePrecision)
    }
    
    // 在計算價格放置買賣訂單
    orderForms := []types.SubmitOrder{
        {
            Symbol:   s.Symbol,
            Side:     types.SideTypeBuy,
            Type:     types.OrderTypeLimit,
            Quantity: quantity,
            Price:    price,
        },
        {
            Symbol:      s.Symbol,
            Side:        types.SideTypeSell,
            Type:        types.OrderTypeLimit,
            Quantity:    quantity,
            Price:       price,
            TimeInForce: types.TimeInForceIOC,
        },
    }
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易交易所的交易對符號 |
| `tradingExchange` | 字符串 | 是 | 將要放置訂單的交易所 |
| `sourceExchange` | 字符串 | 是 | 監控價格參考的交易所 |
| `sourceSymbol` | 字符串 | 否 | 源交易所的符號（默認為symbol） |

### 價差和定價
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `minSpread` | 百分比 | 否 | 放置訂單所需的最小價差 |
| `sellBelowBestAsk` | 布爾值 | 否 | 在最佳賣價下1個tick放置賣單 |
| `simulatePrice` | 布爾值 | 否 | 當價差過大時使用源交易所價格 |

### 訂單管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quantity` | 數字 | 否 | 固定訂單數量（如未設置，使用計算數量） |
| `maxJitterQuantity` | 數字 | 否 | 隨機化的最大數量抖動 |
| `updateInterval` | 持續時間 | 否 | 訂單更新間隔（默認：1秒） |

### 製造價差功能
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `makeSpread.enabled` | 布爾值 | 否 | 啟用主動價差製造 |
| `makeSpread.skipLargeQuantityThreshold` | 數字 | 否 | 如果數量超過閾值則跳過製造價差 |

### 成交量控制
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `dailyMaxVolume` | 數字 | 否 | 每日最大交易量 |
| `dailyTargetVolume` | 數字 | 否 | 每日目標交易量 |
| `simulateVolume` | 布爾值 | 否 | 基於源交易所模擬成交量 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `dryRun` | 布爾值 | 否 | 啟用模擬運行模式（無實際訂單） |

## 配置示例

```yaml
crossExchangeStrategies:
- xgap:
    # 基本設置
    symbol: BTCUSDT
    tradingExchange: max          # 放置訂單的交易所
    sourceExchange: binance       # 監控價格的交易所
    sourceSymbol: BTCUSDT         # 源交易所的符號
    
    # 價差和定價
    minSpread: 0.001              # 0.1%最小價差
    sellBelowBestAsk: false       # 使用中間價策略
    simulatePrice: true           # 價差過大時使用源價格
    
    # 訂單管理
    quantity: 0.01                # 每訂單固定0.01 BTC
    maxJitterQuantity: 0.005      # 添加最多0.005 BTC抖動
    updateInterval: 30s           # 每30秒更新訂單
    
    # 製造價差功能
    makeSpread:
      enabled: true               # 啟用價差製造
      skipLargeQuantityThreshold: 1.0  # 數量>1.0 BTC時跳過
    
    # 成交量控制
    dailyMaxVolume: 10.0          # 每日最大10 BTC成交量
    simulateVolume: true          # 匹配源交易所成交量
    
    # 風險管理
    dryRun: false                 # 實盤交易模式
    
    # 費用預算（繼承自通用策略）
    dailyFeeBudgets:
      MAX: 100                    # 每日$100費用預算
```

## 策略工作流程

### 1. 市場數據收集
- 訂閱源交易所和交易交易所的訂單簿數據
- 監控1分鐘K線進行成交量分析
- 維護實時最佳買賣價格

### 2. 價差分析
- 計算交易交易所最佳買賣價間的價差
- 與最小價差要求比較
- 如果交易價差過大（>5%），回退到源交易所價格

### 3. 倉位評估
- 檢查當前倉位狀態
- 如果有盈利機會，嘗試放置倉位調整訂單
- 優先平倉盈利倉位而非開新倉位

### 4. 訂單放置
- 計算最優訂單價格（中間價或低於最佳賣價）
- 根據配置或成交量模擬確定訂單數量
- 同時放置買賣訂單
- 賣單使用IOC（立即成交或取消）防止自成交

### 5. 訂單管理
- 1秒後自動取消訂單
- 根據市場條件持續更新訂單
- 實施抖動時機避免可預測模式

## 定價策略

### 中間價策略（默認）
```
價格 = (最佳買價 + 最佳賣價) / 2
```
- 在價差中間放置訂單
- 最大化雙邊執行概率

### 低於最佳賣價策略
```
價格 = 最佳賣價 - 1個Tick
```
- 在最佳賣價下方放置訂單
- 更激進的定價以實現更快執行

## 成交量模擬

### 固定數量模式
- 對所有訂單使用配置的`quantity`
- 如果設置了`maxJitterQuantity`則添加隨機抖動

### 成交量模擬模式
- 監控源交易所和交易交易所間的成交量差異
- 根據成交量差距調整訂單數量
- 幫助在交易所間維持相似的交易活動

### 每日目標模式
- 計算數量以達到`dailyTargetVolume`
- 在交易間隔內平均分配成交量

## 風險管理功能

### 1. 價差控制
- **最小價差**：防止價差過緊時交易
- **最大價差**：價差過寬時回退到源定價
- **Tick大小驗證**：確保價差至少為2個tick寬

### 2. 倉位管理
- **自動調整**：立即平倉盈利倉位
- **餘額驗證**：放置訂單前確保餘額充足
- **塵埃數量過濾**：防止低於最小閾值的訂單

### 3. 訂單保護措施
- **防止自成交**：使用IOC訂單防止匹配自己的訂單
- **價格驗證**：確保訂單在買賣價範圍內
- **餘額檢查**：驗證每個訂單的充足資金

### 4. 製造價差功能
- **流動性注入**：當市場過於緊密時主動創造價差
- **數量限制**：跳過大數量以避免市場影響
- **餘額保護**：製造價差前確保餘額充足

## 性能優化

### 時機和抖動
- 使用抖動更新間隔避免可預測模式
- 實施隨機延遲減少市場影響
- 平衡響應性與穩定性

### 訂單生命週期
- 短期訂單（1秒）最小化暴露
- 立即取消和替換以獲得新鮮定價
- 賣方使用IOC訂單防止累積

## 常見用例

### 1. 保守套利
```yaml
minSpread: 0.002              # 0.2%最小價差
quantity: 0.001               # 小固定數量
updateInterval: 60s           # 較慢更新
makeSpread:
  enabled: false              # 無主動價差製造
```

### 2. 激進做市
```yaml
minSpread: 0.0005             # 0.05%最小價差
simulateVolume: true          # 匹配源成交量
updateInterval: 10s           # 更快更新
makeSpread:
  enabled: true               # 主動價差製造
```

### 3. 成交量匹配
```yaml
simulateVolume: true          # 匹配源交易所成交量
dailyTargetVolume: 50.0       # 目標每日50 BTC
maxJitterQuantity: 0.01       # 添加數量隨機化
```

## 最佳實踐

1. **交易所選擇**：選擇API可靠性好且延遲低的交易所
2. **價差調整**：根據典型市場條件設置最小價差
3. **成交量管理**：使用每日限制控制暴露
4. **費用監控**：設置適當的費用預算維持盈利能力
5. **風險限制**：從小數量開始逐漸增加
6. **市場時間**：考慮交易所間不同的交易時間

## 限制

1. **延遲敏感性**：性能取決於交易所間的網絡延遲
2. **API速率限制**：受交易所API速率限制約束
3. **市場條件**：在高波動或流動性不足的市場中效果較差
4. **交易所風險**：暴露於交易所特定風險和停機時間
5. **監管合規**：必須遵守兩個交易所的法規

## 故障排除

### 常見問題

**未放置訂單**
- 檢查最小價差要求
- 驗證交易交易所餘額充足
- 確保價差至少為2個tick寬

**頻繁訂單取消**
- 根據市場條件調整更新間隔
- 檢查價差是否滿足最小要求
- 驗證兩個交易所的網絡連接

**倉位累積**
- 啟用倉位調整功能
- 檢查調整訂單是否成交
- 檢查價差閾值和市場條件

**盈利能力低**
- 增加最小價差要求
- 優化訂單數量
- 檢查兩個交易所的費用結構

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/xgap.yaml)
- [跨交易所交易指南](../../doc/topics/cross-exchange.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
