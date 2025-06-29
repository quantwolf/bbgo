# DCA2 策略（定投策略 v2）

## 概述

DCA2策略是一個高級定投實現，自動在遞減價格水平下單多個買單，並在平均成本達到目標利潤比率時止盈。它專為在市場下跌期間積累資產而設計，通過系統性的倉位建立和止盈來維持嚴格的風險管理。

## 工作原理

1. **多級買單**：基於價格偏差在遞減價格水平放置多個買單
2. **等額名義分配**：將報價投資平均分配到所有買單
3. **自動止盈**：基於平均成本和利潤比率計算止盈價格
4. **狀態機管理**：使用複雜的狀態機管理訂單生命週期
5. **輪次交易**：以輪次方式運行，輪次間有冷卻期
6. **利潤再投資**：自動將利潤再投資到下一輪

## 主要特性

- **系統性DCA方法**：在當前市場價格下方的預定價格水平下單
- **狀態機架構**：強大的訂單生命週期狀態管理
- **自動恢復**：可從中斷中恢復並繼續交易
- **利潤跟踪**：基於輪次跟踪的全面利潤統計
- **風險管理**：內置保護措施和驗證機制
- **靈活配置**：針對不同市場條件的高度可配置參數
- **訂單組管理**：使用訂單組進行高效訂單管理
- **冷卻機制**：通過可配置的冷卻期防止過度交易

## 策略邏輯

### 狀態機流程

<augment_code_snippet path="pkg/strategy/dca2/state.go" mode="EXCERPT">
```go
type State int64

const (
    None State = iota
    IdleWaiting             // 空閒等待
    OpenPositionReady       // 開倉準備
    OpenPositionOrderFilled // 開倉訂單成交
    OpenPositionFinished    // 開倉完成
    TakeProfitReady         // 止盈準備
)

var stateTransition map[State]State = map[State]State{
    IdleWaiting:             OpenPositionReady,
    OpenPositionReady:       OpenPositionOrderFilled,
    OpenPositionOrderFilled: OpenPositionFinished,
    OpenPositionFinished:    TakeProfitReady,
    TakeProfitReady:         IdleWaiting,
}
```
</augment_code_snippet>

### 開倉訂單生成

<augment_code_snippet path="pkg/strategy/dca2/open_position.go" mode="EXCERPT">
```go
func generateOpenPositionOrders(market types.Market, enableQuoteInvestmentReallocate bool, quoteInvestment, profit, price, priceDeviation fixedpoint.Value, maxOrderCount int64, orderGroupID uint32) ([]types.SubmitOrder, error) {
    factor := fixedpoint.One.Sub(priceDeviation)
    
    // 計算所有有效價格
    var prices []fixedpoint.Value
    for i := 0; i < int(maxOrderCount); i++ {
        if i > 0 {
            price = price.Mul(factor)
        }
        price = market.TruncatePrice(price)
        if price.Compare(market.MinPrice) < 0 {
            break
        }
        prices = append(prices, price)
    }
    
    notional, orderNum := calculateNotionalAndNumOrders(market, quoteInvestment, prices)
    
    // 生成等額名義的提交訂單
    var submitOrders []types.SubmitOrder
    for i := 0; i < orderNum; i++ {
        var quantity fixedpoint.Value
        if i == 0 {
            // 第一個訂單包含累積利潤
            quantity = market.TruncateQuantity(notional.Add(profit).Div(prices[i]))
        } else {
            quantity = market.TruncateQuantity(notional.Div(prices[i]))
        }
        submitOrders = append(submitOrders, types.SubmitOrder{
            Symbol:      market.Symbol,
            Type:        types.OrderTypeLimit,
            Price:       prices[i],
            Side:        types.SideTypeBuy,
            Quantity:    quantity,
            GroupID:     orderGroupID,
        })
    }
    
    return submitOrders, nil
}
```
</augment_code_snippet>

### 止盈計算

<augment_code_snippet path="pkg/strategy/dca2/take_profit.go" mode="EXCERPT">
```go
func generateTakeProfitOrder(market types.Market, takeProfitRatio fixedpoint.Value, position *types.Position, orderGroupID uint32) types.SubmitOrder {
    takeProfitPrice := market.TruncatePrice(position.AverageCost.Mul(fixedpoint.One.Add(takeProfitRatio)))
    return types.SubmitOrder{
        Symbol:      market.Symbol,
        Type:        types.OrderTypeLimit,
        Price:       takeProfitPrice,
        Side:        types.SideTypeSell,
        Quantity:    position.GetBase().Abs(),
        GroupID:     orderGroupID,
    }
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `quoteInvestment` | 數字 | 是 | 每輪投資的總報價貨幣 |
| `maxOrderCount` | 整數 | 是 | 要放置的最大買單數量 |
| `priceDeviation` | 百分比 | 是 | 連續訂單間的價格偏差 |
| `takeProfitRatio` | 百分比 | 是 | 止盈計算的利潤比率 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `coolDownInterval` | 持續時間 | 是 | 輪次間的冷卻期（秒） |
| `orderGroupID` | 整數 | 否 | 訂單管理的自定義訂單組ID |
| `disableOrderGroupIDFilter` | 布爾值 | 否 | 禁用訂單組ID過濾 |

### 恢復和持久化
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `recoverWhenStart` | 布爾值 | 否 | 策略啟動時啟用恢復 |
| `disableProfitStatsRecover` | 布爾值 | 否 | 禁用利潤統計恢復 |
| `disablePositionRecover` | 布爾值 | 否 | 禁用倉位恢復 |
| `persistenceTTL` | 持續時間 | 否 | 持久化數據的生存時間 |

### 高級選項
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `enableQuoteInvestmentReallocate` | 布爾值 | 否 | 當訂單低於最小值時允許重新分配 |
| `keepOrdersWhenShutdown` | 布爾值 | 否 | 關閉時保持訂單活躍 |
| `useCancelAllOrdersApiWhenClose` | 布爾值 | 否 | 關閉時使用取消所有訂單API |

### 監控和調試
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `logFields` | 對象 | 否 | 調試用的額外日誌字段 |
| `prometheusLabels` | 對象 | 否 | 指標的自定義Prometheus標籤 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  dca2:
    symbol: BTCUSDT
    
    # 投資參數
    quoteInvestment: 1000        # 每輪投資$1000
    maxOrderCount: 5             # 放置5個買單
    priceDeviation: 0.02         # 訂單間2%價格偏差
    takeProfitRatio: 0.005       # 0.5%止盈比率
    
    # 風險管理
    coolDownInterval: 3600       # 輪次間1小時冷卻
    
    # 恢復設置
    recoverWhenStart: true       # 啟動時啟用恢復
    disableProfitStatsRecover: false
    disablePositionRecover: false
    
    # 訂單管理
    keepOrdersWhenShutdown: false
    useCancelAllOrdersApiWhenClose: false
    enableQuoteInvestmentReallocate: true
    
    # 持久化
    persistenceTTL: 24h          # 保持數據24小時
    
    # 監控
    logFields:
      environment: "production"
    prometheusLabels:
      instance: "dca-btc-001"
```

## 策略工作流程

### 1. 輪次初始化（空閒等待 → 開倉準備）
- 等待冷卻期完成
- 驗證沒有現有的開倉訂單
- 基於當前市場價格和價格偏差計算訂單價格
- 在遞減價格水平放置多個買單

### 2. 倉位建立（開倉準備 → 開倉訂單成交）
- 監控買單成交
- 隨著訂單成交更新倉位
- 計算運行平均成本

### 3. 倉位完成（開倉訂單成交 → 開倉完成）
- 當價格達到止盈水平或所有訂單成交時觸發
- 轉換到止盈階段

### 4. 止盈（開倉完成 → 止盈準備）
- 取消剩餘買單
- 在計算價格放置止盈賣單
- 止盈價格 = 平均成本 × (1 + 止盈比率)

### 5. 輪次完成（止盈準備 → 空閒等待）
- 監控止盈訂單執行
- 更新利潤統計
- 為下一輪重置倉位
- 開始冷卻期

## 訂單放置邏輯

### 價格計算
```
訂單1：當前價格
訂單2：當前價格 × (1 - 價格偏差)
訂單3：當前價格 × (1 - 價格偏差)²
訂單4：當前價格 × (1 - 價格偏差)³
訂單5：當前價格 × (1 - 價格偏差)⁴
```

### 數量計算
```
每訂單名義 = 報價投資 ÷ 訂單數量
數量 = 名義 ÷ 訂單價格

第一個訂單的特殊情況：
第一訂單數量 = (名義 + 累積利潤) ÷ 第一訂單價格
```

## 風險管理功能

### 1. 等額名義分配
- 每個訂單具有相同的名義價值
- 確保平衡的倉位建立
- 防止在任何價格水平過度集中

### 2. 最小要求驗證
- 驗證訂單滿足交易所最小名義要求
- 需要時自動調整訂單數量（啟用重新分配時）
- 防止訂單提交失敗

### 3. 餘額驗證
- 放置止盈訂單前驗證充足餘額
- 防止過度賣出情況
- 警告餘額差異

### 4. 狀態機保護措施
- 驗證狀態轉換
- 防止無效操作
- 確保適當的訂單生命週期管理

## 性能指標

### 利潤統計
- **總利潤**：所有輪次的累積利潤
- **當前輪次利潤**：當前輪次的利潤
- **報價投資**：包括再投資利潤的總投資
- **輪次計數**：完成的輪次數量

### 訂單指標
- **活躍訂單**：當前活躍訂單數量
- **成交率**：訂單成交百分比
- **平均成交價格**：成交量加權平均成交價格

## 常見用例

### 1. 保守DCA設置
```yaml
quoteInvestment: 500
maxOrderCount: 3
priceDeviation: 0.05         # 訂單間5%
takeProfitRatio: 0.01        # 1%止盈
coolDownInterval: 86400      # 24小時冷卻
```

### 2. 激進DCA設置
```yaml
quoteInvestment: 1000
maxOrderCount: 7
priceDeviation: 0.015        # 訂單間1.5%
takeProfitRatio: 0.003       # 0.3%止盈
coolDownInterval: 3600       # 1小時冷卻
```

### 3. 保守長期設置
```yaml
quoteInvestment: 2000
maxOrderCount: 4
priceDeviation: 0.08         # 訂單間8%
takeProfitRatio: 0.02        # 2%止盈
coolDownInterval: 604800     # 1週冷卻
```

## 最佳實踐

1. **價格偏差調整**：根據資產波動性設置價格偏差
2. **止盈優化**：在頻繁利潤和利潤規模間平衡
3. **冷卻管理**：在趨勢市場中使用更長冷卻期
4. **投資規模**：根據可用資本和風險承受能力確定投資規模
5. **恢復設置**：在生產環境中啟用恢復
6. **監控**：使用Prometheus標籤進行全面監控

## 限制

1. **下跌趋势依賴性**：在區間或輕微下跌市場中效果最佳
2. **資本要求**：需要充足資本進行多個訂單
3. **交易所限制**：受交易所最小名義和數量限制約束
4. **市場影響**：大訂單可能影響市場價格
5. **時機風險**：在快速價格變動期間可能錯過機會

## 故障排除

### 常見問題

**訂單未放置**
- 檢查報價投資是否滿足最小要求
- 驗證價格偏差不會創建低於最小值的價格
- 確保賬戶餘額充足

**止盈未觸發**
- 驗證止盈比率對市場條件合理
- 檢查價格是否達到計算的止盈水平
- 確保止盈訂單成功放置

**恢復失敗**
- 檢查持久化配置
- 驗證訂單組ID一致性
- 查看日誌文件獲取具體錯誤消息

**狀態機卡住**
- 監控日誌中的狀態轉換
- 檢查網絡連接問題
- 驗證交易所API響應

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/dca2.yaml)
- [原始DCA策略](../dca/README.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
