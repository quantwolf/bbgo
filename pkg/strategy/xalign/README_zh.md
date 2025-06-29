# XAlign 策略

## 概述

XAlign策略是一個複雜的跨交易所餘額對齊策略，自動維護多個交易所間的目標餘額分配。它監控餘額與預期水平的偏差，當差異超過容忍閾值並持續一段時間時執行交易以重新對齊投資組合。該策略特別適用於在不同交易所間維持一致的投資組合分配或實施跨交易所套利機會。

## 工作原理

1. **餘額監控**：持續監控多個交易所的帳戶餘額
2. **偏差檢測**：使用統計分析檢測餘額何時偏離預期目標
3. **持續偏差跟踪**：跟踪偏差持續多長時間後觸發行動
4. **智能訂單放置**：放置買賣訂單以將餘額重新對齊到目標分配
5. **報價貨幣選擇**：智能選擇不同交易方向的最優報價貨幣
6. **警報系統**：當檢測到或糾正大差異時發送通知

## 主要特性

- **跨交易所餘額管理**：同時監控和對齊多個交易所的餘額
- **偏差檢測系統**：高級統計分析識別顯著餘額偏差
- **持續偏差跟踪**：僅對持續可配置時間段的偏差採取行動
- **智能報價貨幣選擇**：買賣操作使用不同的報價貨幣
- **大額警報**：顯著餘額差異的Slack通知
- **靈活容忍設置**：可配置的容忍範圍和時間閾值
- **模擬運行支持**：策略驗證的測試模式，無實際交易
- **訂單類型選擇**：支持做市商和吃單者訂單

## 策略架構

### 偏差檢測系統

<augment_code_snippet path="pkg/strategy/xalign/detector/deviation.go" mode="EXCERPT">
```go
type DeviationDetector[T any] struct {
    mu            sync.Mutex
    expectedValue T             // 用於比較的預期值
    tolerance     float64       // 容忍百分比（例如，0.01表示1%）
    duration      time.Duration // 持續偏差的時間限制

    toFloat64Amount func(T) (float64, error) // 將T轉換為float64的函數
    records         []Record[T]              // 跟踪偏差記錄
}

func (d *DeviationDetector[T]) AddRecord(at time.Time, value T) (bool, time.Duration) {
    // 計算偏差百分比
    deviationPercentage := math.Abs((current - expected) / expected)
    
    // 如果偏差在容忍範圍內則重置記錄
    if deviationPercentage <= d.tolerance {
        d.records = nil
        return false, 0
    }
    
    // 跟踪持續偏差
    d.records = append(d.records, record)
    return d.ShouldFix()
}
```
</augment_code_snippet>

### 餘額對齊邏輯

<augment_code_snippet path="pkg/strategy/xalign/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) align(ctx context.Context, sessions map[string]*bbgo.ExchangeSession) {
    for currency, expectedBalance := range s.ExpectedBalances {
        // 計算所有會話的總餘額
        totalBalance := s.calculateTotalBalance(currency, sessions)
        
        // 使用檢測器檢查偏差
        shouldFix, sustainedDuration := s.detectors[currency].AddRecord(time.Now(), totalBalance)
        
        if shouldFix {
            delta := totalBalance.Sub(expectedBalance)
            s.executeBalanceAlignment(ctx, currency, delta, sustainedDuration, sessions)
        }
    }
}
```
</augment_code_snippet>

### 警報系統

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

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `sessions` | 數組 | 是 | 要監控的交易所會話列表 |
| `interval` | 持續時間 | 是 | 餘額檢查間隔（例如："1m"、"30s"） |
| `for` | 持續時間 | 是 | 持續偏差持續時間閾值 |
| `expectedBalances` | 對象 | 是 | 每種貨幣的目標餘額 |

### 交易設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quoteCurrencies.buy` | 數組 | 是 | 買單的首選報價貨幣 |
| `quoteCurrencies.sell` | 數組 | 是 | 賣單的首選報價貨幣 |
| `useTakerOrder` | 布爾值 | 否 | 使用吃單者訂單而非做市商訂單 |
| `balanceToleranceRange` | 百分比 | 是 | 餘額偏差的容忍範圍 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `maxAmounts` | 對象 | 是 | 每種報價貨幣的最大交易金額 |
| `dryRun` | 布爾值 | 否 | 啟用模擬運行模式（無實際交易） |

### 警報配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `largeAmountAlert.quoteCurrency` | 字符串 | 否 | 警報計算的報價貨幣 |
| `largeAmountAlert.amount` | 數字 | 否 | 大額警報的閾值 |
| `largeAmountAlert.slack.channel` | 字符串 | 否 | 警報的Slack頻道 |
| `largeAmountAlert.slack.mentions` | 數組 | 否 | 警報中要提及的用戶/群組 |

## 配置示例

```yaml
crossExchangeStrategies:
- xalign:
    # 監控設置
    interval: 1m                    # 每分鐘檢查餘額
    for: 5m                         # 對持續5分鐘以上的偏差採取行動
    
    # 要監控的交易所會話
    sessions:
    - binance
    - max
    - okex
    
    # 目標餘額分配
    expectedBalances:
      BTC: 1.0                      # 在交易所間維持總計1.0 BTC
      ETH: 10.0                     # 在交易所間維持總計10.0 ETH
      USDT: 50000                   # 維持總計$50,000 USDT
    
    # 報價貨幣偏好
    quoteCurrencies:
      buy: [USDT, USDC, BUSD]       # 買單偏好USDT
      sell: [USDT, USDC]            # 賣單偏好USDT
    
    # 風險管理
    balanceToleranceRange: 2%       # 觸發前2%容忍度
    maxAmounts:
      USDT: 1000                    # 每筆交易最大$1000
      USDC: 1000                    # 每筆交易最大$1000 USDC
      BUSD: 1000                    # 每筆交易最大$1000 BUSD
    
    # 訂單執行
    useTakerOrder: false            # 使用做市商訂單獲得更好費率
    dryRun: false                   # 實盤交易模式
    
    # 警報系統
    largeAmountAlert:
      quoteCurrency: USDT
      amount: 5000                  # 交易>$5000時警報
      slack:
        channel: "trading-alerts"
        mentions:
        - '<@USER_ID>'              # 提及特定用戶
        - '<!subteam^TEAM_ID>'      # 提及團隊
```

## 餘額對齊邏輯

### 偏差檢測過程
1. **持續監控**：按指定間隔檢查餘額
2. **偏差計算**：`偏差 = |當前餘額 - 預期餘額| / 預期餘額`
3. **容忍檢查**：將偏差與`balanceToleranceRange`比較
4. **持續跟踪**：記錄超過容忍度的偏差
5. **行動觸發**：當偏差持續`for`持續時間時採取行動

### 交易執行邏輯
1. **差額計算**：確定要買入或賣出多少
2. **報價貨幣選擇**：根據交易方向選擇最優報價貨幣
3. **市場選擇**：找到具有最佳流動性/定價的交易所
4. **訂單放置**：使用適當的訂單類型執行交易
5. **確認**：驗證交易執行並更新跟踪

## 報價貨幣選擇

### 買單
```yaml
quoteCurrencies:
  buy: [USDT, USDC, BUSD]
```
- 策略首先嘗試USDT，然後USDC，然後BUSD
- 選擇第一個具有足夠流動性的可用對
- 考慮目標交易所的餘額可用性

### 賣單
```yaml
quoteCurrencies:
  sell: [USDT, USDC]
```
- 策略首先嘗試USDT，然後USDC
- 優化最佳價格和流動性
- 確保結算有足夠的報價貨幣餘額

## 風險管理功能

### 餘額容忍度
- **基於百分比**：使用相對容忍度（例如，預期餘額的2%）
- **持續時間**：僅對持續偏差採取行動
- **最大金額**：限制每種貨幣的個別交易規模

### 訂單管理
- **做市商訂單**：默認使用做市商訂單獲得更好的費率
- **吃單者訂單**：可選吃單者訂單立即執行
- **市場選擇**：為每筆交易選擇最優交易所

### 警報系統
- **大額警報**：重大交易的通知
- **偏差警告**：行動閾值前的早期警告
- **Slack整合**：帶用戶提及的實時通知

## 使用案例

### 1. 跨交易所投資組合平衡
```yaml
expectedBalances:
  BTC: 2.0
  ETH: 20.0
  USDT: 100000
balanceToleranceRange: 1%
for: 10m
```
**目的**：在交易所間維持一致的投資組合分配
**好處**：減少集中風險並優化資本效率

### 2. 套利機會準備
```yaml
expectedBalances:
  BTC: 0.5
  USDT: 25000
balanceToleranceRange: 0.5%
for: 2m
useTakerOrder: true
```
**目的**：快速重新對齊餘額以獲得套利機會
**好處**：使用吃單者訂單更快執行時間敏感機會

### 3. 保守餘額管理
```yaml
expectedBalances:
  BTC: 1.0
  USDT: 50000
balanceToleranceRange: 5%
for: 30m
maxAmounts:
  USDT: 500
```
**目的**：使用保守限制進行漸進式餘額調整
**好處**：最小化市場影響和交易成本

## 性能優化

### 監控頻率
- **高頻（30s-1m）**：用於主動套利策略
- **中頻（5m-15m）**：用於一般投資組合管理
- **低頻（1h+）**：用於長期分配維護

### 容忍度調整
- **緊密容忍度（0.5-1%）**：更頻繁的重新平衡，更高精度
- **中等容忍度（2-3%）**：大多數用例的平衡方法
- **寬鬆容忍度（5%+）**：最小重新平衡，更低成本

## 最佳實踐

1. **從模擬運行開始**：實盤交易前測試配置
2. **保守限制**：從小的最大金額開始
3. **監控警報**：設置適當的Slack通知
4. **定期檢查**：定期評估預期餘額目標
5. **費用考慮**：盡可能使用做市商訂單減少成本
6. **交易所選擇**：確保所有監控交易所的良好流動性

## 限制

1. **市場影響**：大型重新平衡交易可能影響市場價格
2. **交易所限制**：受個別交易所交易限制約束
3. **網絡延遲**：跨交易所協調可能有延遲
4. **費用成本**：頻繁重新平衡可能累積大量費用
5. **市場條件**：極端波動可能觸發過度交易

## 故障排除

### 常見問題

**無重新平衡行動**
- 檢查偏差是否超過容忍閾值
- 驗證是否滿足持續時間要求
- 確保有足夠的交易餘額

**過度交易**
- 增加容忍範圍
- 延長持續時間要求
- 檢查預期餘額目標

**警報垃圾郵件**
- 調整大額警報閾值
- 檢查容忍設置
- 檢查交易所連接問題

**訂單失敗**
- 驗證交易所API權限
- 檢查最小訂單要求
- 確保報價貨幣餘額充足

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/xalign.yaml)
- [跨交易所交易指南](../../doc/topics/cross-exchange.md)
- [餘額管理最佳實踐](../../doc/topics/balance-management.md)
