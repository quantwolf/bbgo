# 艾略特波浪振盪器 (EWO) 數字交易策略

## 概述

EWO數字交易（ewoDgtrd）策略是一個複雜的技術分析策略，結合艾略特波浪振盪器與多個互補指標來識別高概率交易機會。它使用多層方法，包括CCI隨機指標、移動平均線、基於ATR的止損，以及可選的平均K線來增強信號質量。

## 工作原理

1. **艾略特波浪振盪器 (EWO)**：計算快速（5週期）和慢速（34週期）移動平均線之間的百分比差異
2. **信號線**：使用可配置窗口大小從EWO創建平滑信號線
3. **CCI隨機過濾器**：使用自定義CCI-隨機指標過濾交易信號
4. **趨勢確認**：採用多個移動平均線和K線模式進行趨勢驗證
5. **動態止損**：實施基於ATR的止損和止盈機制
6. **平均K線支持**：可選的平均K線用於更平滑的價格行為分析

## 主要特性

- **多指標匯合**：結合EWO、CCI隨機、移動平均線和ATR
- **靈活的移動平均線**：支持EMA、SMA或成交量加權EMA（VWEMA）
- **平均K線整合**：可選的平均K線用於降噪
- **高級止損**：多種止損機制，包括基於ATR和百分比的止損
- **動態止盈**：峰值/底部跟踪與基於ATR的止盈
- **信號過濾**：EWO變化率過濾器避免低波動期間的虛假信號
- **全面報告**：詳細的性能分析和交易分類
- **倉位管理**：使用所有可用餘額的智能倉位規模

## 策略邏輯

### 艾略特波浪振盪器計算

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
// EWO = (MA5 / MA34 - 1) * 100
s.ewo = s.ma5.Div(s.ma34).Minus(1.0).Mul(100.)
s.ewoHistogram = s.ma5.Minus(s.ma34)

// 信號線是EWO的移動平均
windowSignal := types.IntervalWindow{Interval: s.Interval, Window: s.SignalWindow}
if s.UseEma {
    sig := &indicator.EWMA{IntervalWindow: windowSignal}
    // ... 信號計算
    s.ewoSignal = sig
}
```
</augment_code_snippet>

### CCI隨機過濾器

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
type CCISTOCH struct {
    cci        *indicator.CCI
    stoch      *indicator.STOCH
    ma         *indicator.SMA
    filterHigh float64
    filterLow  float64
}

func (inc *CCISTOCH) BuySignal() bool {
    hasGrey := false
    for i := 0; i < inc.ma.Values.Length(); i++ {
        v := inc.ma.Index(i)
        if v > inc.filterHigh {
            return false
        } else if v >= inc.filterLow && v <= inc.filterHigh {
            hasGrey = true
            continue
        } else if v < inc.filterLow {
            return hasGrey
        }
    }
    return false
}
```
</augment_code_snippet>

### 入場信號生成

<augment_code_snippet path="pkg/strategy/ewoDgtrd/strategy.go" mode="EXCERPT">
```go
longSignal := types.CrossOver(s.ewo, s.ewoSignal)
shortSignal := types.CrossUnder(s.ewo, s.ewoSignal)

// 趨勢確認
bull := clozes.Last(0) > opens.Last(0)
breakThrough := clozes.Last(0) > s.ma5.Last(0) && clozes.Last(0) > s.ma34.Last(0)
breakDown := clozes.Last(0) < s.ma5.Last(0) && clozes.Last(0) < s.ma34.Last(0)

// 帶過濾器的最終信號
IsBull := bull && breakThrough && s.ccis.BuySignal() && 
          s.ewoChangeRate < s.EwoChangeFilterHigh && s.ewoChangeRate > s.EwoChangeFilterLow
IsBear := !bull && breakDown && s.ccis.SellSignal() && 
          s.ewoChangeRate < s.EwoChangeFilterHigh && s.ewoChangeRate > s.EwoChangeFilterLow

// 入場條件
if (longSignal.Index(1) && !shortSignal.Last() && IsBull) || lastPrice.Float64() <= buyLine {
    // 下買單
}
if (shortSignal.Index(1) && !longSignal.Last() && IsBear) || lastPrice.Float64() >= sellLine {
    // 下賣單
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `interval` | 間隔 | 是 | 分析時間間隔（例如："15m"、"1h"） |
| `stoploss` | 百分比 | 是 | 從入場價格的止損百分比 |

### 移動平均線配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `useEma` | 布爾值 | 否 | 使用指數移動平均線（默認：false） |
| `useSma` | 布爾值 | 否 | 當EMA為false時使用簡單移動平均線（默認：false） |
| `sigWin` | 整數 | 是 | EWO信號線的信號窗口大小 |

### CCI隨機過濾器
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `cciStochFilterHigh` | 數字 | 是 | CCI隨機的高過濾閾值（例如：80） |
| `cciStochFilterLow` | 數字 | 是 | CCI隨機的低過濾閾值（例如：20） |

### EWO變化率過濾器
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `ewoChangeFilterHigh` | 數字 | 是 | EWO變化率的高過濾器（1.0 = 無過濾） |
| `ewoChangeFilterLow` | 數字 | 是 | EWO變化率的低過濾器（0.0 = 無過濾） |

### 高級功能
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `useHeikinAshi` | 布爾值 | 否 | 使用平均K線（默認：false） |
| `disableShortStop` | 布爾值 | 否 | 禁用空頭倉位止損（默認：false） |
| `disableLongStop` | 布爾值 | 否 | 禁用多頭倉位止損（默認：false） |
| `record` | 布爾值 | 否 | 啟用詳細交易記錄（默認：false） |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  ewo_dgtrd:
    symbol: BTCUSDT
    
    # 分析時間間隔
    interval: 15m
    
    # 移動平均線設置
    useEma: false              # 使用EMA（false = 使用VWEMA）
    useSma: false              # 當useEma為false時使用SMA
    sigWin: 5                  # 信號線窗口大小
    
    # 風險管理
    stoploss: 2%               # 從入場價格2%止損
    disableShortStop: false    # 啟用空頭止損
    disableLongStop: false     # 啟用多頭止損
    
    # 平均K線
    useHeikinAshi: true        # 使用平均K線獲得更平滑信號
    
    # CCI隨機過濾器
    cciStochFilterHigh: 80     # 上閾值
    cciStochFilterLow: 20      # 下閾值
    
    # EWO變化率過濾器
    ewoChangeFilterHigh: 1.0   # 上界（1.0 = 無過濾）
    ewoChangeFilterLow: 0.0    # 下界（0.0 = 無過濾）
    
    # 調試和分析
    record: false              # 禁用詳細日誌記錄
```

## 策略組件

### 1. 艾略特波浪振盪器 (EWO)
- **快速MA**：5週期移動平均線
- **慢速MA**：34週期移動平均線
- **計算**：`(快速MA / 慢速MA - 1) × 100`
- **信號線**：使用可配置窗口的平滑EWO

### 2. CCI隨機指標
- **CCI**：28週期商品通道指數
- **隨機**：應用於CCI值
- **平滑**：隨機%D的3週期SMA
- **過濾**：信號驗證的高/低閾值

### 3. 移動平均線類型
- **EMA**：指數移動平均線（響應性）
- **SMA**：簡單移動平均線（穩定性）
- **VWEMA**：成交量加權EMA（默認，成交量感知）

### 4. 平均K線實現

<augment_code_snippet path="pkg/strategy/ewoDgtrd/heikinashi.go" mode="EXCERPT">
```go
func (inc *HeikinAshi) Update(kline types.KLine) {
    open := kline.Open.Float64()
    cloze := kline.Close.Float64()
    high := kline.High.Float64()
    low := kline.Low.Float64()
    
    newClose := (open + high + low + cloze) / 4.
    newOpen := (inc.Open.Last(0) + inc.Close.Last(0)) / 2.
    
    inc.Close.Update(newClose)
    inc.Open.Update(newOpen)
    inc.High.Update(math.Max(math.Max(high, newOpen), newClose))
    inc.Low.Update(math.Min(math.Min(low, newOpen), newClose))
    inc.Volume.Update(kline.Volume.Float64())
}
```
</augment_code_snippet>

## 入場和出場規則

### 入場條件

**多頭入場**：
1. EWO穿越信號線上方（前一根K線）且無穿越下方（當前K線）
2. 看漲K線（收盤價 > 開盤價）
3. 價格高於MA5和MA34
4. CCI隨機買入信號
5. EWO變化率在過濾器範圍內
6. 或價格觸及MA34 - ATR × 3

**空頭入場**：
1. EWO穿越信號線下方（前一根K線）且無穿越上方（當前K線）
2. 看跌K線（收盤價 < 開盤價）
3. 價格低於MA5和MA34
4. CCI隨機賣出信號
5. EWO變化率在過濾器範圍內
6. 或價格觸及MA34 + ATR × 3

### 出場條件

**止盈**：
1. EWO樞軸高/低信號（相反方向）
2. 價格達到MA34 ± ATR × 2
3. 峰值/底部跟踪與基於ATR的出場
4. 價格從峰值/底部有利移動ATR距離

**止損**：
1. 從入場價格的百分比止損
2. 基於ATR的止損（入場 ± ATR）
3. 可分別禁用多頭/空頭倉位的止損

## 風險管理功能

### 1. 多種止損機制
- **百分比止損**：從入場價格的固定百分比
- **ATR止損**：基於市場波動性的動態止損
- **選擇性禁用**：可分別禁用多頭或空頭倉位的止損

### 2. 倉位規模
- **全餘額**：每筆交易使用整個可用餘額
- **餘額驗證**：下單前確保資金充足
- **市場最小值**：遵守交易所最小數量和名義要求

### 3. 峰值/底部跟踪
- **動態跟踪**：持續更新峰值（多頭）和底部（空頭）價格
- **基於ATR的出場**：當價格對倉位不利移動ATR距離時出場
- **利潤保護**：當有利移動超過閾值時鎖定利潤

## 性能分析

策略提供全面的交易分析：

### 交易類別
- **峰值/底部與ATR**：基於峰值/底部跟踪的出場
- **CCI隨機**：基於CCI隨機信號的出場
- **多頭/空頭信號**：基於EWO樞軸信號的出場
- **MA34和ATR×2**：基於移動平均線水平的出場
- **活躍訂單**：來自新對立信號的出場
- **入場止損**：來自百分比/ATR止損的出場

### 報告指標
- **勝率**：盈利交易的百分比
- **平均盈虧**：按出場類別的平均利潤/虧損
- **交易計數**：每種出場類型的交易數量
- **性能摘要**：策略關閉時的詳細分解

## 常見用例

### 1. 趨勢跟隨設置
```yaml
interval: 1h
useEma: true
sigWin: 10
stoploss: 3%
useHeikinAshi: true
```

### 2. 剝頭皮設置
```yaml
interval: 5m
useEma: false
useSma: true
sigWin: 3
stoploss: 1%
useHeikinAshi: false
```

### 3. 保守設置
```yaml
interval: 4h
useEma: false
sigWin: 8
stoploss: 5%
cciStochFilterHigh: 70
cciStochFilterLow: 30
```

## 最佳實踐

1. **時間框架選擇**：更高時間框架（1h+）用於趨勢跟隨，更低（5m-15m）用於剝頭皮
2. **平均K線使用**：在波動市場中啟用以獲得更平滑信號
3. **過濾器調整**：根據市場條件調整CCI隨機過濾器
4. **止損管理**：在強趨勢市場中考慮禁用止損
5. **EWO變化過濾器**：用於避免低波動期間的交易
6. **信號窗口**：較小窗口用於更快信號，較大窗口用於穩定性

## 限制

1. **鋸齒風險**：在橫盤市場中可能產生虛假信號
2. **滯後性**：多個指標造成固有信號滯後
3. **全倉位**：使用整個餘額，限制風險管理靈活性
4. **複雜性**：許多參數需要仔細調整
5. **市場依賴性**：在不同市場條件下性能差異顯著

## 故障排除

### 常見問題

**未生成交易**
- 檢查是否滿足所有過濾條件
- 驗證CCI隨機閾值是否適當
- 確保EWO變化率過濾器不過於嚴格

**過度虧損**
- 增加止損百分比
- 收緊CCI隨機過濾器
- 考慮啟用平均K線以獲得更平滑信號

**信號質量差**
- 調整信號窗口大小
- 修改EWO變化率過濾器
- 嘗試不同的移動平均線類型

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/ewo_dgtrd.yaml)
- [艾略特波浪理論](../../doc/topics/elliott-wave.md)
- [技術指標指南](../../doc/topics/indicators.md)
