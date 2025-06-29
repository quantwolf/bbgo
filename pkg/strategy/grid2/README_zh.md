# Grid2 策略

## 概述

Grid2是一個高級網格交易策略，在預定價格水平放置買賣訂單，從市場波動中獲利。它在價格範圍內創建訂單"網格"，隨著市場振盪自動低買高賣。該策略在橫盤或區間震盪市場中特別有效，價格在定義範圍內波動。

## 工作原理

1. **網格設置**：在上下價格邊界之間創建買賣訂單網格
2. **訂單放置**：在當前價格下方放置買單，上方放置賣單
3. **利潤捕獲**：當訂單成交時，策略放置新訂單以維持網格
4. **複利增長**：可選擇將利潤再投資以隨時間增加倉位規模
5. **風險管理**：包括止損、止盈和倉位管理功能
6. **恢復系統**：可在重啟或中斷後恢復和重建網格

## 主要特性

- **靈活網格配置**：支持算術和幾何網格間距
- **自動範圍檢測**：從歷史數據自動確定價格範圍
- **複利交易**：將利潤再投資以增加網格訂單規模
- **基礎貨幣收益**：選擇以基礎貨幣而非報價貨幣賺取利潤
- **恢復系統**：在重啟或連接問題後恢復網格狀態
- **利潤跟踪**：全面的利潤統計和性能指標
- **風險控制**：止損、止盈和倉位規模管理
- **訂單管理**：具有錯誤處理的高級訂單生命週期管理

## 策略架構

### 網格結構

<augment_code_snippet path="pkg/strategy/grid2/grid.go" mode="EXCERPT">
```go
type Grid struct {
    UpperPrice fixedpoint.Value `json:"upperPrice"`
    LowerPrice fixedpoint.Value `json:"lowerPrice"`
    
    // Size是網格總數
    Size fixedpoint.Value `json:"size"`
    
    // TickSize是價格tick大小，用於截斷價格
    TickSize fixedpoint.Value `json:"tickSize"`
    
    // Spread是不可變數字
    Spread fixedpoint.Value `json:"spread"`
    
    // Pins是固定的網格價格，從低到高
    Pins []Pin `json:"pins"`
}
```
</augment_code_snippet>

### 策略配置

<augment_code_snippet path="pkg/strategy/grid2/strategy.go" mode="EXCERPT">
```go
type Strategy struct {
    Symbol string `json:"symbol"`
    
    // ProfitSpread是您想要提交賣單的固定利潤價差
    ProfitSpread fixedpoint.Value `json:"profitSpread"`
    
    // GridNum是網格數量，您想在訂單簿上發布多少訂單
    GridNum int64 `json:"gridNumber"`
    
    UpperPrice fixedpoint.Value `json:"upperPrice"`
    LowerPrice fixedpoint.Value `json:"lowerPrice"`
    
    // Compound選項用於在賣單成交獲利時購買更多庫存
    Compound bool `json:"compound"`
    
    // EarnBase選項用於以基礎貨幣賺取利潤
    EarnBase bool `json:"earnBase"`
    
    QuoteInvestment fixedpoint.Value `json:"quoteInvestment"`
    BaseInvestment  fixedpoint.Value `json:"baseInvestment"`
}
```
</augment_code_snippet>

### 網格計算

<augment_code_snippet path="pkg/strategy/grid2/grid.go" mode="EXCERPT">
```go
func calculateArithmeticPins(lower, upper, spread, tickSize fixedpoint.Value) []Pin {
    var pins []Pin
    
    // 基於tick大小計算精度
    var ts = tickSize.Float64()
    var prec = int(math.Round(math.Log10(ts) * -1.0))
    
    // 從下限到上限以價差間隔生成pins
    for p := lower; p.Compare(upper.Sub(spread)) <= 0; p = p.Add(spread) {
        price := util.RoundAndTruncatePrice(p, prec)
        pins = append(pins, Pin(price))
    }
    
    // 確保包含上限價格
    upperPrice := util.RoundAndTruncatePrice(upper, prec)
    pins = append(pins, Pin(upperPrice))
    
    return pins
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `symbol` | 字符串 | 是 | 交易對符號（例如："BTCUSDT"） |
| `gridNumber` | 整數 | 是 | 上下價格之間的網格級別數 |
| `upperPrice` | 數字 | 是* | 網格上邊界 |
| `lowerPrice` | 數字 | 是* | 網格下邊界 |

*除非使用`autoRange`，否則必需

### 投資設置（選擇一種）
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `amount` | 數字 | 否 | 每訂單固定報價金額（例如：10 USDT） |
| `quantity` | 數字 | 否 | 每訂單固定基礎數量（例如：0.001 BTC） |
| `quoteInvestment` | 數字 | 否 | 投資的總報價貨幣 |
| `baseInvestment` | 數字 | 否 | 投資的總基礎貨幣 |

### 高級設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `autoRange` | 持續時間 | 否 | 從歷史數據自動檢測價格範圍 |
| `profitSpread` | 數字 | 否 | 自定義利潤價差（覆蓋計算價差） |
| `compound` | 布爾值 | 否 | 將利潤再投資以增加訂單規模 |
| `earnBase` | 布爾值 | 否 | 以基礎貨幣而非報價貨幣賺取利潤 |
| `triggerPrice` | 數字 | 否 | 觸發網格激活的價格水平 |
| `stopLossPrice` | 數字 | 否 | 關閉網格並清算的價格水平 |
| `takeProfitPrice` | 數字 | 否 | 關閉網格並止盈的價格水平 |

### 恢復和管理
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `recoverOrdersWhenStart` | 布爾值 | 否 | 重啟時恢復網格訂單 |
| `keepOrdersWhenShutdown` | 布爾值 | 否 | 關閉時保持訂單活躍 |
| `clearOpenOrdersWhenStart` | 布爾值 | 否 | 啟動時清除現有訂單 |
| `closeWhenCancelOrder` | 布爾值 | 否 | 如果任何訂單被手動取消則關閉網格 |

## 配置示例

### 1. 基本網格交易
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 50
    upperPrice: 50000
    lowerPrice: 30000
    quoteInvestment: 10000
    compound: false
```

### 2. 自動範圍複利網格
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: ETHUSDT
    gridNumber: 100
    autoRange: 14d              # 使用14天價格範圍
    quoteInvestment: 5000
    baseInvestment: 2.0
    compound: true              # 利潤再投資
    earnBase: true              # 賺取ETH而非USDT
```

### 3. 風險管理保守網格
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 30
    upperPrice: 45000
    lowerPrice: 35000
    amount: 100                 # 每訂單固定$100
    stopLossPrice: 32000        # $32,000止損
    takeProfitPrice: 48000      # $48,000止盈
    closeWhenCancelOrder: true  # 手動取消時關閉
```

### 4. 高頻網格
```yaml
exchangeStrategies:
- on: binance
  grid2:
    symbol: BTCUSDT
    gridNumber: 200
    upperPrice: 42000
    lowerPrice: 38000
    quantity: 0.001             # 每訂單固定0.001 BTC
    compound: true
    profitSpread: 50            # 自定義$50利潤價差
```

## 網格類型

### 算術網格
- **等價格間隔**：每個網格級別由相同價格差分隔
- **計算**：`價差 = (上限價格 - 下限價格) / (網格數 - 1)`
- **最適合**：大多數市場條件，更容易理解和配置

### 幾何網格
- **等百分比間隔**：每個網格級別由相同百分比分隔
- **計算**：在價格級別之間使用指數間距
- **最適合**：高波動市場或具有指數增長模式的資產

## 投資模式

### 1. 固定金額模式
```yaml
amount: 100  # 每訂單$100 USDT
```
- 每個訂單使用相同的報價貨幣金額
- 訂單數量隨價格水平變化
- 簡單且可預測的資本分配

### 2. 固定數量模式
```yaml
quantity: 0.01  # 每訂單0.01 BTC
```
- 每個訂單使用相同的基礎貨幣數量
- 訂單價值隨價格水平變化
- 對積累特定數量有用

### 3. 投資模式
```yaml
quoteInvestment: 10000  # $10,000總投資
baseInvestment: 0.5     # 0.5 BTC現有倉位
```
- 策略計算最優訂單規模
- 在所有網格級別分配投資
- 最資本高效的方法

## 自動範圍檢測

### 使用方法
```yaml
autoRange: 14d  # 使用14天價格範圍
```

### 工作原理
1. **歷史分析**：分析指定期間的價格數據
2. **樞軸檢測**：在時間框架內找到樞軸高點和低點
3. **範圍計算**：將上限價格設為樞軸高點，下限價格設為樞軸低點
4. **動態調整**：自動適應市場條件

### 有效格式
- `7d` - 7天
- `2w` - 2週
- `1h` - 1小時
- `30m` - 30分鐘

## 利潤機制

### 標準模式（報價貨幣利潤）
- **低買**：在當前價格下方放置買單
- **高賣**：在當前價格上方放置賣單
- **利潤**：從價差捕獲中賺取報價貨幣（例如USDT）

### 賺取基礎模式
```yaml
earnBase: true
```
- **積累**：專注於積累基礎貨幣（例如BTC）
- **策略**：使用利潤購買更多基礎貨幣
- **長期**：更適合長期資產積累

### 複利模式
```yaml
compound: true
```
- **再投資**：自動將利潤再投資到更大的訂單中
- **增長**：隨著利潤積累，訂單規模隨時間增加
- **指數**：在有利條件下可導致指數增長

## 風險管理

### 止損
```yaml
stopLossPrice: 30000
```
- 如果價格跌破水平，自動關閉網格並清算倉位
- 防止重大市場下跌
- 將所有基礎貨幣轉換為報價貨幣

### 止盈
```yaml
takeProfitPrice: 60000
```
- 如果價格升破水平，自動關閉網格並實現利潤
- 在強勁上升趨勢中鎖定收益
- 維持當前倉位分配

### 倉位限制
- **餘額檢查**：放置訂單前確保餘額充足
- **最小名義**：遵守交易所最小訂單要求
- **風險分配**：限制每個網格級別的最大暴露

## 恢復系統

### 訂單恢復
```yaml
recoverOrdersWhenStart: true
```
- 啟動時掃描現有訂單
- 從活躍訂單重建網格狀態
- 重啟後無縫繼續交易

### 交易恢復
```yaml
recoverGridByScanningTrades: true
recoverGridWithin: 72h
```
- 分析最近交易歷史
- 從已成交訂單重構網格狀態
- 恢復利潤統計和倉位數據

### 錯誤處理
- **連接問題**：自動重連並恢復狀態
- **訂單失敗**：使用指數退避重試失敗訂單
- **餘額不匹配**：調整網格以匹配可用餘額

## 性能優化

### 訂單管理
- **批量操作**：為效率組合訂單提交
- **速率限制**：遵守交易所API速率限制
- **錯誤恢復**：優雅處理臨時失敗

### 內存效率
- **訂單緩存**：緩存活躍訂單以減少API調用
- **狀態持久化**：保存關鍵狀態以在重啟後存活
- **垃圾收集**：定期清理舊數據

## 最佳實踐

1. **範圍選擇**：基於歷史支撐/阻力位選擇範圍
2. **網格密度**：更多網格=更多交易但更高費用
3. **資本分配**：不要投資超過您能承受損失的金額
4. **市場條件**：在區間/橫盤市場中效果最佳
5. **費用考慮**：確保網格價差>2×交易費用
6. **監控**：定期檢查網格性能並根據需要調整

## 常見用例

### 橫盤市場交易
- **場景**：價格在支撐和阻力之間振盪
- **策略**：中等密度的寬網格
- **利潤**：從區間交易中獲得一致利潤

### 波動性收穫
- **場景**：高波動性但無明確趨勢
- **策略**：啟用複利的密集網格
- **利潤**：從價格波動中捕獲利潤

### 定投平均
- **場景**：資產的長期積累
- **策略**：啟用earnBase的網格
- **利潤**：隨時間積累基礎貨幣

### 套利增強
- **場景**：在低波動期間增強回報
- **策略**：小價差的緊密網格
- **利潤**：從微小變動中獲得額外回報

## 限制

1. **趨勢市場**：在強趨勢中可能積累虧損倉位
2. **跳空風險**：無法防止價格跳空或突然變動
3. **費用敏感性**：高交易頻率增加費用成本
4. **資本要求**：需要大量資本才能有效網格
5. **市場風險**：受整體市場方向和波動性影響

## 故障排除

### 常見問題

**網格未啟動**
- 檢查價格範圍有效性（上限>下限）
- 驗證最小訂單的餘額充足
- 確保網格價差>最小利潤閾值

**訂單未成交**
- 檢查網格價格是否具有競爭力
- 驗證網格級別的市場流動性
- 考慮調整網格密度或範圍

**利潤低於預期**
- 檢查交易費用與網格價差
- 檢查市場波動性和交易頻率
- 考慮啟用複利模式

**恢復問題**
- 啟用訂單恢復選項
- 檢查交易歷史權限
- 驗證API密鑰權限

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [配置示例](../../../config/grid2.yaml)
- [網格交易指南](../../doc/topics/grid-trading.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
