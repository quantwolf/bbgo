# 趨勢交易策略

## 概述

趨勢交易策略是一個趨勢跟隨策略，識別並交易動態趨勢線的突破。它使用樞軸高點和樞軸低點構建收斂趨勢線（支撐和阻力），當價格突破這些趨勢線時執行交易，表明潛在的趨勢延續或反轉。

## 工作原理

1. **樞軸點檢測**：使用可配置的窗口期間識別樞軸高點和樞軸低點
2. **趨勢線構建**：從樞軸點構建動態支撐和阻力趨勢線
3. **收斂分析**：檢測支撐和阻力線何時收斂（形成三角形模式）
4. **突破檢測**：監控價格突破阻力上方或支撐下方
5. **倉位管理**：在向上突破時開多倉，在向下突破時開空倉
6. **退出策略**：使用可配置的退出方法，包括追蹤止損進行利潤保護

## 主要特性

- **動態趨勢線**：自動從樞軸點構建趨勢線
- **收斂檢測**：識別收斂趨勢模式以獲得更高概率的設置
- **突破交易**：在確認的趨勢線突破時執行交易
- **靈活的退出方法**：支持多種退出策略，包括追蹤止損
- **倉位反轉**：當出現新信號時自動平倉相反倉位
- **市價訂單執行**：在突破時使用市價訂單立即執行
- **可配置參數**：可調整的樞軸窗口和數量設置

## 策略邏輯

### 樞軸點分析

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
s.pivotHigh = standardIndicator.PivotHigh(types.IntervalWindow{
	Interval: s.Interval,
	Window:   int(3. * s.PivotRightWindow), RightWindow: &s.PivotRightWindow})

s.pivotLow = standardIndicator.PivotLow(types.IntervalWindow{
	Interval: s.Interval,
	Window:   int(3. * s.PivotRightWindow), RightWindow: &s.PivotRightWindow})
```
</augment_code_snippet>

### 趨勢線構建

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
// 從樞軸高點計算阻力斜率
if line(resistancePrices.Index(2), resistancePrices.Index(1), resistancePrices.Index(0)) < 0 {
	resistanceSlope1 = (resistancePrices.Index(1) - resistancePrices.Index(2)) / resistanceDuration.Index(1)
	resistanceSlope2 = (resistancePrices.Index(0) - resistancePrices.Index(1)) / resistanceDuration.Index(0)
	resistanceSlope = (resistanceSlope1 + resistanceSlope2) / 2.
}

// 從樞軸低點計算支撐斜率
if line(supportPrices.Index(2), supportPrices.Index(1), supportPrices.Index(0)) > 0 {
	supportSlope1 = (supportPrices.Index(1) - supportPrices.Index(2)) / supportDuration.Index(1)
	supportSlope2 = (supportPrices.Index(0) - supportPrices.Index(1)) / supportDuration.Index(0)
	supportSlope = (supportSlope1 + supportSlope2) / 2.
}
```
</augment_code_snippet>

### 突破檢測和交易

<augment_code_snippet path="pkg/strategy/trendtrader/trend.go" mode="EXCERPT">
```go
if converge(resistanceSlope, supportSlope) {
	// 計算當前趨勢線水平
	currentResistance := resistanceSlope*pivotHighDurationCounter + resistancePrices.Last(0)
	currentSupport := supportSlope*pivotLowDurationCounter + supportPrices.Last(0)
	
	// 向上突破 - 做多
	if kline.High.Float64() > currentResistance {
		if position.IsShort() {
			s.orderExecutor.ClosePosition(context.Background(), one)
		}
		if position.IsDust(kline.Close) || position.IsClosed() {
			s.placeOrder(context.Background(), types.SideTypeBuy, s.Quantity, symbol)
		}
	
	// 向下突破 - 做空
	} else if kline.Low.Float64() < currentSupport {
		if position.IsLong() {
			s.orderExecutor.ClosePosition(context.Background(), one)
		}
		if position.IsDust(kline.Close) || position.IsClosed() {
			s.placeOrder(context.Background(), types.SideTypeSell, s.Quantity, symbol)
		}
	}
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |

### 趨勢線配置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `trendLine` | 對象 | 是 | 趨勢線檢測設置 |
| `trendLine.interval` | 間隔 | 是 | 分析時間間隔（例如："30m"、"1h"） |
| `trendLine.pivotRightWindow` | 整數 | 是 | 樞軸檢測的右窗口大小 |
| `trendLine.quantity` | 數字 | 是 | 每筆交易的固定數量 |
| `trendLine.marketOrder` | 布爾值 | 否 | 啟用市價訂單執行（默認：true） |

### 退出方法
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `exits` | 數組 | 否 | 退出方法配置列表 |
| `exits[].trailingStop` | 對象 | 否 | 追蹤止損配置 |
| `exits[].trailingStop.callbackRate` | 百分比 | 否 | 追蹤止損的回調率 |
| `exits[].trailingStop.activationRatio` | 百分比 | 否 | 追蹤止損的激活比率 |
| `exits[].trailingStop.closePosition` | 百分比 | 否 | 要平倉的倉位百分比 |
| `exits[].trailingStop.minProfit` | 百分比 | 否 | 激活前的最小利潤 |
| `exits[].trailingStop.interval` | 間隔 | 否 | 追蹤止損的更新間隔 |
| `exits[].trailingStop.side` | 字符串 | 否 | 追蹤止損的方向（"buy"或"sell"） |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  trendtrader:
    symbol: BTCUSDT
    
    # 趨勢線配置
    trendLine:
      interval: 30m              # 30分鐘分析間隔
      pivotRightWindow: 40       # 樞軸的40期右窗口
      quantity: 0.001            # 每筆交易0.001 BTC
      marketOrder: true          # 使用市價訂單執行
    
    # 退出策略
    exits:
      # 多倉追蹤止損
      - trailingStop:
          callbackRate: 1%       # 1%回調率
          activationRatio: 1%    # 1%利潤後激活
          closePosition: 100%    # 平倉整個倉位
          minProfit: 15%         # 需要最少15%利潤
          interval: 1m           # 每分鐘更新
          side: buy              # 用於多倉
      
      # 空倉追蹤止損
      - trailingStop:
          callbackRate: 1%       # 1%回調率
          activationRatio: 1%    # 1%利潤後激活
          closePosition: 100%    # 平倉整個倉位
          minProfit: 15%         # 需要最少15%利潤
          interval: 1m           # 每分鐘更新
          side: sell             # 用於空倉
```

## 策略組件

### 1. 樞軸點檢測
- **目的**：識別重要的價格轉折點
- **方法**：使用可配置的右窗口確認樞軸有效性
- **輸出**：用於趨勢線構建的樞軸高點和樞軸低點序列

### 2. 趨勢線構建
- **阻力線**：從下降的樞軸高點構建
- **支撐線**：從上升的樞軸低點構建
- **斜率計算**：使用最近樞軸斜率的時間加權平均

### 3. 收斂檢測
- **條件**：支撐斜率 > 阻力斜率
- **意義**：表明收斂趨勢線形成三角形模式
- **交易含義**：更高概率的突破設置

### 4. 突破執行
- **向上突破**：價格高點超過當前阻力水平
- **向下突破**：價格低點跌破當前支撐水平
- **執行**：市價訂單立即進入倉位

## 數學基礎

### 線方向函數
```go
func line(p1, p2, p3 float64) int64 {
	if p1 >= p2 && p2 >= p3 {
		return -1  // 下降線
	} else if p1 <= p2 && p2 <= p3 {
		return +1  // 上升線
	}
	return 0       // 無明確方向
}
```

### 收斂函數
```go
func converge(mr, ms float64) bool {
	return ms > mr  // 支撐斜率 > 阻力斜率
}
```

### 當前趨勢線水平
```
當前阻力 = 阻力斜率 × 自上次樞軸以來的時間 + 上次樞軸高點
當前支撐 = 支撐斜率 × 自上次樞軸以來的時間 + 上次樞軸低點
```

## 風險管理功能

### 1. 倉位反轉
- 當出現新信號時自動平倉相反倉位
- 防止在波動市場中出現衝突倉位
- 確保乾淨的倉位管理

### 2. 塵埃倉位處理
- 在開新交易前檢查塵埃倉位
- 防止不必要的小倉位累積
- 維持乾淨的倉位規模

### 3. 退出方法整合
- 同時支持多種退出策略
- 追蹤止損進行利潤保護
- 可配置的利潤目標和止損

## 常見用例

### 1. 波段交易設置
```yaml
trendLine:
  interval: 4h
  pivotRightWindow: 20
  quantity: 0.01
exits:
  - trailingStop:
      callbackRate: 2%
      minProfit: 10%
```

### 2. 日內交易設置
```yaml
trendLine:
  interval: 15m
  pivotRightWindow: 10
  quantity: 0.005
exits:
  - trailingStop:
      callbackRate: 0.5%
      minProfit: 5%
```

### 3. 保守設置
```yaml
trendLine:
  interval: 1d
  pivotRightWindow: 50
  quantity: 0.002
exits:
  - trailingStop:
      callbackRate: 3%
      minProfit: 20%
```

## 最佳實踐

1. **時間框架選擇**：選擇與您的交易風格匹配的間隔
2. **樞軸窗口調整**：更大的窗口用於更重要的樞軸，更小的用於響應性
3. **數量管理**：從較小的數量開始測試策略性能
4. **退出策略**：始終配置適當的退出方法進行風險管理
5. **市場條件**：在具有明確突破模式的趨勢市場中效果最佳
6. **回測**：在實盤交易前測試不同的參數組合

## 限制

1. **橫盤市場**：在區間震盪市場中可能產生虛假信號
2. **鋸齒風險**：快速反轉可能導致多次小額虧損
3. **滯後性**：基於樞軸的方法在信號生成中具有固有滯後
4. **市價訂單**：在波動期間可能經歷滑點

## 故障排除

### 常見問題

**未生成交易**
- 檢查樞軸窗口是否適合時間框架
- 驗證是否滿足收斂條件
- 確保有足夠的價格波動進行突破

**過度交易**
- 增加樞軸右窗口以獲得更重要的樞軸
- 添加最小利潤要求
- 考慮更長的時間框架

**性能不佳**
- 檢查退出策略設置
- 調整追蹤止損參數
- 考慮市場條件和波動性

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/trendtrader.yaml)
- [超級趨勢策略](../supertrend/README.md) - 替代趨勢跟隨方法
- [退出方法文檔](../../doc/topics/exit-methods.md)
