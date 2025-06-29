# SuperTrend 策略

## 概述

SuperTrend策略是一個複雜的趨勢跟蹤策略，結合SuperTrend指標與額外確認信號來識別和交易市場趨勢。它使用平均真實範圍（ATR）計算動態支撐和阻力位，並通過雙重DEMA（雙指數移動平均）和線性回歸過濾器增強，以減少虛假信號並提高交易質量。

## 工作原理

1. **SuperTrend計算**：使用ATR和乘數創建動態支撐/阻力帶
2. **趨勢識別**：基於價格相對於SuperTrend帶的位置確定看漲/看跌趨勢
3. **信號確認**：使用雙重DEMA和線性回歸作為額外確認過濾器
4. **倉位管理**：當所有信號一致時開倉，基於各種退出條件平倉
5. **風險管理**：整合多種退出方法，包括止損、止盈和追蹤止損
6. **性能跟踪**：全面的利潤統計和可視化功能

## 主要特性

- **SuperTrend指標**：使用基於ATR的帶狀動態趨勢識別
- **多信號確認**：雙重DEMA和線性回歸過濾器提高信號質量
- **靈活倉位規模**：支持固定數量和基於槓桿的規模
- **高級退出方法**：多種退出策略，包括反向信號退出
- **噪音過濾**：DEMA交叉檢測以過濾市場噪音
- **趨勢確認**：線性回歸斜率分析進行趨勢驗證
- **可視化支持**：內置盈虧圖表生成進行性能分析
- **全面風險管理**：多種止損和止盈機制

## 策略架構

### SuperTrend計算

<augment_code_snippet path="pkg/strategy/supertrend/strategy.go" mode="EXCERPT">
```go
// 使用ATR和乘數計算SuperTrend
func (s *Strategy) calculateSuperTrend(kline types.KLine) {
    // 獲取ATR值
    atr := s.atr.Last(0)
    
    // 計算基本上下帶
    hl2 := (kline.High.Add(kline.Low)).Div(fixedpoint.NewFromFloat(2.0))
    upperBand := hl2.Add(fixedpoint.NewFromFloat(s.SupertrendMultiplier * atr))
    lowerBand := hl2.Sub(fixedpoint.NewFromFloat(s.SupertrendMultiplier * atr))
    
    // 確定趨勢方向
    if kline.Close.Compare(s.superTrend.Last(0)) > 0 {
        s.currentTrend = types.DirectionUp
    } else {
        s.currentTrend = types.DirectionDown
    }
}
```
</augment_code_snippet>

### 雙重DEMA信號

<augment_code_snippet path="pkg/strategy/supertrend/double_dema.go" mode="EXCERPT">
```go
// getDemaSignal 獲取當前DEMA信號
func (dd *DoubleDema) getDemaSignal(openPrice float64, closePrice float64) types.Direction {
    var demaSignal types.Direction = types.DirectionNone

    // 看漲突破：收盤價高於兩個DEMA但開盤價不是
    if closePrice > dd.fastDEMA.Last(0) && closePrice > dd.slowDEMA.Last(0) && 
       !(openPrice > dd.fastDEMA.Last(0) && openPrice > dd.slowDEMA.Last(0)) {
        demaSignal = types.DirectionUp
    // 看跌跌破：收盤價低於兩個DEMA但開盤價不是
    } else if closePrice < dd.fastDEMA.Last(0) && closePrice < dd.slowDEMA.Last(0) && 
              !(openPrice < dd.fastDEMA.Last(0) && openPrice < dd.slowDEMA.Last(0)) {
        demaSignal = types.DirectionDown
    }

    return demaSignal
}
```
</augment_code_snippet>

### 線性回歸趨勢

<augment_code_snippet path="pkg/strategy/supertrend/linreg.go" mode="EXCERPT">
```go
// GetSignal 獲取線性回歸信號
func (lr *LinReg) GetSignal() types.Direction {
    var lrSignal types.Direction = types.DirectionNone

    switch {
    case lr.Last(0) > 0:  // 正斜率 = 上升趨勢
        lrSignal = types.DirectionUp
    case lr.Last(0) < 0:  // 負斜率 = 下降趨勢
        lrSignal = types.DirectionDown
    }

    return lrSignal
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `interval` | 間隔 | 是 | 分析時間間隔（例如："1m"、"5m"） |
| `window` | 整數 | 是 | SuperTrend計算的ATR窗口 |
| `supertrendMultiplier` | 浮點數 | 是 | SuperTrend帶中ATR的乘數 |

### 倉位規模
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quantity` | 數字 | 否 | 每筆交易的固定數量 |
| `leverage` | 浮點數 | 否 | 倉位規模的槓桿乘數 |

### 信號過濾器
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `fastDEMAWindow` | 整數 | 否 | 快速DEMA週期（默認：144） |
| `slowDEMAWindow` | 整數 | 否 | 慢速DEMA週期（默認：169） |
| `linearRegression.interval` | 間隔 | 否 | 線性回歸分析間隔 |
| `linearRegression.window` | 整數 | 否 | 線性回歸窗口週期 |

### 退出條件
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `takeProfitAtrMultiplier` | 浮點數 | 否 | 基於ATR倍數的止盈 |
| `stopLossByTriggeringK` | 布爾值 | 否 | 將止損設為觸發蠟燭低點 |
| `stopByReversedSupertrend` | 布爾值 | 否 | 反向SuperTrend信號時退出 |
| `stopByReversedDema` | 布爾值 | 否 | 反向DEMA信號時退出 |
| `stopByReversedLinGre` | 布爾值 | 否 | 反向線性回歸信號時退出 |

### 可視化
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `drawGraph` | 布爾值 | 否 | 啟用盈虧圖表生成 |
| `graphPNLPath` | 字符串 | 否 | 盈虧圖表輸出路徑 |
| `graphCumPNLPath` | 字符串 | 否 | 累積盈虧圖表輸出路徑 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  supertrend:
    symbol: BTCUSDT
    
    # 基本SuperTrend設置
    interval: 1m
    window: 220                    # ATR窗口
    supertrendMultiplier: 10       # 帶狀ATR乘數
    
    # 倉位規模
    leverage: 1.0                  # 使用1倍槓桿
    # quantity: 0.5                # 或固定0.5 BTC數量
    
    # DEMA過濾器設置
    fastDEMAWindow: 28             # 快速DEMA週期
    slowDEMAWindow: 170            # 慢速DEMA週期
    
    # 線性回歸確認
    linearRegression:
      interval: 1m
      window: 18                   # 18週期線性回歸
    
    # 退出方法設置
    takeProfitAtrMultiplier: 0     # 禁用（0）
    stopLossByTriggeringK: false   # 不使用觸發蠟燭
    stopByReversedSupertrend: false # 不在反向信號時退出
    stopByReversedDema: false
    stopByReversedLinGre: false
    
    # 可視化
    drawGraph: true
    graphPNLPath: "./pnl.png"
    graphCumPNLPath: "./cumpnl.png"
    
    # 高級退出方法
    exits:
    - roiStopLoss:
        percentage: 2%             # 2%止損
    - trailingStop:
        callbackRate: 2%           # 2%追蹤止損
        minProfit: 10%             # 10%利潤後激活
        interval: 1m
        side: both
        closePosition: 100%
```

## SuperTrend指標解釋

### 什麼是SuperTrend？
SuperTrend是一個趨勢跟蹤指標，使用平均真實範圍（ATR）計算動態支撐和阻力位。它基於價格相對於趨勢帶的位置提供清晰的買賣信號。

### 計算方法
1. **ATR計算**：`ATR = 指定窗口內的平均真實範圍`
2. **基本帶狀**：
   - `上帶 = (最高價 + 最低價) / 2 + (乘數 × ATR)`
   - `下帶 = (最高價 + 最低價) / 2 - (乘數 × ATR)`
3. **趨勢確定**：
   - **上升趨勢**：價格高於SuperTrend線（使用下帶）
   - **下降趨勢**：價格低於SuperTrend線（使用上帶）

### 信號生成
- **買入信號**：價格突破SuperTrend線上方
- **賣出信號**：價格跌破SuperTrend線下方
- **趨勢延續**：價格保持在SuperTrend線同一側

## 多信號確認系統

### 1. SuperTrend（主要信號）
- **目的**：主要趨勢識別
- **優勢**：清晰的趨勢方向，滯後最小
- **劣勢**：在震盪市場中可能產生虛假信號

### 2. 雙重DEMA過濾器
- **目的**：噪音減少和突破確認
- **邏輯**：要求價格突破快慢DEMA上方/下方
- **好處**：過濾小幅價格波動

### 3. 線性回歸確認
- **目的**：趨勢強度驗證
- **方法**：分析線性回歸線的斜率
- **信號**：正斜率=上升趨勢，負斜率=下降趨勢

## 交易邏輯

### 入場條件

**多頭入場：**
1. SuperTrend指示上升趨勢（價格高於SuperTrend線）
2. DEMA過濾器確認看漲突破（可選）
3. 線性回歸顯示正斜率（可選）
4. 所有啟用的過濾器必須一致

**空頭入場：**
1. SuperTrend指示下降趨勢（價格低於SuperTrend線）
2. DEMA過濾器確認看跌跌破（可選）
3. 線性回歸顯示負斜率（可選）
4. 所有啟用的過濾器必須一致

### 退出條件

**標準退出：**
- 反向SuperTrend信號
- 反向DEMA信號
- 反向線性回歸信號
- 基於ATR的止盈
- 觸發蠟燭低點止損

**高級退出（通過退出方法）：**
- 基於ROI的止損
- 追蹤止損
- 更高高點/更低低點模式
- 基於時間的退出

## 倉位規模

### 固定數量模式
```yaml
quantity: 0.5  # 每筆交易固定0.5 BTC
```
- 所有交易使用相同數量
- 簡單且可預測的倉位規模
- 適合每筆交易一致風險

### 槓桿模式
```yaml
leverage: 1.0  # 1倍槓桿
```
- 基於帳戶淨值計算倉位規模
- `倉位規模 = 帳戶價值 × 槓桿 / 當前價格`
- 自動適應帳戶餘額變化

## 風險管理功能

### 內置保護
- **基於ATR的止損**：基於波動性的動態止損水平
- **趨勢反轉退出**：趨勢變化時自動退出
- **多重確認**：減少虛假信號風險

### 可配置退出
- **止損**：基於百分比或ATR
- **止盈**：ATR倍數或基於百分比
- **追蹤止損**：動態利潤保護
- **基於模式的退出**：更高高點/更低低點檢測

## 性能優化

### 參數調整

**ATR窗口：**
- **較短（50-100）**：更響應，更多信號
- **中等（150-250）**：平衡方法
- **較長（300+）**：更平滑，更少虛假信號

**SuperTrend乘數：**
- **較低（3-7）**：更敏感，更多交易
- **中等（8-12）**：平衡敏感性
- **較高（15+）**：較不敏感，僅強趨勢

**DEMA窗口：**
- **快速DEMA**：20-50週期響應性
- **慢速DEMA**：100-200週期穩定性

## 常見用例

### 1. 保守趨勢跟蹤
```yaml
window: 300
supertrendMultiplier: 15
fastDEMAWindow: 50
slowDEMAWindow: 200
leverage: 0.5
```

### 2. 激進趨勢交易
```yaml
window: 100
supertrendMultiplier: 5
fastDEMAWindow: 20
slowDEMAWindow: 100
leverage: 2.0
```

### 3. 平衡方法
```yaml
window: 220
supertrendMultiplier: 10
fastDEMAWindow: 28
slowDEMAWindow: 170
leverage: 1.0
```

## 最佳實踐

1. **市場選擇**：在趨勢市場中效果最佳
2. **時間框架**：較高時間框架通常更可靠
3. **確認**：使用多個信號提高準確性
4. **風險管理**：始終使用止損和倉位規模
5. **回測**：在歷史數據上測試參數
6. **市場條件**：根據波動性調整敏感性

## 限制

1. **橫盤市場**：在區間市場中可能產生許多虛假信號
2. **滯後**：趨勢跟蹤性質意味著信號有一定延遲
3. **鋸齒**：快速趨勢變化可能導致多次損失
4. **參數敏感性**：性能隨參數選擇顯著變化
5. **市場跳空**：無法防止隔夜跳空

## 故障排除

### 常見問題

**過多虛假信號**
- 增加SuperTrend乘數
- 使用更長的ATR窗口
- 啟用DEMA和線性回歸過濾器

**錯過趨勢移動**
- 減少SuperTrend乘數
- 使用更短的ATR窗口
- 減少過濾器要求

**風險/回報差**
- 調整退出方法
- 使用追蹤止損
- 優化止盈水平

**過度回撤**
- 減少槓桿
- 收緊止損
- 使用更保守的參數

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/supertrend.yaml)
- [SuperTrend指標指南](../../doc/topics/supertrend-indicator.md)
- [趨勢跟蹤策略](../../doc/topics/trend-following.md)
