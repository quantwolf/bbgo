# 因子動物園策略

## 概述

因子動物園策略是一個高級量化交易策略，使用機器學習技術實現多因子模型。它結合多個金融因子（價量背離、動量、均值回歸、漂移和成交量動量）與邏輯回歸來預測未來價格走勢並做出交易決策。該策略代表了基於量化金融學術研究的複雜算法交易方法。

## 工作原理

1. **因子計算**：從市場數據計算多個金融因子
2. **特徵工程**：將原始市場數據轉換為預測特徵
3. **機器學習**：使用邏輯回歸建模因子與收益之間的關係
4. **信號生成**：基於模型預測生成二元交易信號（買入/賣出）
5. **倉位管理**：基於預測概率和滾動平均執行交易
6. **風險控制**：與退出方法整合進行全面風險管理

## 主要特性

- **多因子模型**：結合5個不同的金融因子進行全面市場分析
- **機器學習整合**：使用邏輯回歸進行預測建模
- **實時適應**：持續用新市場數據更新模型
- **二元分類**：將連續預測轉換為可操作的交易信號
- **滾動預測**：使用時間序列滾動平均進行信號平滑
- **退出方法整合**：與各種風險管理策略兼容
- **學術基礎**：基於既定的量化金融研究

## 策略架構

### 因子組件

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
type Linear struct {
    // Xs（輸入），因子和指標
    divergence *factorzoo.PVD   // 價量背離
    reversion  *factorzoo.PMR   // 價格均值回歸
    momentum   *factorzoo.MOM   // 來自論文的價格動量，alpha 101
    drift      *indicator.Drift // GBM（幾何布朗運動）
    volume     *factorzoo.VMOM  // 季度成交量動量

    // Y（輸出），內部收益率
    irr *factorzoo.RR
}
```
</augment_code_snippet>

### 機器學習管道

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
// 準備特徵矩陣（X）和目標向量（Y）
a := []floats.Slice{
    s.divergence.Values[len(s.divergence.Values)-s.Window-2 : len(s.divergence.Values)-2],
    s.reversion.Values[len(s.reversion.Values)-s.Window-2 : len(s.reversion.Values)-2],
    s.drift.Values[len(s.drift.Values)-s.Window-2 : len(s.drift.Values)-2],
    s.momentum.Values[len(s.momentum.Values)-s.Window-2 : len(s.momentum.Values)-2],
    s.volume.Values[len(s.volume.Values)-s.Window-2 : len(s.volume.Values)-2],
}

// 二元目標：將收益轉換為0/1分類
b := []floats.Slice{filter(s.irr.Values[len(s.irr.Values)-s.Window-1:len(s.irr.Values)-1], binary)}

// 訓練邏輯回歸模型
model := types.LogisticRegression(x, y[0], s.Window, 8000, 0.0001)

// 使用當前因子值進行預測
input := []float64{
    s.divergence.Last(0),
    s.reversion.Last(0),
    s.drift.Last(0),
    s.momentum.Last(0),
    s.volume.Last(0),
}
pred := model.Predict(input)
```
</augment_code_snippet>

### 交易邏輯

<augment_code_snippet path="pkg/strategy/factorzoo/linear_regression.go" mode="EXCERPT">
```go
// 使用預測的滾動平均進行信號平滑
predLst.Update(pred)

// 基於預測與滾動均值的交易決策
if pred > predLst.Mean() {
    if position.IsShort() {
        s.ClosePosition(ctx, one)
        s.placeMarketOrder(ctx, types.SideTypeBuy, qty, symbol)
    } else if position.IsClosed() {
        s.placeMarketOrder(ctx, types.SideTypeBuy, qty, symbol)
    }
} else if pred < predLst.Mean() {
    if position.IsLong() {
        s.ClosePosition(ctx, one)
        s.placeMarketOrder(ctx, types.SideTypeSell, qty, symbol)
    } else if position.IsClosed() {
        s.placeMarketOrder(ctx, types.SideTypeSell, qty, symbol)
    }
}
```
</augment_code_snippet>

## 金融因子解釋

### 1. 價量背離（PVD）

<augment_code_snippet path="pkg/strategy/factorzoo/factors/price_volume_divergence.go" mode="EXCERPT">
```go
// 測量價格和成交量之間的背離
// 負相關表示背離
func (inc *PVD) Update(price float64, volume float64) {
    inc.Prices.Update(price)
    inc.Volumes.Update(volume)
    if inc.Prices.Length() >= inc.Window && inc.Volumes.Length() >= inc.Window {
        divergence := -types.Correlation(inc.Prices, inc.Volumes, inc.Window)
        inc.Values.Push(divergence)
    }
}
```
</augment_code_snippet>

**理論**：當價格和成交量朝相反方向移動時，通常預示潛在反轉。高負相關表明強烈背離。

### 2. 價格均值回歸（PMR）

<augment_code_snippet path="pkg/strategy/factorzoo/factors/price_mean_reversion.go" mode="EXCERPT">
```go
// 測量價格回歸到移動平均線的趨勢
func (inc *PMR) Update(price float64) {
    inc.SMA.Update(price)
    if inc.SMA.Length() >= inc.Window {
        reversion := inc.SMA.Last(0) / price  // SMA/價格比率
        inc.Values.Push(reversion)
    }
}
```
</augment_code_snippet>

**理論**：假設價格傾向於回歸到移動平均線。值>1表示價格低於平均（潛在買入），值<1表示價格高於平均（潛在賣出）。

### 3. 動量（MOM）

<augment_code_snippet path="pkg/strategy/factorzoo/factors/momentum.go" mode="EXCERPT">
```go
// 跳空動量 - 測量開盤價跳空
func (inc *MOM) Update(open, close float64) {
    inc.opens.Update(open)
    inc.closes.Update(close)
    if inc.opens.Length() >= inc.Window && inc.closes.Length() >= inc.Window {
        gap := inc.opens.Last(0)/inc.closes.Index(1) - 1  // 跳空比率
        inc.Values.Push(gap)
    }
}
```
</augment_code_snippet>

**理論**：當前開盤價與前一收盤價之間的大跳空表明強勁動量。正跳空表明看漲動量，負跳空表明看跌動量。

### 4. 漂移（幾何布朗運動）

使用標準BBGO漂移指標測量價格走勢中的潛在趨勢方向和強度。

### 5. 成交量動量（VMOM）

測量季度期間交易量的動量，表明機構興趣和市場參與度變化。

### 6. 收益率（RR）- 目標變量

<augment_code_snippet path="pkg/strategy/factorzoo/factors/return_rate.go" mode="EXCERPT">
```go
// 簡單收益率計算
func (inc *RR) Update(price float64) {
    inc.prices.Update(price)
    irr := inc.prices.Last(0)/inc.prices.Index(1) - 1  // 期間收益
    inc.Values.Push(irr)
}
```
</augment_code_snippet>

**目的**：作為機器學習模型的目標變量（Y），轉換為二元分類（正收益=1，負收益=0）。

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `linear.interval` | 間隔 | 是 | 分析時間間隔（例如："1d"、"4h"） |
| `linear.window` | 整數 | 是 | 因子計算和模型訓練的回望窗口 |
| `linear.quantity` | 數字 | 是 | 每筆交易的固定數量 |

### 高級設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `linear.marketOrder` | 布爾值 | 否 | 啟用市價單執行（默認：true） |
| `linear.stopEMARange` | 數字 | 否 | 基於EMA的止損範圍 |
| `linear.stopEMA` | 對象 | 否 | 止損的EMA配置 |

### 退出方法
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `exits` | 數組 | 否 | 退出方法配置數組 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  factorzoo:
    symbol: BTCUSDT
    
    # 線性回歸模型配置
    linear:
      interval: 1d              # 每日分析
      window: 5                 # 5天回望窗口
      quantity: 0.01            # 每筆交易0.01 BTC
      marketOrder: true         # 使用市價單
      
    # 風險管理
    exits:
    - trailingStop:
        callbackRate: 1%        # 1%追蹤止損
        activationRatio: 1%     # 1%利潤後激活
        closePosition: 100%     # 平倉整個倉位
        minProfit: 15%          # 激活前最小15%利潤
        interval: 1m            # 每分鐘檢查
        side: buy               # 用於多頭倉位
    - trailingStop:
        callbackRate: 1%
        activationRatio: 1%
        closePosition: 100%
        minProfit: 15%
        interval: 1m
        side: sell              # 用於空頭倉位
```

## 因子配置詳情

### 默認因子窗口
- **PVD（價量背離）**：60個週期
- **PMR（價格均值回歸）**：60個週期
- **MOM（動量）**：1個週期（跳空檢測）
- **Drift（漂移）**：7個週期
- **VMOM（成交量動量）**：90個週期
- **RR（收益率）**：2個週期

### 模型參數
- **訓練迭代次數**：8000
- **學習率**：0.0001
- **分類閾值**：預測的滾動均值

## 機器學習詳情

### 邏輯回歸模型
策略使用邏輯回歸建模正收益的概率：

```
P(收益 > 0) = 1 / (1 + e^(-(β₀ + β₁×PVD + β₂×PMR + β₃×MOM + β₄×Drift + β₅×VMOM)))
```

### 特徵工程
1. **標準化**：因子計算為標準化值
2. **窗口化**：使用滾動窗口捕獲時間模式
3. **二元目標**：將連續收益轉換為二元分類
4. **滯後結構**：使用滯後因子值預測未來收益

### 模型訓練
- **在線學習**：模型在每個新數據點重新訓練
- **滾動窗口**：使用固定窗口大小的訓練數據
- **梯度下降**：迭代優化模型參數

## 交易邏輯

### 信號生成
1. **因子計算**：計算當前期間的所有5個因子
2. **模型預測**：生成概率預測（0-1尺度）
3. **信號平滑**：將預測與滾動平均比較
4. **決策制定**：基於預測與閾值進行交易

### 倉位管理
- **多頭信號**：預測 > 滾動均值 → 買入
- **空頭信號**：預測 < 滾動均值 → 賣出
- **倉位反轉**：自動平倉相反倉位
- **市價單**：使用市價單立即執行

### 風險管理
- **退出方法**：與追蹤止損和其他退出策略整合
- **倉位規模**：每筆交易固定數量
- **訂單取消**：新交易前優雅取消訂單

## 性能考慮

### 計算複雜度
- **因子計算**：每個因子每期O(n)
- **模型訓練**：O(n×m×i)，其中n=樣本，m=特徵，i=迭代
- **內存使用**：存儲因子值的滾動窗口

### 市場條件
- **趨勢市場**：動量和漂移因子表現良好
- **橫盤市場**：均值回歸因子提供價值
- **高波動性**：成交量背離因子捕獲制度變化
- **低流動性**：可能需要更大窗口以獲得穩定信號

## 常見用例

### 1. 每日系統交易
```yaml
linear:
  interval: 1d
  window: 10
  quantity: 0.1
```

### 2. 日內因子交易
```yaml
linear:
  interval: 4h
  window: 20
  quantity: 0.05
```

### 3. 保守長期
```yaml
linear:
  interval: 1w
  window: 4
  quantity: 0.2
exits:
- roiStopLoss:
    percentage: 0.1
```

## 最佳實踐

1. **窗口選擇**：更大窗口提供穩定性，更小窗口提供響應性
2. **因子驗證**：監控個別因子性能
3. **模型監控**：隨時間跟踪預測準確性
4. **風險管理**：始終使用退出方法進行下行保護
5. **市場制度**：考慮不同市場條件的不同參數
6. **回測**：實盤部署前進行廣泛歷史測試

## 限制

1. **模型假設**：假設因子與收益之間存在線性關係
2. **過擬合風險**：小窗口可能導致過擬合
3. **因子穩定性**：因子有效性可能隨時間變化
4. **市場制度變化**：模型可能無法快速適應新制度
5. **交易成本**：高頻重新訓練可能增加成本
6. **數據要求**：需要充足的歷史數據進行訓練

## 故障排除

### 常見問題

**預測準確性差**
- 增加窗口大小以獲得更穩定的訓練
- 檢查因子相關性和多重共線性
- 驗證數據質量和完整性

**過度交易**
- 增加預測閾值敏感性
- 使用更長間隔進行信號生成
- 添加最小持倉期限制

**因子計算錯誤**
- 驗證因子窗口的充足數據
- 檢查缺失或無效的價格/成交量數據
- 監控因子值範圍的異常

**模型訓練失敗**
- 降低學習率以實現收斂
- 增加迭代次數以處理複雜模式
- 檢查數值穩定性問題

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/factorzoo.yaml)
- [量化因子模型](../../doc/topics/factor-models.md)
- [交易中的機器學習](../../doc/topics/ml-trading.md)
