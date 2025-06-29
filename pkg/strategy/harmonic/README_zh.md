# 諧波策略

## 概述

諧波策略是一個高級技術分析策略，結合諧波模式識別與隱馬爾可夫模型（HMM）信號過濾。它使用SHARK諧波模式指標識別潛在反轉點，並採用HMM對交易信號進行去噪，提供基於模式的複雜交易方法和統計信號處理。

## 工作原理

1. **SHARK模式檢測**：使用斐波那契比率識別價格行為中的SHARK諧波模式
2. **信號生成**：基於模式完成計算多頭和空頭分數
3. **HMM過濾**：應用隱馬爾可夫模型對交易信號進行去噪和過濾
4. **狀態管理**：使用狀態機方法，包含多頭（1）、中性（0）和空頭（-1）狀態
5. **倉位管理**：基於過濾信號和狀態轉換開倉和平倉
6. **性能分析**：全面的利潤跟踪和可視化功能

## 主要特性

- **SHARK諧波模式識別**：使用斐波那契回撤的高級模式檢測
- **隱馬爾可夫模型過濾**：統計信號去噪以提高準確性
- **多狀態交易系統**：多頭、中性和空頭狀態管理
- **實時模式評分**：持續評估看漲和看跌模式
- **全面分析**：詳細的利潤報告和性能可視化
- **互動命令**：實時盈虧和累積利潤圖表生成
- **回測支持**：完整的回測功能和圖表生成
- **退出方法整合**：與各種退出策略兼容

## 策略邏輯

### SHARK模式檢測

<augment_code_snippet path="pkg/strategy/harmonic/shark.go" mode="EXCERPT">
```go
func (inc SHARK) SharkLong(highs, lows floats.Slice, p float64, lookback int) float64 {
    score := 0.
    for x := 5; x < lookback; x++ {
        if lows.Index(x-1) > lows.Index(x) && lows.Index(x) < lows.Index(x+1) {
            X := lows.Index(x)
            for a := 4; a < x; a++ {
                if highs.Index(a-1) < highs.Index(a) && highs.Index(a) > highs.Index(a+1) {
                    A := highs.Index(a)
                    XA := math.Abs(X - A)
                    hB := A - 0.382*XA  // 38.2%斐波那契回撤
                    lB := A - 0.618*XA  // 61.8%斐波那契回撤
                    // 模式驗證繼續...
                }
            }
        }
    }
    return score
}
```
</augment_code_snippet>

### 隱馬爾可夫模型實現

<augment_code_snippet path="pkg/strategy/harmonic/strategy.go" mode="EXCERPT">
```go
func hmm(y_t []float64, x_t []float64, l int) float64 {
    // 信號過濾的HMM實現
    // 狀態：-1（空頭）、0（中性）、1（多頭）
    
    for n := 2; n <= len(x_t); n++ {
        for j := -1; j <= 1; j++ {
            // 計算轉換概率
            for i := -1; i <= 1; i++ {
                transitProb := transitProbability(i, j)
                observeProb := observeDistribution(y_t[n-1], float64(j))
                // 更新每個狀態的alpha值
            }
        }
    }
    
    // 返回最可能的狀態
    if maximum[0] == long {
        return 1
    } else if maximum[0] == short {
        return -1
    }
    return 0
}
```
</augment_code_snippet>

### 交易信號邏輯

<augment_code_snippet path="pkg/strategy/harmonic/strategy.go" mode="EXCERPT">
```go
s.session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {
    log.Infof("shark score: %f, current price: %f", s.shark.Last(0), kline.Close.Float64())
    
    nextState := hmm(s.shark.Array(s.Window), states.Array(s.Window), s.Window)
    states.Update(nextState)
    log.Infof("Denoised signal via HMM: %f", states.Last(0))
    
    if states.Length() < s.Window {
        return
    }
    
    // 當信號變為中性時平倉
    if s.Position.IsOpened(kline.Close) && states.Mean(5) == 0 {
        s.orderExecutor.ClosePosition(ctx, fixedpoint.One)
    }
    
    // 開多頭倉位
    if states.Mean(5) == 1 && direction != 1 {
        s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
            Symbol:   s.Symbol,
            Side:     types.SideTypeBuy,
            Quantity: s.Quantity,
            Type:     types.OrderTypeMarket,
            Tag:      "sharkLong",
        })
    }
    
    // 開空頭倉位
    if states.Mean(5) == -1 && direction != -1 {
        s.orderExecutor.SubmitOrders(ctx, types.SubmitOrder{
            Symbol:   s.Symbol,
            Side:     types.SideTypeSell,
            Quantity: s.Quantity,
            Type:     types.OrderTypeMarket,
            Tag:      "sharkShort",
        })
    }
}))
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `interval` | 間隔 | 是 | 分析時間間隔（例如："1s"、"1m"） |
| `window` | 整數 | 是 | 模式檢測的回望窗口 |
| `quantity` | 數字 | 是 | 每筆交易的固定數量 |

### 可視化設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `drawGraph` | 布爾值 | 否 | 回測期間啟用圖表生成 |
| `graphPNLPath` | 字符串 | 否 | 盈虧百分比圖表輸出路徑 |
| `graphCumPNLPath` | 字符串 | 否 | 累積盈虧圖表輸出路徑 |

### 退出方法
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `exits` | 數組 | 否 | 退出方法配置數組 |

### 累積利潤報告
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `accumulatedProfitReport.accumulatedProfitMAWindow` | 整數 | 否 | 累積利潤的SMA窗口（默認：60） |
| `accumulatedProfitReport.intervalWindow` | 整數 | 否 | 間隔窗口（天）（默認：7） |
| `accumulatedProfitReport.numberOfInterval` | 整數 | 否 | 輸出間隔數量（默認：1） |
| `accumulatedProfitReport.tsvReportPath` | 字符串 | 否 | TSV報告輸出路徑 |
| `accumulatedProfitReport.accumulatedDailyProfitWindow` | 整數 | 否 | 每日利潤累積窗口（默認：7） |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  harmonic:
    symbol: BTCUSDT
    
    # 時間設置
    interval: 1s              # 1秒分析間隔
    window: 60                # 60週期回望窗口
    
    # 倉位規模
    quantity: 0.005           # 每筆交易0.005 BTC
    
    # 可視化（用於回測）
    drawGraph: true           # 啟用圖表生成
    graphPNLPath: "./pnl.png"
    graphCumPNLPath: "./cumpnl.png"
    
    # 退出方法
    exits:
    - roiStopLoss:
        percentage: 0.02      # 2%止損
    - roiTakeProfit:
        percentage: 0.05      # 5%止盈
    
    # 累積利潤報告
    accumulatedProfitReport:
      accumulatedProfitMAWindow: 60
      intervalWindow: 7
      numberOfInterval: 4
      tsvReportPath: "./profit_report.tsv"
      accumulatedDailyProfitWindow: 7
```

## SHARK諧波模式

### 模式結構
SHARK模式是一個5點諧波模式（X-A-B-C-D），具有特定的斐波那契比率：

**看漲SHARK模式：**
- **點X**：初始低點
- **點A**：X後的高點
- **點B**：XA的回撤（38.2% - 61.8%）
- **點C**：AB的延伸（113% - 161.8%）
- **點D**：XC的回撤（88.6% - 113%）和BC（161.8% - 224%）

**看跌SHARK模式：**
- **點X**：初始高點
- **點A**：X後的低點
- **點B**：XA的回撤（38.2% - 61.8%）
- **點C**：AB的延伸（113% - 161.8%）
- **點D**：XC的延伸（88.6% - 113%）和BC（161.8% - 224%）

### 使用的斐波那契比率
- **B點**：XA的0.382 - 0.618
- **C點**：AB的1.13 - 1.618
- **D點**：XC的0.886 - 1.13和BC的1.618 - 2.24

## 隱馬爾可夫模型

### 狀態定義
- **狀態1**：多頭/看漲信號
- **狀態0**：中性信號
- **狀態-1**：空頭/看跌信號

### 轉換概率
- **相同狀態**：0.99（保持當前狀態的高概率）
- **狀態變化**：0.01（狀態轉換的低概率）

### 觀測模型
觀測分佈評估SHARK指標值與當前狀態之間的一致性：
- **一致**：返回1.0（高概率）
- **不一致**：返回0.0（低概率）

## 交易邏輯

### 入場條件

**多頭入場：**
1. HMM過濾狀態在5個週期的平均值等於1
2. 當前倉位不是已經多頭
3. SHARK多頭分數表明看漲模式完成

**空頭入場：**
1. HMM過濾狀態在5個週期的平均值等於-1
2. 當前倉位不是已經空頭
3. SHARK空頭分數表明看跌模式完成

### 出場條件

**中性信號出場：**
1. HMM過濾狀態在5個週期的平均值等於0
2. 倉位當前開放
3. 平倉100%的倉位

**退出方法整合：**
- 與基於ROI的止損和止盈兼容
- 支持追蹤止損和其他退出策略
- 可以結合多種退出方法

## 性能分析

### 實時監控
- **SHARK分數跟踪**：持續監控模式分數
- **HMM狀態記錄**：實時狀態轉換記錄
- **倉位跟踪**：詳細的倉位和利潤監控

### 累積利潤報告
- **每日利潤計算**：跟踪每日盈虧
- **移動平均分析**：累積利潤的SMA
- **勝率跟踪**：每日勝率統計
- **利潤因子分析**：風險調整後的性能指標
- **交易計數監控**：每期交易數量

### 可視化功能
- **盈虧百分比圖表**：逐筆交易利潤百分比
- **累積盈虧圖表**：隨時間的總投資組合價值
- **互動命令**：`/pnl`和`/cumpnl`用於實時圖表
- **TSV報告導出**：詳細性能數據導出

## 互動命令

### 實時圖表生成
```
/pnl      - 生成盈虧百分比圖表
/cumpnl   - 生成累積盈虧圖表
```

這些命令生成可通過消息平台分享的實時性能圖表。

## 常見用例

### 1. 高頻模式交易
```yaml
interval: 1s
window: 30
quantity: 0.001
drawGraph: false
```

### 2. 中期模式識別
```yaml
interval: 5m
window: 100
quantity: 0.01
exits:
- roiStopLoss:
    percentage: 0.03
- roiTakeProfit:
    percentage: 0.08
```

### 3. 保守模式交易
```yaml
interval: 15m
window: 200
quantity: 0.005
exits:
- roiStopLoss:
    percentage: 0.02
- roiTakeProfit:
    percentage: 0.06
- trailingStop:
    callbackRate: 0.01
```

## 最佳實踐

1. **窗口大小選擇**：更大的窗口用於更可靠的模式，更小的用於響應性
2. **間隔優化**：將間隔與交易風格和市場波動性匹配
3. **風險管理**：始終使用退出方法進行風險控制
4. **模式驗證**：監控SHARK分數以了解模式強度
5. **HMM調整**：考慮為不同市場調整轉換概率
6. **回測**：使用圖表生成來可視化策略性能

## 限制

1. **模式依賴性**：依賴可能罕見的諧波模式形成
2. **計算密集性**：複雜的模式檢測需要大量處理
3. **市場條件**：在不同市場制度下性能差異顯著
4. **滯後性**：模式完成檢測引入固有滯後
5. **虛假信號**：HMM過濾減少但不能消除虛假信號

## 故障排除

### 常見問題

**無交易信號**
- 檢查窗口大小是否允許充分的模式檢測
- 驗證SHARK分數是否正在生成
- 確保HMM狀態正確轉換

**過度交易**
- 增加窗口大小以獲得更穩定的信號
- 調整HMM轉換概率
- 添加額外的退出方法進行風險控制

**模式識別差**
- 驗證價格數據質量和完整性
- 檢查斐波那契比率計算
- 監控SHARK分數生成日誌

**HMM狀態問題**
- 檢查觀測分佈邏輯
- 檢查轉換概率設置
- 驗證狀態數組初始化

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/harmonic.yaml)
- [諧波模式交易指南](../../doc/topics/harmonic-patterns.md)
- [交易中的隱馬爾可夫模型](../../doc/topics/hmm-trading.md)
