# BlueLink 債券市場 API 使用指南

## 概述

本文檔說明 BlueLink 後端債券市場 API 的實現細節和使用方法。

## 已實現的功能

### ✅ 1. GET /api/v1/bonds - 獲取所有債券

**端點**: `GET https://bluelink-backend-2tdo.onrender.com/api/v1/bonds`

**認證**: 不需要 (公開訪問)

**響應格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "on_chain_id": "0xabc123...",
      "issuer_address": "0x1234567890abcdef...",
      "issuer_name": "某某科技股份有限公司",
      "bond_name": "綠色能源債券 2024",
      "bond_image_url": "https://example.com/bond.png",
      "token_image_url": "https://example.com/token.png",
      "total_amount": 1000000000000,
      "amount_raised": 500000000000,
      "amount_redeemed": 0,
      "tokens_issued": 50,
      "tokens_redeemed": 0,
      "annual_interest_rate": 500,
      "maturity_date": "2025-12-31T00:00:00Z",
      "issue_date": "2024-01-01T00:00:00Z",
      "active": true,
      "redeemable": false,
      "metadata_url": "https://example.com/metadata.json",
      "created_at": "2024-01-01T08:00:00Z",
      "updated_at": "2024-11-01T10:30:00Z"
    }
  ]
}
```

**測試命令**:
```bash
# 基本請求
curl -X GET 'https://bluelink-backend-2tdo.onrender.com/api/v1/bonds' \
  -H 'Accept: application/json'

# 本地測試
curl -X GET 'http://localhost:8080/api/v1/bonds' \
  -H 'Accept: application/json'
```

**數據類型說明**:

| 字段 | 類型 | 單位 | 說明 |
|------|------|------|------|
| `total_amount` | number | MIST | 1 SUI = 1,000,000,000 MIST |
| `amount_raised` | number | MIST | 已募集金額 |
| `amount_redeemed` | number | MIST | 已贖回金額 |
| `tokens_issued` | number | 個 | 已發行代幣數量 |
| `tokens_redeemed` | number | 個 | 已贖回代幣數量 |
| `annual_interest_rate` | number | 基點 | 5% = 500, 3.5% = 350 |
| `maturity_date` | string | ISO 8601 | YYYY-MM-DDTHH:mm:ssZ |
| `issue_date` | string | ISO 8601 | YYYY-MM-DDTHH:mm:ssZ |
| `created_at` | string | ISO 8601 | YYYY-MM-DDTHH:mm:ssZ |
| `updated_at` | string | ISO 8601 | YYYY-MM-DDTHH:mm:ssZ |

---

### ✅ 2. GET /api/v1/bonds/:id - 獲取單個債券詳情

**端點**: `GET https://bluelink-backend-2tdo.onrender.com/api/v1/bonds/{id}`

**認證**: 不需要 (公開訪問)

**響應格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "on_chain_id": "0xabc123...",
    "issuer_address": "0x1234567890abcdef...",
    "issuer_name": "某某科技股份有限公司",
    "bond_name": "綠色能源債券 2024",
    "bond_image_url": "https://example.com/bond.png",
    "token_image_url": "https://example.com/token.png",
    "total_amount": 1000000000000,
    "amount_raised": 500000000000,
    "amount_redeemed": 0,
    "tokens_issued": 50,
    "tokens_redeemed": 0,
    "annual_interest_rate": 500,
    "maturity_date": "2025-12-31T00:00:00Z",
    "issue_date": "2024-01-01T00:00:00Z",
    "active": true,
    "redeemable": false,
    "metadata_url": "https://example.com/metadata.json",
    "created_at": "2024-01-01T08:00:00Z",
    "updated_at": "2024-11-01T10:30:00Z"
  }
}
```

**測試命令**:
```bash
curl -X GET 'https://bluelink-backend-2tdo.onrender.com/api/v1/bonds/1' \
  -H 'Accept: application/json'
```

---

### ✅ 3. POST /api/v1/bonds/sync - 同步鏈上交易

**端點**: `POST https://bluelink-backend-2tdo.onrender.com/api/v1/bonds/sync`

**認證**: 需要 (session cookie)

**請求格式**:
```json
{
  "transaction_digest": "ABC123DEF456...",
  "event_type": "bond_created"
}
```

**支持的事件類型**:

| event_type | 說明 |
|------------|------|
| `bond_created` | 創建債券 |
| `bond_purchased` | 購買債券 |
| `bond_redeemed` | 贖回債券 |
| `funds_withdrawn` | 提取資金 |
| `redemption_deposited` | 存入償還資金 |

**響應格式**:
```json
{
  "code": 200,
  "message": "Transaction will be indexed by event listener",
  "data": null
}
```

**測試命令**:
```bash
# 需要先登入獲取 session cookie
curl -X POST 'https://bluelink-backend-2tdo.onrender.com/api/v1/bonds/sync' \
  -H 'Content-Type: application/json' \
  -H 'Cookie: session_id=xxx' \
  -d '{
    "transaction_digest": "ABC123DEF456...",
    "event_type": "bond_created"
  }'
```

**注意事項**:
- 此端點主要用於前端通知後端有新交易需要索引
- 實際的交易處理由後台事件監聽器 (`EventListener`) 自動完成
- 前端調用此端點後,後台會自動查詢並處理鏈上交易

---

## 架構說明

### 數據流向

```
前端 (React)
    ↓ GET /api/v1/bonds
routes.go
    ↓
bond_handler.go (GetAllBonds)
    ↓
bond_service.go
    ↓
bond_repository.go (SQL查詢)
    ↓
PostgreSQL 數據庫
    ↓
響應數據經過格式轉換 (response.go)
    ↓
返回給前端 (JSON)
```

### 交易同步流程

```
前端創建鏈上交易
    ↓
交易成功
    ↓
前端調用 POST /bonds/sync
    ↓
後台事件監聽器自動處理:
    ├─ 從 Sui 鏈上查詢交易
    ├─ 解析交易事件
    ├─ 更新數據庫
    └─ 記錄交易日誌
```

### 關鍵文件

| 文件 | 說明 |
|------|------|
| `routes/routes.go` | 路由設置,將 `/bonds` 設為公開端點 |
| `handlers/bonds/bond_handler.go` | 處理 HTTP 請求 |
| `handlers/bonds/response.go` | 數據格式轉換 (snake_case + ISO 8601) |
| `handlers/bonds/request.go` | 請求結構驗證 |
| `services/bond_service.go` | 業務邏輯層 |
| `repository/bond_repository.go` | 數據庫操作層 |
| `models/bond.go` | 數據模型 |
| `blockchain/event_listener.go` | 鏈上事件監聽器 |

---

## 與前端的集成

### 前端調用示例

```typescript
// src/lib/api.ts
export async function getAllBonds(): Promise<Bond[]> {
  const response = await fetch('https://bluelink-backend-2tdo.onrender.com/api/v1/bonds', {
    method: 'GET',
    headers: {
      'Accept': 'application/json',
    },
    credentials: 'include', // 如果需要 cookies
  });

  if (!response.ok) {
    throw new Error('Failed to fetch bonds');
  }

  const json = await response.json();
  return json.data; // 返回 data 數組
}
```

### 數據轉換

前端收到的數據已經是正確的格式:
- ✅ 字段名稱使用 `snake_case`
- ✅ 金額是 `number` 類型 (MIST 單位)
- ✅ 利率是 `number` 類型 (基點)
- ✅ 日期是 `string` 類型 (ISO 8601)
- ✅ 空數組返回 `[]` 而不是 `null`

---

## 後台事件監聽器

### 自動同步機制

後台運行的 `EventListener` 會自動:

1. **每 15 秒**輪詢 Sui 鏈上的新事件
2. **過濾**系統事件,只處理 BlueLink 合約事件
3. **解析**事件數據並更新數據庫
4. **防止重複處理**同一個交易

### 支持的事件類型

#### BondProjectCreated
- 創建新的 Bond 記錄
- 自動創建發行者 User 記錄
- 記錄創建交易

#### BondTokensPurchased
- 創建 BondToken 記錄
- 更新 Bond 的 `amount_raised` 和 `tokens_issued`
- 創建購買交易記錄

#### BondTokenRedeemed
- 更新 BondToken 的 `is_redeemed` 狀態
- 更新 Bond 的 `tokens_redeemed` 和 `amount_redeemed`
- 創建贖回交易記錄

#### FundsWithdrawn
- 記錄資金提取事件
- 創建提取交易記錄

#### RedemptionFundsDeposited
- 記錄償還資金存入事件
- 創建存款交易記錄

---

## 數據庫模型

### bonds 表結構

```sql
CREATE TABLE bonds (
    id BIGSERIAL PRIMARY KEY,
    on_chain_id VARCHAR(66) UNIQUE NOT NULL,
    issuer_address VARCHAR(66) NOT NULL,
    issuer_name VARCHAR(255),
    bond_name VARCHAR(255) NOT NULL,
    bond_image_url TEXT,
    token_image_url TEXT,
    metadata_url TEXT,
    total_amount BIGINT NOT NULL,
    amount_raised BIGINT DEFAULT 0,
    amount_redeemed BIGINT DEFAULT 0,
    tokens_issued BIGINT DEFAULT 0,
    tokens_redeemed BIGINT DEFAULT 0,
    annual_interest_rate BIGINT NOT NULL,
    maturity_date VARCHAR(10) NOT NULL,  -- YYYY-MM-DD
    issue_date VARCHAR(10) NOT NULL,     -- YYYY-MM-DD
    active BOOLEAN DEFAULT TRUE,
    redeemable BOOLEAN DEFAULT FALSE,
    raised_funds_balance BIGINT DEFAULT 0,
    redemption_pool_balance BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_bonds_issuer_address ON bonds(issuer_address);
CREATE INDEX idx_bonds_active ON bonds(active);
CREATE INDEX idx_bonds_on_chain_id ON bonds(on_chain_id);
```

---

## 測試清單

### ✅ API 測試

- [x] GET /api/v1/bonds 返回 200
- [x] 響應格式符合 `{code, message, data}` 結構
- [x] `data` 是數組
- [x] 字段名稱使用 snake_case
- [x] 數值字段是 number 類型
- [x] 日期字段是 ISO 8601 字符串
- [x] 空結果返回 `[]` 而不是 `null`

### ✅ CORS 測試

- [x] 前端域名可以訪問
- [x] 允許 credentials (cookies)
- [x] OPTIONS 預檢請求正常

### ✅ 數據庫測試

- [x] 債券記錄可以正確創建
- [x] 金額字段是 BIGINT 類型
- [x] 日期字段是 VARCHAR 類型
- [x] 索引已創建

---

## 部署說明

### 環境變數

確保以下環境變數已設置:

```bash
# Sui 區塊鏈設定
SUI_RPC_URL=https://fullnode.testnet.sui.io:443
SUI_PACKAGE_ID=0x...  # BlueLink 合約地址

# 資料庫設定 (生產環境)
DATABASE_URL=postgresql://user:pass@host:port/dbname?sslmode=require

# CORS 設定
CORS_ALLOWED_ORIGINS=https://bluelink-frontend.vercel.app,http://localhost:5173

# 其他設定
ENV=production
PORT=8080
```

### 啟動服務

```bash
cd backend
go run cmd/main.go
```

服務會自動:
1. 連接數據庫
2. 執行遷移
3. 啟動事件監聽器
4. 啟動 HTTP 服務器

---

## 常見問題

### Q: 前端收到的債券列表為空?

**A**: 檢查以下幾點:
1. 數據庫中是否有債券記錄 (`SELECT COUNT(*) FROM bonds WHERE deleted_at IS NULL`)
2. 後台事件監聽器是否正常運行 (檢查日誌)
3. `SUI_PACKAGE_ID` 環境變數是否正確設置

### Q: CORS 錯誤?

**A**: 確保:
1. `CORS_ALLOWED_ORIGINS` 包含前端域名
2. 前端請求包含 `credentials: 'include'` (如果需要 cookies)
3. 後端 CORS 中間件允許 credentials

### Q: 日期格式不正確?

**A**: 
- 數據庫中日期存儲為 `YYYY-MM-DD` 格式
- API 響應會自動轉換為 ISO 8601 格式 (`YYYY-MM-DDTHH:mm:ssZ`)
- 檢查 `response.go` 中的 `formatDateToISO8601` 函數

### Q: 數值顯示不正確?

**A**: 
- 後端以 MIST 為單位存儲金額 (1 SUI = 1,000,000,000 MIST)
- 利率以基點存儲 (5% = 500)
- 前端需要自行轉換顯示單位

---

## 開發提示

### 添加新字段

如果需要在 Bond 模型中添加新字段:

1. 更新 `models/bond.go`
2. 更新 `handlers/bonds/response.go` 的 `BondResponse`
3. 更新 `repository/bond_repository.go` 的 SQL 查詢
4. 執行數據庫遷移
5. 更新 API 文檔

### 調試技巧

```bash
# 查看後台日誌
tail -f logs/bluelink.log

# 測試數據庫連接
psql $DATABASE_URL

# 查看債券記錄
SELECT id, bond_name, amount_raised, tokens_issued FROM bonds;

# 查看最近的交易
SELECT tx_hash, event_type, timestamp FROM transactions ORDER BY timestamp DESC LIMIT 10;
```

---

## 聯繫支持

如有問題請聯繫開發團隊或查看:
- GitHub: https://github.com/BrianHuang813/Blue-Bond-Funding-Platform
- API 文檔: https://bluelink-backend-2tdo.onrender.com/api/v1/health
