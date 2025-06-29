# 漂移策略

## 概述

漂移策略是一個複雜的動量交易策略，使用自定義的漂移移動平均（DriftMA）指標來識別市場動量變化和趨勢轉換。它結合加權漂移計算與多重平滑技術和高級風險管理功能，包括追蹤止損、基於ATR的止損和動態再平衡。

## 工作原理

1. **漂移計算**：使用加權漂移指標測量價格運動的動量變化
2. **信號平滑**：使用EWMA和Fisher變換應用多層平滑
3. **趨勢檢測**：通過漂移交叉和方向變化識別動量轉移
4. **動態定位**：根據市場波動性和趨勢強度調整倉位規模
5. **風險管理**：實施多種止損機制和追蹤止損
6. **智能訂單管理**：使用智能訂單取消和替換邏輯

## 主要特性

- **自定義DriftMA指標**：結合加權漂移與平滑的專有動量指標
- **多時間框架分析**：在主間隔和最小間隔上操作以實現精確時機
- **高級止損**：多種止損類型，包括百分比、基於ATR和追蹤止損
- **動態再平衡**：基於趨勢線回歸的自動倉位再平衡
- **智能訂單管理**：基於市場條件的智能訂單取消
- **全面分析**：詳細的性能跟踪和圖表生成
- **高頻交易**：支持極短間隔（1秒）的剝頭皮策略
- **基於波動性的定價**：使用標準差調整訂單價格

## 策略邏輯

### DriftMA指標

<augment_code_snippet path="pkg/strategy/drift/driftma.go" mode="EXCERPT">
```go
type DriftMA struct {
    types.SeriesBase
    drift *indicator.WeightedDrift
    ma1   types.UpdatableSeriesExtend
    ma2   types.UpdatableSeriesExtend
}

func (s *DriftMA) Update(value, weight float64) {
    s.ma1.Update(value)
    if s.ma1.Length() == 0 {
        return
    }
    s.drift.Update(s.ma1.Last(0), weight)
    if s.drift.Length() == 0 {
        return
    }
    s.ma2.Update(s.drift.Last(0))
}
```
</augment_code_snippet>

### 信號生成

<augment_code_snippet path="pkg/strategy/drift/strategy.go" mode="EXCERPT">
```go
drift := s.drift.Array(2)
ddrift := s.drift.drift.Array(2)

shortCondition := drift[1] >= 0 && drift[0] <= 0 || 
                 (drift[1] >= drift[0] && drift[1] <= 0) || 
                 ddrift[1] >= 0 && ddrift[0] <= 0 || 
                 (ddrift[1] >= ddrift[0] && ddrift[1] <= 0)

longCondition := drift[1] <= 0 && drift[0] >= 0 || 
                (drift[1] <= drift[0] && drift[1] >= 0) || 
                ddrift[1] <= 0 && ddrift[0] >= 0 || 
                (ddrift[1] <= ddrift[0] && ddrift[1] >= 0)

// 使用價格方向解決衝突
if shortCondition && longCondition {
    if s.priceLines.Index(1) > s.priceLines.Last(0) {
        longCondition = false
    } else {
        shortCondition = false
    }
}
```
</augment_code_snippet>

### 止損實現

<augment_code_snippet path="pkg/strategy/drift/stoploss.go" mode="EXCERPT">
```go
func (s *Strategy) CheckStopLoss() bool {
    if s.UseStopLoss {
        stoploss := s.StopLoss.Float64()
        if s.sellPrice > 0 && s.sellPrice*(1.+stoploss) <= s.highestPrice ||
           s.buyPrice > 0 && s.buyPrice*(1.-stoploss) >= s.lowestPrice {
            return true
        }
    }
    if s.UseAtr {
        atr := s.atr.Last(0)
        if s.sellPrice > 0 && s.sellPrice+atr <= s.highestPrice ||
           s.buyPrice > 0 && s.buyPrice-atr >= s.lowestPrice {
            return true
        }
    }
    return false
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `interval` | 間隔 | 是 | 主分析間隔（例如："1s"、"1m"） |
| `minInterval` | 間隔 | 是 | 止損檢查的最小間隔 |
| `window` | 整數 | 是 | 主移動平均線的窗口大小 |

### 漂移指標設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `predictOffset` | 整數 | 是 | 線性回歸預測的回望長度 |
| `smootherWindow` | 整數 | 是 | 漂移EWMA平滑的窗口 |
| `fisherTransformWindow` | 整數 | 是 | Fisher變換過濾的窗口 |
| `hlRangeWindow` | 整數 | 是 | 高/低方差計算的窗口 |
| `hlVarianceMultiplier` | 浮點數 | 是 | 基於波動性的價格調整乘數 |

### 風險管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `useStopLoss` | 布爾值 | 否 | 啟用基於百分比的止損 |
| `stoploss` | 百分比 | 否 | 從入場價格的止損百分比 |
| `useAtr` | 布爾值 | 否 | 啟用基於ATR的止損 |
| `atrWindow` | 整數 | 是 | ATR計算的窗口 |
| `noTrailingStopLoss` | 布爾值 | 否 | 禁用追蹤止損 |

### 追蹤止損配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `trailingActivationRatio` | 數組 | 否 | 追蹤止損的激活比率 |
| `trailingCallbackRate` | 數組 | 否 | 追蹤止損的回調率 |

### 訂單管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `pendingMinInterval` | 整數 | 是 | 取消未成交訂單前的時間 |
| `limitOrder` | 布爾值 | 否 | 使用限價單而非市價單 |

### 再平衡
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `noRebalance` | 布爾值 | 否 | 禁用自動再平衡 |
| `trendWindow` | 整數 | 是 | 趨勢線計算的窗口 |
| `rebalanceFilter` | 浮點數 | 是 | 再平衡決策的Beta過濾器 |

### 分析和調試
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `debug` | 布爾值 | 否 | 啟用調試日誌 |
| `generateGraph` | 布爾值 | 否 | 生成性能圖表 |
| `graphPNLDeductFee` | 布爾值 | 否 | 從盈虧圖表中扣除費用 |
| `canvasPath` | 字符串 | 否 | 指標圖表的路徑 |
| `graphPNLPath` | 字符串 | 否 | 盈虧圖表的路徑 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  drift:
    symbol: BTCUSDT
    
    # 時間間隔
    interval: 1s              # 主分析間隔
    minInterval: 1s           # 止損檢查的最小間隔
    window: 2                 # 主MA窗口
    
    # 漂移指標設置
    predictOffset: 2          # 線性回歸回望
    smootherWindow: 10        # EWMA平滑窗口
    fisherTransformWindow: 45 # Fisher變換窗口
    hlRangeWindow: 6          # 高/低方差窗口
    hlVarianceMultiplier: 0.7 # 價格調整乘數
    
    # 風險管理
    useStopLoss: true         # 啟用百分比止損
    stoploss: 0.01%           # 0.01%止損
    useAtr: true              # 啟用ATR止損
    atrWindow: 24             # ATR計算窗口
    noTrailingStopLoss: false # 啟用追蹤止損
    
    # 追蹤止損（多級）
    trailingActivationRatio: [0.0008, 0.002, 0.01]
    trailingCallbackRate: [0.00014, 0.0003, 0.0016]
    
    # 訂單管理
    pendingMinInterval: 6     # 6個間隔後取消訂單
    limitOrder: true          # 使用限價單
    
    # 再平衡
    noRebalance: false        # 啟用再平衡
    trendWindow: 4            # 趨勢線窗口
    rebalanceFilter: 2        # Beta過濾器閾值
    
    # 分析
    debug: false              # 禁用調試日誌
    generateGraph: true       # 生成性能圖表
    graphPNLDeductFee: false  # 不從圖表中扣除費用
    canvasPath: "./output.png"
    graphPNLPath: "./pnl.png"
```

## 策略組件

### 1. DriftMA指標
- **第一次平滑**：對價格源應用EWMA
- **漂移計算**：平滑價格的加權漂移
- **第二次平滑**：對漂移值應用Fisher變換
- **信號生成**：漂移的交叉和方向變化

### 2. 基於波動性的定價
- **高/低跟踪**：使用標準差監控價格波動性
- **動態調整**：根據市場波動性調整訂單價格
- **風險調整入場**：考慮市場條件在最佳價格下單

### 3. 多級追蹤止損
- **激活比率**：追蹤激活前的多個利潤級別
- **回調率**：每個激活級別的不同回調率
- **動態跟踪**：持續更新最高/最低價格進行追蹤

### 4. 智能訂單管理
- **基於時間的取消**：在指定時間間隔後取消訂單
- **基於價格的取消**：當市場大幅移動時取消訂單
- **智能替換**：用更新的價格替換已取消的訂單

## 入場和出場規則

### 入場條件

**多頭入場**：
1. 漂移從負值穿越到正值 或
2. 漂移為正值且增加 或
3. 漂移導數從負值穿越到正值 或
4. 漂移導數為正值且增加
5. 價格根據波動性向下調整以獲得更好入場

**空頭入場**：
1. 漂移從正值穿越到負值 或
2. 漂移為負值且減少 或
3. 漂移導數從正值穿越到負值 或
4. 漂移導數為負值且減少
5. 價格根據波動性向上調整以獲得更好入場

### 出場條件

**止損**：
1. 從入場價格的基於百分比的止損
2. 使用市場波動性的基於ATR的止損
3. 具有不同激活比率的多級追蹤止損

**追蹤止損邏輯**：
- 當利潤超過激活比率時激活
- 以指定的回調率追蹤
- 不同利潤範圍的多個級別

## 風險管理功能

### 1. 多種止損類型
- **百分比止損**：從入場價格的固定百分比
- **ATR止損**：基於市場波動性的動態止損
- **追蹤止損**：保護利潤的追蹤機制

### 2. 動態再平衡
- **趨勢分析**：在趨勢線上使用線性回歸
- **Beta過濾**：僅在趨勢強度顯著變化時再平衡
- **倉位調整**：根據趨勢方向自動調整倉位

### 3. 訂單管理
- **智能取消**：基於時間和價格移動取消訂單
- **波動性調整**：根據市場條件調整訂單價格
- **執行優化**：智能選擇限價單和市價單

## 性能分析

### 實時監控
- **倉位跟踪**：持續監控最高/最低價格
- **性能指標**：實時盈虧計算和跟踪
- **交易統計**：全面的交易分析和報告

### 圖表生成
- **指標圖表**：漂移和其他指標的視覺表示
- **盈虧圖表**：隨時間的利潤/虧損可視化
- **累積性能**：隨時間的資產價值變化
- **執行時間**：策略執行的性能監控

## 常見用例

### 1. 高頻剝頭皮
```yaml
interval: 1s
minInterval: 1s
window: 2
smootherWindow: 5
pendingMinInterval: 3
trailingActivationRatio: [0.0005, 0.001, 0.005]
trailingCallbackRate: [0.0001, 0.0002, 0.001]
```

### 2. 短期動量交易
```yaml
interval: 1m
minInterval: 1s
window: 5
smootherWindow: 15
pendingMinInterval: 10
trailingActivationRatio: [0.002, 0.005, 0.015]
trailingCallbackRate: [0.0005, 0.001, 0.003]
```

### 3. 中期趨勢跟隨
```yaml
interval: 5m
minInterval: 1m
window: 10
smootherWindow: 30
pendingMinInterval: 20
trailingActivationRatio: [0.01, 0.02, 0.05]
trailingCallbackRate: [0.002, 0.005, 0.01]
```

## 最佳實踐

1. **間隔選擇**：剝頭皮使用較短間隔，趨勢跟隨使用較長間隔
2. **波動性調整**：根據市場波動性調整`hlVarianceMultiplier`
3. **止損調整**：結合百分比和ATR止損以實現最佳風險管理
4. **追蹤止損級別**：為不同利潤範圍使用多個級別
5. **再平衡**：在趨勢市場中啟用再平衡，在區間市場中禁用
6. **訂單管理**：根據市場流動性調整`pendingMinInterval`

## 限制

1. **高頻要求**：在低延遲連接下效果最佳
2. **市場依賴性**：在不同市場條件下性能差異顯著
3. **參數敏感性**：需要仔細調整以實現最佳性能
4. **計算密集性**：多個指標和頻繁計算
5. **滑點風險**：高頻交易在波動市場中可能經歷滑點

## 故障排除

### 常見問題

**未生成交易**
- 檢查漂移值是否穿越零點
- 驗證預測偏移是否有足夠數據
- 確保市場波動性足以生成信號

**過度止損**
- 增加止損百分比
- 調整ATR窗口以獲得更平滑的波動性測量
- 檢查追蹤止損激活比率

**訂單成交率差**
- 增加`hlVarianceMultiplier`以獲得更好定價
- 減少`pendingMinInterval`以更快更新訂單
- 在波動條件下考慮使用市價單

**頻繁再平衡**
- 增加`rebalanceFilter`閾值
- 延長`trendWindow`以獲得更穩定的趨勢檢測
- 在區間市場中考慮禁用再平衡

## 高級配置

### 退出方法整合
```yaml
exits:
  - roiStopLoss:
      percentage: 0.35%
  - roiTakeProfit:
      percentage: 0.7%
  - protectiveStopLoss:
      activationRatio: 0.5%
      stopLossRatio: 0.2%
      placeStopOrder: false
```

### 多重追蹤止損
```yaml
# 保守設置
trailingActivationRatio: [0.002, 0.005, 0.02]
trailingCallbackRate: [0.0005, 0.001, 0.005]

# 激進設置
trailingActivationRatio: [0.0005, 0.001, 0.005]
trailingCallbackRate: [0.0001, 0.0002, 0.001]
```

## 性能優化

### 高頻交易
- 使用1秒間隔和最小平滑
- 啟用智能訂單取消
- 為市場條件優化`pendingMinInterval`
- 使用限價單獲得更好定價

### 趨勢跟隨
- 使用較長間隔（5分鐘+）和更多平滑
- 啟用再平衡以適應趨勢
- 使用更寬的追蹤止損級別
- 專注於趨勢強度指標

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/driftBTC.yaml)
- [技術指標指南](../../doc/topics/indicators.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
