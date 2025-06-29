# 市值策略

## 概述

市值策略是一個投資組合再平衡策略，它根據 CoinMarketCap 的市值數據自動調整您的加密貨幣持倉。此策略有助於維持一個反映不同加密貨幣相對市場價值的多元化投資組合。

## 工作原理

1. **數據收集**：策略從 CoinMarketCap API 獲取實時市值數據
2. **權重計算**：根據所選加密貨幣的相對市值計算目標投資組合權重
3. **投資組合再平衡**：策略自動下達買賣訂單，使實際投資組合與目標權重保持一致
4. **持續監控**：持續運行，按指定間隔進行再平衡

## 主要特性

- **基於市值的配置**：根據加密貨幣市值自動分配資金
- **可配置的再平衡**：設置自定義的投資組合再平衡間隔
- **閾值控制**：僅在差異超過指定閾值時進行再平衡
- **風險管理**：內建最大訂單金額限制
- **模擬運行模式**：在不下達真實訂單的情況下測試策略
- **多種訂單類型**：支持 LIMIT、LIMIT_MAKER 和 MARKET 訂單

## 前置條件

### API 密鑰設置
您必須獲取 CoinMarketCap API 密鑰並將其設置為環境變量：

```bash
export COINMARKETCAP_API_KEY="your_api_key_here"
```

從以下網址獲取免費 API 密鑰：https://coinmarketcap.com/api/

### 交易所配置
確保在您的 BBGO 配置文件中正確配置了交易所會話。

## 配置參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| `interval` | 字符串 | 是 | 再平衡間隔（例如："5m"、"1h"、"1d"） |
| `quoteCurrency` | 字符串 | 是 | 投資組合基礎貨幣（例如："USDT"、"TWD"） |
| `quoteCurrencyWeight` | 百分比 | 是 | 報價貨幣的權重分配 |
| `baseCurrencies` | 數組 | 是 | 要包含在投資組合中的加密貨幣列表 |
| `threshold` | 百分比 | 是 | 觸發再平衡的最小權重差異 |
| `maxAmount` | 數字 | 否 | 每筆訂單的最大金額（以報價貨幣計） |
| `queryInterval` | 字符串 | 否 | 獲取市值數據的間隔（默認："1h"） |
| `orderType` | 字符串 | 否 | 訂單類型：LIMIT、LIMIT_MAKER、MARKET（默認：LIMIT_MAKER） |
| `dryRun` | 布爾值 | 否 | 啟用模擬運行模式（默認：false） |

## 配置示例

```yaml
exchangeStrategies:
  - on: binance
    marketcap:
      interval: 1h                    # 每小時再平衡一次
      quoteCurrency: USDT            # 使用 USDT 作為基礎貨幣
      quoteCurrencyWeight: 10%       # 保持 10% 的 USDT
      baseCurrencies:                # 投資組合加密貨幣
        - BTC
        - ETH
        - BNB
        - ADA
        - DOT
      threshold: 2%                  # 當差異 > 2% 時再平衡
      maxAmount: 1000               # 每筆訂單最大 $1000
      queryInterval: 30m            # 每 30 分鐘更新市值數據
      orderType: LIMIT_MAKER        # 使用掛單以獲得更好的手續費
      dryRun: false                 # 執行真實交易
```

## 策略邏輯

### 權重計算
1. 獲取所有 `baseCurrencies` 的市值數據
2. 根據市值比例計算相對權重
3. 按 `(1 - quoteCurrencyWeight)` 縮放權重，為報價貨幣預留空間
4. 將報價貨幣權重添加到目標配置中

### 再平衡決策
對於投資組合中的每種貨幣：
1. 計算當前權重：`當前價值 / 總投資組合價值`
2. 與市值數據的目標權重進行比較
3. 如果 `|當前權重 - 目標權重| > 閾值`，觸發再平衡
4. 計算達到目標權重所需的買賣數量

### 訂單執行
1. 取消任何現有的活躍訂單
2. 根據權重差異生成新訂單
3. 如果指定了 `maxAmount`，則應用限制
4. 向交易所提交訂單（除非處於模擬運行模式）

## 風險管理

### 內建保護機制
- **閾值控制**：防止因微小權重變化而過度交易
- **最大訂單規模**：限制每筆交易的風險敞口
- **模擬運行模式**：允許在無財務風險的情況下測試策略
- **訂單類型選擇**：根據風險承受能力選擇適當的訂單類型

### 推薦設置
- **保守型**：`threshold: 5%`、`maxAmount: 500`、`orderType: LIMIT_MAKER`
- **中等型**：`threshold: 2%`、`maxAmount: 1000`、`orderType: LIMIT_MAKER`
- **激進型**：`threshold: 1%`、`maxAmount: 2000`、`orderType: MARKET`

## 監控和日誌記錄

策略提供詳細的日誌記錄：
- 市值數據更新
- 權重計算和比較
- 訂單生成和執行
- 再平衡決策和閾值

監控日誌以確保策略按預期工作：
```bash
bbgo run --config your_config.yaml --verbose
```

## 常見用例

### 1. 前十大加密貨幣投資組合
維持按市值排名前十的加密貨幣投資組合：
```yaml
baseCurrencies: [BTC, ETH, BNB, XRP, ADA, DOGE, MATIC, SOL, DOT, LTC]
quoteCurrencyWeight: 5%
threshold: 3%
```

### 2. 保守型 DeFi 投資組合
專注於成熟的 DeFi 代幣，現金配置較高：
```yaml
baseCurrencies: [ETH, BNB, UNI, AAVE, COMP]
quoteCurrencyWeight: 20%
threshold: 5%
```

### 3. 激進增長投資組合
針對新興加密貨幣，頻繁再平衡：
```yaml
baseCurrencies: [ETH, SOL, AVAX, NEAR, FTM, ATOM]
quoteCurrencyWeight: 5%
threshold: 1%
interval: 30m
```

## 故障排除

### 常見問題

**API 密鑰錯誤**
- 確保設置了 `COINMARKETCAP_API_KEY` 環境變量
- 驗證 API 密鑰有效且有足夠的配額
- 檢查 API 密鑰權限

**餘額不足**
- 確保報價貨幣有足夠的餘額進行再平衡
- 考慮減少 `maxAmount` 或增加 `threshold`

**訂單失敗**
- 檢查交易所特定的最小訂單規模
- 驗證交易對在您的交易所可用
- 確保有足夠的餘額支付交易手續費

**市場數據問題**
- 驗證加密貨幣符號與 CoinMarketCap 列表匹配
- 檢查網絡連接
- 監控 API 速率限制

### 調試模式
啟用詳細日誌記錄進行故障排除：
```bash
bbgo run --config your_config.yaml --debug
```

## 性能考慮

- **API 限制**：CoinMarketCap 免費版有速率限制；相應調整 `queryInterval`
- **交易手續費**：頻繁再平衡會產生交易手續費；平衡 `threshold` 與優化
- **市場影響**：大額訂單可能影響價格；使用適當的 `maxAmount` 設置
- **延遲**：設置短再平衡間隔時考慮交易所延遲

## 相關資源

- [BBGO 策略開發指南](../../doc/topics/developing-strategy.md)
- [再平衡策略](../rebalance/README.md) - 替代再平衡方法
- [配置示例](../../../config/marketcap.yaml)
- [CoinMarketCap API 文檔](https://coinmarketcap.com/api/documentation/v1/)
