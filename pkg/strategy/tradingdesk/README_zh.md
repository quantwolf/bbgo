# 交易台策略

## 概述

交易台策略是一個手動交易助手，基於風險管理原則提供智能倉位規模計算。它充當"交易台"的角色，通過根據止損水平和最大虧損限制自動計算最優倉位規模，幫助交易者執行具有適當風險控制的手動交易。

## 工作原理

1. **手動交易執行**：為具有風險控制的手動交易執行提供接口
2. **基於風險的倉位規模計算**：根據止損距離和最大虧損容忍度自動計算倉位規模
3. **餘額驗證**：在執行交易前確保賬戶餘額充足
4. **價格類型支持**：支持做市商和吃單者定價進行倉位規模計算
5. **多資產支持**：適用於交易所支持的任何交易對

## 主要特性

- **智能倉位規模計算**：基於風險參數計算最優倉位規模
- **風險管理**：強制執行每筆交易的最大虧損限制
- **餘額保護**：在交易執行前驗證賬戶餘額
- **止損整合**：使用止損價格進行風險計算
- **靈活定價**：支持做市商/吃單者定價模型
- **手動控制**：專為具有自動風險控制的自主交易而設計

## 策略邏輯

### 倉位規模計算

<augment_code_snippet path="pkg/strategy/tradingdesk/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) calculatePositionSize(ctx context.Context, param OpenPositionParam) (fixedpoint.Value, error) {
	// 如果未設置止損，返回原始數量
	if param.StopLossPrice.IsZero() {
		return param.Quantity, nil
	}

	// 獲取當前市場價格
	ticker, err := s.Session.Exchange.QueryTicker(ctx, param.Symbol)
	if err != nil {
		return fixedpoint.Zero, fmt.Errorf("failed to query ticker for %s: %w", param.Symbol, err)
	}

	// 根據方向和價格類型計算入場價格
	var entryPrice fixedpoint.Value
	if param.Side == types.SideTypeBuy {
		entryPrice = ticker.Buy // 買入的賣價
	} else {
		entryPrice = ticker.Sell // 賣出的買價
	}

	// 計算每單位風險
	var riskPerUnit fixedpoint.Value
	if param.Side == types.SideTypeBuy {
		if param.StopLossPrice.Compare(entryPrice) >= 0 {
			return fixedpoint.Zero, fmt.Errorf("invalid stop loss price for buy order")
		}
		riskPerUnit = entryPrice.Sub(param.StopLossPrice)
	} else {
		if param.StopLossPrice.Compare(entryPrice) <= 0 {
			return fixedpoint.Zero, fmt.Errorf("invalid stop loss price for sell order")
		}
		riskPerUnit = param.StopLossPrice.Sub(entryPrice)
	}

	// 基於風險計算最大數量
	var maxQuantityByRisk fixedpoint.Value
	if s.MaxLossLimit.Sign() > 0 && riskPerUnit.Sign() > 0 {
		maxQuantityByRisk = s.MaxLossLimit.Div(riskPerUnit)
	} else {
		maxQuantityByRisk = param.Quantity
	}

	// 基於可用餘額計算最大數量
	maxQuantityByBalance := s.calculateMaxQuantityByBalance(param.Symbol, param.Side, entryPrice)

	// 返回請求、風險限制和餘額限制數量的最小值
	return fixedpoint.Min(param.Quantity, fixedpoint.Min(maxQuantityByRisk, maxQuantityByBalance)), nil
}
```
</augment_code_snippet>

### 餘額計算

<augment_code_snippet path="pkg/strategy/tradingdesk/strategy.go" mode="EXCERPT">
```go
func (s *Strategy) calculateMaxQuantityByBalance(symbol string, side types.SideType, price fixedpoint.Value) fixedpoint.Value {
	market, ok := s.Session.Market(symbol)
	if !ok {
		return fixedpoint.Zero
	}

	account := s.Session.GetAccount()
	if side == types.SideTypeBuy {
		// 對於買單，檢查報價貨幣餘額
		balance := account.Balance(market.QuoteCurrency)
		if balance.Available.Sign() <= 0 || price.Sign() <= 0 {
			return fixedpoint.Zero
		}
		return balance.Available.Div(price)
	} else {
		// 對於賣單，檢查基礎貨幣餘額
		balance := account.Balance(market.BaseCurrency)
		return balance.Available
	}
}
```
</augment_code_snippet>

## 配置參數

### 基本設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `maxLossLimit` | 數字 | 否 | 每筆交易的最大虧損限制（以報價貨幣計） |
| `priceType` | 字符串 | 否 | 計算的價格類型（"maker"或"taker"） |

### 高級設置
| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `session` | 字符串 | 否 | 交易所會話名稱（默認為主會話） |

## 配置示例

```yaml
exchangeStrategies:
  - on: binance
    tradingdesk:
      maxLossLimit: 100.0        # 每筆交易最大虧損$100
      priceType: maker           # 使用做市商定價進行計算

  # 具有更高風險容忍度的替代配置
  - on: max
    tradingdesk:
      maxLossLimit: 500.0        # 每筆交易最大虧損$500
      priceType: taker           # 使用吃單者定價進行計算
```

## 使用模式

### 手動交易執行

交易台策略設計用於程序化使用或通過交易界面使用。以下是倉位規模計算的工作方式：

```go
// Go代碼中的使用示例
param := OpenPositionParam{
    Symbol:        "BTCUSDT",
    Side:          types.SideTypeBuy,
    Quantity:      fixedpoint.NewFromFloat(1.0),    // 想要買入1 BTC
    StopLossPrice: fixedpoint.NewFromFloat(45000),  // 止損在$45,000
}

// 策略計算最優倉位規模
optimalSize, err := strategy.calculatePositionSize(ctx, param)
// 如果BTC在$50,000且maxLossLimit為$1000：
// 每單位風險 = $50,000 - $45,000 = $5,000
// 基於風險的最大數量 = $1,000 / $5,000 = 0.2 BTC
// 最終數量 = min(1.0, 0.2, 可用餘額) = 0.2 BTC
```

## 風險管理功能

### 1. 倉位規模優化
- **基於風險的規模計算**：根據止損距離計算倉位規模
- **虧損限制執行**：確保單筆交易不能超過最大虧損限制
- **餘額保護**：防止超出可用資金的過度槓桿

### 2. 價格驗證
- **止損驗證**：確保止損價格邏輯正確
- **市場價格整合**：使用實時市場價格進行計算
- **價格類型靈活性**：支持做市商和吃單者定價模型

### 3. 餘額管理
- **實時餘額檢查**：在交易執行前驗證可用餘額
- **多貨幣支持**：處理基礎和報價貨幣約束
- **精度處理**：遵守市場精度要求

## 數學基礎

### 風險計算公式

對於**買單**：
```
每單位風險 = 入場價格 - 止損價格
基於風險的最大數量 = 最大虧損限制 / 每單位風險
```

對於**賣單**：
```
每單位風險 = 止損價格 - 入場價格
基於風險的最大數量 = 最大虧損限制 / 每單位風險
```

### 最終倉位規模
```
最終數量 = min(
    請求數量,
    基於風險的最大數量,
    基於餘額的最大數量
)
```

## 使用案例

### 1. 保守風險管理
```yaml
tradingdesk:
  maxLossLimit: 50.0           # 每筆交易僅風險$50
  priceType: maker             # 使用做市商定價獲得更好成交
```

### 2. 中等風險交易
```yaml
tradingdesk:
  maxLossLimit: 200.0          # 每筆交易風險最多$200
  priceType: taker             # 使用吃單者定價立即執行
```

### 3. 高風險交易
```yaml
tradingdesk:
  maxLossLimit: 1000.0         # 每筆交易風險最多$1000
  priceType: maker             # 使用做市商定價最小化成本
```

## 整合示例

### 與交易機器人整合
```go
// 與自動交易系統整合
func executeTrade(symbol string, side types.SideType, quantity float64, stopLoss float64) {
    param := OpenPositionParam{
        Symbol:        symbol,
        Side:          side,
        Quantity:      fixedpoint.NewFromFloat(quantity),
        StopLossPrice: fixedpoint.NewFromFloat(stopLoss),
    }
    
    optimalSize, err := tradingDesk.calculatePositionSize(ctx, param)
    if err != nil {
        log.Error("Failed to calculate position size:", err)
        return
    }
    
    // 使用最優規模執行交易
    executeOrder(symbol, side, optimalSize)
}
```

### 與手動交易界面整合
```go
// 與交易UI整合
func handleTradeRequest(req TradeRequest) {
    // 驗證並計算最優倉位規模
    param := OpenPositionParam{
        Symbol:        req.Symbol,
        Side:          req.Side,
        Quantity:      req.Quantity,
        StopLossPrice: req.StopLoss,
    }
    
    optimalSize, err := tradingDesk.calculatePositionSize(ctx, param)
    if err != nil {
        return TradeResponse{Error: err.Error()}
    }
    
    return TradeResponse{
        OptimalSize: optimalSize,
        RiskAmount:  calculateRisk(optimalSize, req.StopLoss),
    }
}
```

## 最佳實踐

1. **設置適當的虧損限制**：根據您的風險承受能力和賬戶規模選擇 `maxLossLimit`
2. **始終使用止損**：當提供止損價格時策略效果最佳
3. **監控賬戶餘額**：確保有足夠的餘額進行預期的交易規模
4. **選擇正確的價格類型**：使用"maker"獲得更好定價，"taker"立即執行
5. **定期檢查**：根據性能定期檢查和調整虧損限制
6. **倉位規模計算**：讓策略計算最優規模而不是使用固定金額

## 限制

1. **需要手動執行**：策略計算規模但不自動執行交易
2. **止損依賴性**：當提供止損價格時效果最佳
3. **單筆交易焦點**：專為個別交易風險管理而設計，不是投資組合級別風險
4. **市場價格依賴性**：需要實時市場數據進行準確計算

## 錯誤處理

### 常見錯誤場景

**無效止損價格**
```
錯誤："invalid stop loss price for buy order"
解決方案：確保買單的止損低於入場價格，賣單的止損高於入場價格
```

**餘額不足**
```
行為：根據餘額返回最大可用數量
解決方案：確保賬戶餘額充足或減少交易規模
```

**市場數據不可用**
```
錯誤："failed to query ticker"
解決方案：檢查交易所連接性和符號有效性
```

## 測試

策略包含全面的單元測試，涵蓋：
- 正常倉位規模計算
- 邊緣情況（餘額不足、無效止損）
- 不同價格類型和市場條件
- 錯誤處理場景

運行測試：
```bash
go test ./pkg/strategy/tradingdesk/...
```

## 相關資源

- [BBGO策略開發指南](../../doc/topics/developing-strategy.md)
- [風險管理最佳實踐](../../doc/topics/risk-management.md)
- [配置示例](../../../config/tradingdesk.yaml)
- [倉位規模技術](../../doc/topics/position-sizing.md)
