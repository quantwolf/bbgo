# ATR Pin 策略

## 概述

ATR Pin策略是一個做市策略，使用平均真實範圍（ATR）指標在當前市場價格周圍動態放置買賣訂單。它根據波動性在距離市場價格的計算距離處"固定"訂單，創建適應市場條件的價差。該策略旨在通過價格振盪獲利，同時通過基於波動性的定位管理風險。

## 工作原理

1. **ATR計算**：在指定窗口內計算平均真實範圍以測量市場波動性
2. **價格範圍確定**：將ATR乘以可配置的乘數來確定訂單放置距離
3. **動態訂單放置**：在當前市場價格下方放置買單，上方放置賣單
4. **倉位管理**：持有倉位時自動放置止盈訂單
5. **風險保護**：實施最小價格範圍保護和塵埃數量過濾
6. **持續調整**：在每個間隔取消並重新放置訂單以維持最佳定位

## 主要特性

- **波動性自適應定位**：使用ATR根據市場波動性調整訂單放置
- **動態做市**：持續在市場價格周圍放置買賣訂單
- **自動止盈**：持有倉位時放置即時止盈訂單
- **風險保護**：多重保護措施，包括最小價格範圍和餘額驗證
- **塵埃數量過濾**：防止放置低於交易所最小值的訂單
- **靈活倉位規模**：支持固定數量和基於百分比的規模
- **基於餘額的止盈**：可選功能，根據預期餘額管理倉位

## 策略邏輯

### 基於ATR的價格範圍計算

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// 計算ATR並應用乘數
lastAtr := atr.Last(0)

// 保護：確保ATR至少是當前蠟燭範圍
if lastAtr <= k.High.Sub(k.Low).Float64() {
    lastAtr = k.High.Sub(k.Low).Float64()
}

priceRange := fixedpoint.NewFromFloat(lastAtr * s.Multiplier)

// 應用最小價格範圍保護
priceRange = fixedpoint.Max(priceRange, k.Close.Mul(s.MinPriceRange))
```
</augment_code_snippet>

### 動態訂單放置

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// 基於當前報價和價格範圍計算買賣價格
bidPrice := fixedpoint.Max(ticker.Buy.Sub(priceRange), s.Market.TickSize)
askPrice := ticker.Sell.Add(priceRange)

// 計算每個訂單的數量
bidQuantity := s.QuantityOrAmount.CalculateQuantity(bidPrice)
askQuantity := s.QuantityOrAmount.CalculateQuantity(askPrice)
```
</augment_code_snippet>

### 倉位管理和止盈

<augment_code_snippet path="pkg/strategy/atrpin/strategy.go" mode="EXCERPT">
```go
// 檢查是否有需要止盈的倉位
position := s.Strategy.OrderExecutor.Position()
base := position.GetBase()

// 可選：使用預期基礎餘額進行倉位計算
if s.TakeProfitByExpectedBaseBalance {
    base = baseBalance.Available.Sub(s.ExpectedBaseBalance)
}

// 確定止盈方向和價格
side := types.SideTypeSell
takerPrice := ticker.Buy
if base.Sign() < 0 {
    side = types.SideTypeBuy
    takerPrice = ticker.Sell
}

// 如果倉位不是塵埃，放置止盈訂單
positionQuantity := base.Abs()
if !s.Market.IsDustQuantity(positionQuantity, takerPrice) {
    // 在當前市場價格提交止盈訂單
    orderForms = append(orderForms, types.SubmitOrder{
        Symbol:      s.Symbol,
        Type:        types.OrderTypeLimit,
        Side:        side,
        Price:       takerPrice,
        Quantity:    positionQuantity,
        Tag:         "takeProfit",
    })
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `interval` | 間隔 | 是 | ATR計算的時間間隔（例如："5m"、"1h"） |
| `window` | 整數 | 是 | ATR計算窗口（週期數） |
| `multiplier` | 浮點數 | 是 | 應用於ATR的乘數用於價格範圍計算 |
| `minPriceRange` | 百分比 | 是 | 作為當前價格百分比的最小價格範圍 |

### 倉位規模
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `quantity` | 數字 | 否 | 每個訂單的固定數量 |
| `amount` | 數字 | 否 | 每個訂單的固定報價金額 |

### 高級設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `takeProfitByExpectedBaseBalance` | 布爾值 | 否 | 使用預期基礎餘額進行倉位計算 |
| `expectedBaseBalance` | 數字 | 否 | 倉位管理的預期基礎貨幣餘額 |

## 配置示例

```yaml
exchangeStrategies:
- on: binance
  atrpin:
    symbol: BTCUSDT
    
    # ATR計算設置
    interval: 5m              # ATR的5分鐘間隔
    window: 14                # 14週期ATR窗口
    multiplier: 10.0          # 價格範圍的10倍ATR乘數
    minPriceRange: 0.5%       # 最小0.5%價格範圍保護
    
    # 倉位規模
    amount: 100               # 每訂單$100
    
    # 高級倉位管理（可選）
    takeProfitByExpectedBaseBalance: false
    expectedBaseBalance: 0.0
```

## ATR（平均真實範圍）解釋

### 什麼是ATR？
ATR通過計算指定期間內真實範圍的平均值來測量市場波動性。真實範圍是以下的最大值：
- 當前最高價 - 當前最低價
- |當前最高價 - 前一收盤價|
- |當前最低價 - 前一收盤價|

### ATR Pin如何使用ATR
1. **波動性測量**：ATR量化當前市場波動性
2. **動態間距**：更高的ATR = 更寬的訂單間距，更低的ATR = 更緊的間距
3. **風險適應**：自動適應變化的市場條件
4. **利潤優化**：平衡捕獲率與每筆交易利潤

### ATR乘數影響
- **低乘數（1-5）**：緊密價差，更高成交率，每筆交易利潤較低
- **中等乘數（5-15）**：大多數市場條件的平衡方法
- **高乘數（15+）**：寬價差，較低成交率，每筆交易利潤較高

## 策略工作流程

### 1. 市場數據處理
- 訂閱K線數據進行ATR計算
- 訂閱1分鐘數據進行響應式訂單管理
- 使用指定窗口和間隔計算ATR

### 2. 訂單管理週期
1. **取消現有訂單**：優雅地取消所有開放訂單
2. **更新帳戶**：刷新餘額信息
3. **計算ATR**：獲取帶保護的最新ATR值
4. **確定價格範圍**：應用乘數和最小範圍保護
5. **查詢市場**：獲取當前買賣價格
6. **檢查倉位**：評估當前倉位以進行止盈

### 3. 訂單放置邏輯
- **止盈優先**：如果存在倉位，首先放置止盈訂單
- **做市訂單**：如果沒有倉位或倉位是塵埃，放置買賣訂單
- **餘額驗證**：確保每個訂單有足夠的餘額
- **塵埃過濾**：防止低於交易所最小值的訂單

### 4. 風險管理
- **最小價格範圍**：防止極低波動性
- **餘額檢查**：訂單放置前驗證可用餘額
- **塵埃數量保護**：過濾低於最小閾值的訂單
- **Tick大小合規**：確保買價滿足最小tick要求

## 價格範圍保護

### ATR保護
```go
// 確保ATR至少是當前蠟燭範圍
if lastAtr <= k.High.Sub(k.Low).Float64() {
    lastAtr = k.High.Sub(k.Low).Float64()
}
```

### 最小範圍保護
```go
// 應用最小價格範圍（例如，當前價格的0.5%）
priceRange = fixedpoint.Max(priceRange, k.Close.Mul(s.MinPriceRange))
```

### Tick大小保護
```go
// 確保買價至少是一個tick大小
bidPrice := fixedpoint.Max(ticker.Buy.Sub(priceRange), s.Market.TickSize)
```

## 倉位管理模式

### 標準模式
使用策略的內部倉位跟踪：
```yaml
takeProfitByExpectedBaseBalance: false
```

### 預期餘額模式
使用預期基礎餘額進行倉位計算：
```yaml
takeProfitByExpectedBaseBalance: true
expectedBaseBalance: 1.0  # 預期1.0 BTC餘額
```

此模式對於處理缺失交易或外部倉位變化很有用。

## 性能優化

### 間隔選擇
- **1m**：非常響應，較高交易成本
- **5m**：響應性和效率的良好平衡
- **15m**：較低頻率，適合波動性較小的市場
- **1h**：長期定位，最小交易成本

### 窗口大小調整
- **小窗口（7-10）**：對最近波動性變化更響應
- **中等窗口（14-21）**：標準ATR計算，平衡方法
- **大窗口（30+）**：更平滑的ATR，對短期峰值不太敏感

### 乘數優化
- **市場條件**：根據當前波動性制度調整
- **價差競爭**：考慮交易所的典型價差
- **風險容忍度**：更高乘數用於更保守的方法

## 常見用例

### 1. 高頻做市
```yaml
interval: 1m
window: 7
multiplier: 5.0
minPriceRange: 0.1%
amount: 50
```

### 2. 中頻平衡
```yaml
interval: 5m
window: 14
multiplier: 10.0
minPriceRange: 0.5%
amount: 100
```

### 3. 保守長期
```yaml
interval: 1h
window: 21
multiplier: 20.0
minPriceRange: 1.0%
amount: 200
```

## 最佳實踐

1. **ATR週期選擇**：使用14週期作為起點，根據市場特徵調整
2. **乘數調整**：從10倍開始，增加以獲得更保守的方法
3. **最小範圍保護**：設置為0.5-1%以處理低波動期
4. **倉位規模**：使用基於金額的規模以獲得一致的風險暴露
5. **市場選擇**：在具有一致波動性的流動市場中效果最佳
6. **監控**：定期檢查成交率和盈利能力

## 限制

1. **趨勢市場**：在強趨勢中可能累積倉位
2. **低波動性**：在安靜期間減少利潤機會
3. **高波動性**：寬價差可能降低成交率
4. **市場跳空**：無法防止隔夜跳空或突然移動
5. **交易所費用**：高頻交易可能產生大量費用
6. **滑點**：止盈訂單使用限價單，可能無法立即成交

## 故障排除

### 常見問題

**未放置訂單**
- 檢查最小價格範圍設置
- 驗證帳戶餘額充足
- 確保ATR計算有足夠數據

**訂單未成交**
- 減少ATR乘數以獲得更緊價差
- 檢查市場流動性和典型價差
- 驗證價格範圍計算

**過度倉位累積**
- 啟用按預期餘額止盈
- 減少倉位規模
- 實施額外退出策略

**ATR計算錯誤**
- 確保有足夠的歷史數據
- 檢查間隔和窗口設置
- 驗證市場數據連接

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/atrpin.yaml)
- [ATR指標指南](../../doc/topics/atr-indicator.md)
- [做市策略](../../doc/topics/market-making.md)
