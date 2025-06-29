# Audacity Maker 策略

## 概述

Audacity Maker策略是一個基於訂單流的高級交易策略，分析實時市場微觀結構以識別交易機會。它使用買方發起與賣方發起的交易量和交易次數的統計分析來檢測市場失衡並進行方向性交易。該策略採用異常值檢測技術識別訂單流模式中的顯著偏差，在捕捉短期動量轉換方面特別有效。

## 工作原理

1. **訂單流分析**：持續監控買方發起和賣方發起的交易
2. **成交量和次數跟踪**：跟踪每一方的交易量和交易次數
3. **統計標準化**：對滾動窗口內的訂單流比率應用最小-最大縮放
4. **異常值檢測**：使用基於標準差的異常值檢測識別顯著失衡
5. **方向性交易**：當檢測到顯著訂單流失衡時在最佳買賣價放置限價單
6. **倉位管理**：與退出方法整合進行全面風險管理

## 主要特性

- **實時訂單流分析**：處理每筆市場交易以構建訂單流指標
- **雙指標系統**：分析基於成交量和基於次數的訂單流
- **統計異常值檢測**：使用3-sigma閾值進行信號生成
- **最小-最大標準化**：將訂單流比率縮放到0-1範圍以進行一致分析
- **限價做市訂單**：在最佳買賣價放置訂單以捕獲價差
- **滾動窗口分析**：使用100週期和200週期滾動窗口
- **市場微觀結構焦點**：利用短期市場低效性
- **退出方法整合**：與各種風險管理策略兼容

## 策略邏輯

### 訂單流計算

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// 跟踪買方和賣方發起的交易
if trade.Side == types.SideTypeBuy {
    // 買方發起交易（激進買入）
    buyTradeSize.Update(trade.Quantity.Float64())
    sellTradeSize.Update(0)
    buyTradesNumber.Update(1)
    sellTradesNumber.Update(0)
} else if trade.Side == types.SideTypeSell {
    // 賣方發起交易（激進賣出）
    buyTradeSize.Update(0)
    sellTradeSize.Update(trade.Quantity.Float64())
    buyTradesNumber.Update(0)
    sellTradesNumber.Update(1)
}

// 計算訂單流比率
sizeFraction := buyTradeSize.Sum() / sellTradeSize.Sum()
numberFraction := buyTradesNumber.Sum() / sellTradesNumber.Sum()
```
</augment_code_snippet>

### 統計標準化

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// 訂單流規模的最小-最大縮放
if orderFlowSize.Length() > 100 {
    ofsMax := orderFlowSize.Tail(100).Max()
    ofsMin := orderFlowSize.Tail(100).Min()
    ofsMinMax := (orderFlowSize.Last(0) - ofsMin) / (ofsMax - ofsMin)
    orderFlowSizeMinMax.Push(ofsMinMax)
}

// 訂單流次數的最小-最大縮放
if orderFlowNumber.Length() > 100 {
    ofnMax := orderFlowNumber.Tail(100).Max()
    ofnMin := orderFlowNumber.Tail(100).Min()
    ofnMinMax := (orderFlowNumber.Last(0) - ofnMin) / (ofnMax - ofnMin)
    orderFlowNumberMinMax.Push(ofnMinMax)
}
```
</augment_code_snippet>

### 異常值檢測和交易邏輯

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
// 使用3-sigma閾值檢測異常值
func outlier(fs floats.Slice, multiplier float64) int {
    stddev := stat.StdDev(fs, nil)
    if fs.Last(0) > fs.Mean()+multiplier*stddev {
        return 1  // 正異常值
    } else if fs.Last(0) < fs.Mean()-multiplier*stddev {
        return -1 // 負異常值
    }
    return 0 // 無異常值
}

// 基於雙重異常值確認的交易決策
if outlier(orderFlowSizeMinMax.Tail(100), threshold) > 0 && 
   outlier(orderFlowNumberMinMax.Tail(100), threshold) > 0 {
    // 檢測到強烈買入壓力
    _ = s.placeOrder(ctx, types.SideTypeBuy, s.Quantity, bid.Price, symbol)
} else if outlier(orderFlowSizeMinMax.Tail(100), threshold) < 0 && 
          outlier(orderFlowNumberMinMax.Tail(100), threshold) < 0 {
    // 檢測到強烈賣出壓力
    _ = s.placeOrder(ctx, types.SideTypeSell, s.Quantity, ask.Price, symbol)
}
```
</augment_code_snippet>

### 訂單放置

<augment_code_snippet path="pkg/strategy/audacitymaker/orderflow.go" mode="EXCERPT">
```go
func (s *PerTrade) placeOrder(
    ctx context.Context, side types.SideType, quantity fixedpoint.Value, price fixedpoint.Value, symbol string,
) error {
    market, _ := s.session.Market(symbol)
    _, err := s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
        Symbol:   symbol,
        Market:   market,
        Side:     side,
        Type:     types.OrderTypeLimitMaker,  // 做市訂單以捕獲價差
        Quantity: quantity,
        Price:    price,
        Tag:      "audacity-limit",
    })
    return err
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："ETHBUSD"） |
| `orderFlow.interval` | 間隔 | 是 | 監控的K線間隔（例如："1m"） |
| `orderFlow.quantity` | 數字 | 是 | 每筆交易的固定數量 |

### 高級設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `orderFlow.marketOrder` | 布爾值 | 否 | 啟用市價單執行（默認：false） |

### 退出方法
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `exits` | 數組 | 否 | 退出方法配置數組 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  audacitymaker:
    symbol: ETHBUSD
    
    # 訂單流配置
    orderFlow:
      interval: 1m            # 監控1分鐘間隔
      quantity: 0.01          # 每筆交易0.01 ETH
      marketOrder: false      # 使用限價做市訂單
    
    # 風險管理（可選）
    exits:
    - roiStopLoss:
        percentage: 0.02      # 2%止損
    - roiTakeProfit:
        percentage: 0.01      # 1%止盈
    - trailingStop:
        callbackRate: 0.005   # 0.5%追蹤止損
```

## 訂單流分析解釋

### 什麼是訂單流？
訂單流代表買方發起和賣方發起交易活動之間的差異。它提供以下洞察：
- **市場情緒**：買方或賣方是否更激進
- **流動性動態**：市場參與者如何與訂單簿互動
- **短期動量**：市場中的即時方向性壓力

### 雙指標方法
策略分析兩個互補指標：

1. **基於成交量的訂單流**：`買入量 / 賣出量`
   - 測量激進交易的規模
   - 表明機構或大型交易者活動
   - 對重大市場變動更敏感

2. **基於次數的訂單流**：`買入次數 / 賣出次數`
   - 測量激進交易的頻率
   - 表明零售或算法活動
   - 對持續方向性壓力更敏感

### 統計處理管道

#### 1. 原始數據收集
- 實時跟踪買方和賣方發起的交易
- 使用200週期滾動隊列處理成交量和交易次數
- 為買賣活動維護獨立隊列

#### 2. 比率計算
```
規模比率 = Sum(買入量) / Sum(賣出量)
次數比率 = Sum(買入次數) / Sum(賣出次數)
```

#### 3. 最小-最大標準化
```
標準化值 = (當前值 - 最小值) / (最大值 - 最小值)
```
- 在100週期窗口內將比率縮放到0-1範圍
- 實現不同市場條件下的一致比較
- 保持數據中的時間關係

#### 4. 異常值檢測
```
異常值閾值 = 均值 ± (3 × 標準差)
```
- 使用3-sigma規則進行異常值識別
- 正異常值表明強烈買入壓力
- 負異常值表明強烈賣出壓力

## 交易邏輯

### 信號生成
策略需要交易信號的**雙重確認**：
- 基於成交量和基於次數的指標都必須顯示異常值
- 相同方向的異常值確認信號強度
- 相反方向的異常值被忽略（不交易）

### 入場條件

**多頭入場：**
1. 基於成交量的訂單流顯示正異常值（> 均值 + 3σ）
2. 基於次數的訂單流顯示正異常值（> 均值 + 3σ）
3. 在當前最佳買價放置限價買單

**空頭入場：**
1. 基於成交量的訂單流顯示負異常值（< 均值 - 3σ）
2. 基於次數的訂單流顯示負異常值（< 均值 - 3σ）
3. 在當前最佳賣價放置限價賣單

### 訂單管理
- **訂單類型**：限價做市訂單以捕獲價差
- **價格**：買單使用最佳買價，賣單使用最佳賣價
- **取消**：放置新訂單前取消現有訂單
- **數量**：根據配置的固定數量

## 性能特徵

### 市場條件
- **高頻市場**：在交易頻繁的市場中表現出色
- **流動性市場**：需要充足的交易流進行分析
- **波動性市場**：受益於明確的方向性變動
- **趨勢市場**：能有效捕捉動量轉換

### 時間敏感性
- **實時處理**：立即處理每筆市場交易
- **短期焦點**：設計用於捕捉快速動量轉換
- **剝頭皮性質**：通常短期持有倉位

### 統計要求
- **最小數據**：需要100+筆交易才能產生可靠信號
- **滾動窗口**：使用100-200週期窗口保持穩定性
- **異常值閾值**：3-sigma閾值平衡敏感性和噪音

## 風險管理

### 內置保護
- **雙重確認**：要求兩個指標都同意
- **限價訂單**：使用做市訂單避免即時市場影響
- **訂單取消**：新交易前取消衝突訂單

### 推薦風險控制
- **倉位規模**：使用小的固定數量
- **止損**：實施緊密止損（1-2%）
- **止盈**：快速止盈（0.5-1%）
- **時間限制**：考慮對陳舊倉位進行基於時間的退出

## 常見用例

### 1. 高頻剝頭皮
```yaml
orderFlow:
  interval: 1m
  quantity: 0.001
exits:
- roiTakeProfit:
    percentage: 0.005  # 0.5%快速利潤
```

### 2. 動量捕捉
```yaml
orderFlow:
  interval: 1m
  quantity: 0.01
exits:
- roiTakeProfit:
    percentage: 0.01   # 1%動量利潤
- roiStopLoss:
    percentage: 0.02   # 2%風險控制
```

### 3. 保守訂單流
```yaml
orderFlow:
  interval: 5m
  quantity: 0.005
exits:
- trailingStop:
    callbackRate: 0.01 # 1%追蹤止損
```

## 最佳實踐

1. **市場選擇**：選擇交易頻繁的高流動性市場
2. **數量規模**：從小數量開始測試有效性
3. **風險管理**：始終使用退出方法進行保護
4. **監控**：關注市場微觀結構的變化
5. **回測**：實盤交易前在歷史數據上充分測試
6. **參數調整**：考慮為不同市場調整異常值閾值

## 限制

1. **數據依賴性**：需要高頻交易數據
2. **市場結構**：性能隨市場微觀結構變化而變化
3. **延遲敏感性**：有效性取決於低延遲執行
4. **虛假信號**：在震盪市場中可能產生虛假信號
5. **交易成本**：高頻性質可能產生大量費用
6. **市場影響**：大數量可能影響正在利用的模式

## 故障排除

### 常見問題

**無交易信號**
- 檢查市場是否有足夠的交易頻率
- 驗證訂單流數據是否正在收集
- 確保已發生100+筆交易進行分析

**過多虛假信號**
- 考慮增加異常值閾值（> 3 sigma）
- 添加額外確認過濾器
- 實施最小持倉期

**成交率差**
- 檢查限價訂單是否具有競爭力
- 考慮對關鍵信號使用市價單
- 驗證訂單簿深度和價差

**高交易成本**
- 通過更高閾值減少交易頻率
- 優化數量規模
- 考慮做市商回扣計劃

## 高級概念

### 市場微觀結構理論
策略基於以下原則：
- 激進交易揭示私人信息
- 訂單流失衡預測短期價格變動
- 統計異常值表明重大市場事件

### 統計基礎
- **中心極限定理**：假設訂單流比率遵循正態分佈
- **異常值檢測**：使用標準差識別罕見事件
- **最小-最大縮放**：標準化數據以在時間上進行一致分析

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/audacitymaker.yaml)
- [訂單流分析指南](../../doc/topics/order-flow.md)
- [市場微觀結構交易](../../doc/topics/microstructure.md)
