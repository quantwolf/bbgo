# 三角套利策略

## 概述

三角套利（tri）策略是一個複雜的套利策略，利用同一交易所內三個相關交易對之間的價格低效性。它自動檢測有利可圖的三角套利機會，並執行三個連續交易，從貨幣對之間的臨時價格差異中獲取無風險利潤。

## 工作原理

1. **路徑檢測**：識別三個相關貨幣對之間的三角路徑
2. **比率計算**：持續計算正向和反向套利比率
3. **機會識別**：檢測套利比率何時超過最小利潤閾值
4. **順序執行**：使用IOC（立即成交或取消）訂單順序執行三筆交易
5. **倉位管理**：跟踪多貨幣倉位並計算利潤
6. **風險控制**：實施餘額限制和保護性訂單定價

## 主要特性

- **自動路徑發現**：自動確定三角路徑的交易方向
- **實時監控**：持續監控訂單簿以尋找套利機會
- **IOC訂單策略**：使用立即成交或取消訂單最小化執行風險
- **多貨幣倉位跟踪**：跟踪多種貨幣的倉位
- **保護性定價**：對市價單應用保護比率以防止滑點
- **餘額管理**：實施餘額限制和緩衝區進行風險控制
- **性能分析**：跟踪IOC勝率和交易統計
- **獨立流**：可選的獨立WebSocket流以獲得更好性能

## 策略邏輯

### 三角路徑結構

<augment_code_snippet path="pkg/strategy/tri/path.go" mode="EXCERPT">
```go
type Path struct {
    marketA, marketB, marketC *ArbMarket
    dirA, dirB, dirC          int
}

func (p *Path) solveDirection() error {
    // 根據貨幣關係自動確定交易方向
    if p.marketA.QuoteCurrency == p.marketB.BaseCurrency || p.marketA.QuoteCurrency == p.marketB.QuoteCurrency {
        p.dirA = 1  // 賣出方向
    } else if p.marketA.BaseCurrency == p.marketB.BaseCurrency || p.marketA.BaseCurrency == p.marketB.QuoteCurrency {
        p.dirA = -1 // 買入方向
    }
    // marketB和marketC的類似邏輯...
}
```
</augment_code_snippet>

### 套利比率計算

<augment_code_snippet path="pkg/strategy/tri/strategy.go" mode="EXCERPT">
```go
// 正向套利：A -> B -> C -> A
func calculateForwardRatio(p *Path) float64 {
    var ratio = 1.0
    ratio *= p.marketA.calculateRatio(p.dirA)
    ratio *= p.marketB.calculateRatio(p.dirB)
    ratio *= p.marketC.calculateRatio(p.dirC)
    return ratio
}

// 反向套利：A <- B <- C <- A
func calculateBackwardRate(p *Path) float64 {
    var ratio = 1.0
    ratio *= p.marketA.calculateRatio(-p.dirA)
    ratio *= p.marketB.calculateRatio(-p.dirB)
    ratio *= p.marketC.calculateRatio(-p.dirC)
    return ratio
}
```
</augment_code_snippet>

### IOC訂單執行

<augment_code_snippet path="pkg/strategy/tri/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) iocOrderExecution(
    ctx context.Context, session *bbgo.ExchangeSession, orders [3]types.SubmitOrder, ratio float64,
) (types.OrderSlice, error) {
    // 將第一個訂單執行為IOC
    orders[0].Type = types.OrderTypeLimit
    orders[0].TimeInForce = types.TimeInForceIOC
    
    iocOrder := s.executeOrder(ctx, orders[0])
    if iocOrder == nil {
        return nil, errors.New("ioc order submit error")
    }
    
    // 等待IOC訂單完成
    o := <-iocOrderC
    filledQuantity := o.ExecutedQuantity
    
    if filledQuantity.IsZero() {
        s.State.IOCLossTimes++
        return nil, nil
    }
    
    // 根據成交數量調整後續訂單
    filledRatio := filledQuantity.Div(iocOrder.Quantity)
    orders[1].Quantity = orders[1].Quantity.Mul(filledRatio)
    orders[2].Quantity = orders[2].Quantity.Mul(filledRatio)
    
    // 以保護性定價執行剩餘訂單作為市價單
    orders[1] = s.toProtectiveMarketOrder(orders[1], s.MarketOrderProtectiveRatio)
    orders[2] = s.toProtectiveMarketOrder(orders[2], s.MarketOrderProtectiveRatio)
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbols` | 數組 | 否 | 交易符號列表（如未指定則從路徑自動檢測） |
| `paths` | 數組 | 是 | 三角路徑數組，每個包含3個符號 |
| `minSpreadRatio` | 數字 | 否 | 所需的最小利潤比率（默認：1.002 = 0.2%） |

### 執行設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `separateStream` | 布爾值 | 否 | 為每個符號使用獨立的WebSocket流 |
| `marketOrderProtectiveRatio` | 數字 | 否 | 市價單的保護比率（默認：0.008） |
| `iocOrderRatio` | 數字 | 否 | IOC訂單的保護比率 |
| `coolingDownTime` | 持續時間 | 否 | 套利執行間的冷卻期 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `limits` | 對象 | 否 | 每種貨幣的最大餘額限制 |
| `resetPosition` | 布爾值 | 否 | 策略啟動時重置倉位跟踪 |
| `dryRun` | 布爾值 | 否 | 啟用模擬運行模式（無實際訂單） |

### 監控
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `notifyTrade` | 布爾值 | 否 | 為執行的交易發送通知 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  tri:
    # 最小利潤閾值
    minSpreadRatio: 1.0015        # 0.15%最小利潤
    
    # 性能優化
    separateStream: true          # 使用獨立流獲得更好性能
    
    # 風險管理
    limits:
      BTC: 0.01                   # 最大0.01 BTC暴露
      ETH: 0.1                    # 最大0.1 ETH暴露
      USDT: 1000.0                # 最大$1000 USDT暴露
    
    # 執行設置
    marketOrderProtectiveRatio: 0.008  # 0.8%保護比率
    iocOrderRatio: 0.005              # 0.5% IOC保護比率
    coolingDownTime: 1s               # 1秒冷卻
    
    # 監控
    notifyTrade: true             # 啟用交易通知
    resetPosition: false          # 重啟時不重置倉位
    dryRun: false                 # 實盤交易模式
    
    # 交易符號（從路徑自動檢測）
    symbols:
    - BTCUSDT
    - ETHUSDT
    - ETHBTC
    - BNBUSDT
    - BNBBTC
    - BNBETH
    
    # 三角套利路徑
    paths:
    - [BTCUSDT, ETHBTC, ETHUSDT]    # BTC -> ETH -> USDT -> BTC
    - [BNBBTC, BNBUSDT, BTCUSDT]    # BNB -> BTC -> USDT -> BNB
    - [BNBETH, BNBUSDT, ETHUSDT]    # BNB -> ETH -> USDT -> BNB
```

## 三角套利路徑

### 路徑結構
每個路徑由三個形成閉環的交易對組成：
```
路徑：[BTCUSDT, ETHBTC, ETHUSDT]
方向：BTC -> ETH -> USDT -> BTC
```

### 自動方向檢測
策略自動確定交易方向：
- **正向路徑**：A → B → C → A
- **反向路徑**：A ← B ← C ← A

### 計算示例

**正向套利示例：**
```
起始：1 BTC
1. BTCUSDT：賣出1 BTC → 獲得50,000 USDT
2. ETHUSDT：用50,000 USDT買入ETH → 獲得20 ETH  
3. ETHBTC：賣出20 ETH → 獲得1.002 BTC
利潤：0.002 BTC（0.2%）
```

**反向套利示例：**
```
起始：1 BTC
1. ETHBTC：用1 BTC買入20 ETH
2. ETHUSDT：賣出20 ETH → 獲得50,100 USDT
3. BTCUSDT：用50,100 USDT買入BTC → 獲得1.002 BTC
利潤：0.002 BTC（0.2%）
```

## 執行工作流程

### 1. 市場數據監控
- 訂閱路徑中所有符號的訂單簿數據
- 持續計算最佳買賣價格
- 實時更新套利比率

### 2. 機會檢測
- 計算正向和反向套利比率
- 將比率與最小價差閾值比較
- 按盈利能力對機會進行排名

### 3. 訂單執行序列
1. **IOC訂單**：將第一個訂單作為立即成交或取消下單
2. **成交驗證**：等待IOC訂單完成
3. **數量調整**：根據成交數量調整剩餘訂單
4. **市價單**：以保護性市價單執行剩餘訂單
5. **交易收集**：收集和分析所有交易

### 4. 倉位和利潤跟踪
- 更新多貨幣倉位跟踪
- 計算美元等值利潤
- 更新性能統計

## 風險管理功能

### 1. 餘額控制
- **餘額限制**：每種貨幣的最大暴露
- **餘額緩衝**：保留小緩衝區防止過度交易
- **最小數量**：驗證訂單滿足交易所最小值

### 2. 保護性定價
- **市價單保護**：應用保護比率防止滑點
- **IOC訂單保護**：IOC訂單的可選保護定價
- **價格驗證**：確保訂單在合理價格範圍內

### 3. 執行保護措施
- **IOC策略**：使用IOC訂單最小化執行風險
- **數量調整**：根據實際成交調整後續訂單
- **訂單驗證**：提交前驗證所有訂單

### 4. 性能監控
- **IOC勝率**：跟踪IOC訂單成功率
- **交易統計**：全面的交易性能分析
- **倉位跟踪**：多貨幣倉位管理

## 性能優化

### 獨立流
```yaml
separateStream: true
```
- 為每個符號創建專用WebSocket連接
- 減少延遲並改善數據新鮮度
- 高頻套利的更好性能

### 冷卻
```yaml
coolingDownTime: 1s
```
- 防止過度交易和交易所速率限制
- 允許倉位在套利週期間結算
- 減少市場影響

## 常見用例

### 1. 保守套利
```yaml
minSpreadRatio: 1.005           # 0.5%最小利潤
coolingDownTime: 5s             # 5秒冷卻
limits:
  BTC: 0.001                    # 小倉位限制
  ETH: 0.01
  USDT: 100.0
```

### 2. 激進套利
```yaml
minSpreadRatio: 1.001           # 0.1%最小利潤
coolingDownTime: 1s             # 1秒冷卻
separateStream: true            # 優化速度
limits:
  BTC: 0.01                     # 較大倉位限制
  ETH: 0.1
  USDT: 1000.0
```

### 3. 高成交量套利
```yaml
minSpreadRatio: 1.002           # 0.2%最小利潤
separateStream: true            # 最大性能
marketOrderProtectiveRatio: 0.005  # 更緊的保護比率
limits:
  BTC: 0.1                      # 高倉位限制
  ETH: 1.0
  USDT: 10000.0
```

## 最佳實踐

1. **路徑選擇**：選擇具有高流動性和緊密價差的路徑
2. **價差閾值**：根據典型市場條件設置最小價差比率
3. **餘額管理**：使用適當的餘額限制控制風險
4. **流優化**：啟用獨立流獲得更好性能
5. **冷卻**：使用冷卻期防止過度交易
6. **監控**：啟用通知跟踪性能

## 限制

1. **交易所依賴性**：僅在單一交易所內工作
2. **流動性要求**：需要所有三個對都有充足流動性
3. **延遲敏感性**：性能取決於低延遲市場數據
4. **市場條件**：在高波動期間效果較差
5. **競爭**：套利機會可能被其他交易者快速消除

## 故障排除

### 常見問題

**無套利機會**
- 檢查最小價差比率是否過高
- 驗證所有符號都有充足流動性
- 確保市場數據流已連接

**IOC訂單未成交**
- 減少IOC保護比率
- 檢查訂單簿深度
- 驗證最小數量要求

**頻繁執行失敗**
- 增加餘額限制
- 檢查交易所API速率限制
- 驗證網絡連接

**盈利能力低**
- 減少最小價差比率
- 優化保護比率
- 在計算中考慮交易費用

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/tri.yaml)
- [套利交易指南](../../doc/topics/arbitrage.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
