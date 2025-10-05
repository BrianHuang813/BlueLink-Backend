# BlueLink Backend 架構說明

## 專案結構

````
backend/
├── cmd/
│   └── main.go                   # 應用程式入口
├── internal/
│   ├── config/
│   │   └── config.go             # 配置管理
│   ├── middleware/
│   │   ├── recovery.go           # Panic 恢復
│   │   ├── request_id.go         # 請求追蹤
│   │   ├── logging.go            # 日誌記錄
│   │   ├── cors.go               # 跨域處理
│   │   ├── rate_limit.go         # 流量限制
│   │   ├── error_handler.go      # 錯誤處理
│   │   ├── wallet_auth.go        # 錢包驗證
│   │   └── user_context.go       # 使用者上下文
│   ├── models/
│   │   ├── user.go               # 使用者資料模型
│   │   └── response.go           # API 回應格式
│   ├── utils/
│   │   └── context.go            # Context 輔助函數
│   ├── services/
│   │   └── user_service.go       # 使用者業務邏輯
│   ├── handlers/
│   │   ├── auth/
│   │   │   └── wallet.go         # 錢包連接處理
│   │   └── users/
│   │       └── profile.go        # 使用者資料處理
│   ├── routes/
│   │   └── routes.go             # 路由設定
│   └── sui/
│       └── client.go             # Sui 區塊鏈客戶端
````

## 架構設計原則

### 1. 分層架構 (Layered Architecture)

````
Request → Middleware → Handler → Service → Database
````

- **Middleware**: 處理橫切關注點（認證、日誌、錯誤處理）
- **Handler**: 處理 HTTP 請求和回應
- **Service**: 業務邏輯層
- **Models**: 資料模型定義

### 2. Middleware 執行順序

#### 全域 Middleware（所有路由）
````go
1. RecoveryMiddleware()        // 防止 panic 崩潰
2. RequestIDMiddleware()       // 追蹤請求
3. LoggingMiddleware()         // 記錄日誌
4. CORSMiddleware()            // 跨域處理
5. RateLimitMiddleware(N)      // 流量限制
6. ErrorHandlerMiddleware()    // 統一錯誤處理
````

#### Protected Routes Middleware（需要驗證的路由）
````go
1. WalletAuthMiddleware()      // 驗證 Sui 錢包簽名
2. UserContextMiddleware()     // 載入使用者基本資訊
````

## 安全性設計

### 1. 錢包驗證流程

#### Step 1: 前端請求挑戰訊息
````
POST /api/v1/auth/challenge
Body: { "wallet_address": "0x..." }

Response: {
  "message": "Login to BlueLink\nWallet: 0x...\nChallenge: xxx\nTimestamp: 123",
  "challenge": "base64_encoded_challenge"
}
````

#### Step 2: 前端使用私鑰簽署訊息
````typescript
const signature = await wallet.signMessage(message)
````

#### Step 3: 後端驗證簽名
````
POST /api/v1/auth/verify
Body: {
  "wallet_address": "0x...",
  "signature": "...",
  "message": "..."
}

Response: {
  "token": "auth_token",
  "wallet_address": "0x...",
  "user": { ... }
}
````

#### Step 4: 使用 Token 訪問受保護的路由
````
GET /api/v1/profile
Authorization: Bearer wallet:signature:message
````

### 2. 簽名驗證（wallet_auth.go）

````go
// 驗證流程
1. 解碼 Base64 簽名
2. 提取簽名方案標誌（Ed25519/Secp256k1/Secp256r1）
3. 從公鑰推導地址，驗證與聲稱的地址匹配
4. 使用 Blake2b-256 雜湊訊息
5. 驗證簽名
````

### 3. 防護機制

- **Rate Limiting**: 防止 DDoS 攻擊（每分鐘限制請求數）
- **Request ID**: 追蹤和除錯問題
- **Recovery**: 捕獲 panic 防止服務崩潰
- **挑戰過期**: 5 分鐘內必須完成驗證
- **CORS**: 限制允許的來源域

## Context 使用原則

### ✅ 存入 Context 的資料（輕量級）
````go
c.Set("RequestID", requestID)        // string
c.Set("WalletAddress", wallet)       // string
c.Set("UserID", userID)              // int64
c.Set("Role", role)                  // string
````

### ❌ 不要存入 Context 的資料（重量級）
````go
// 不要這樣做
c.Set("User", completeUserObject)    // 完整的使用者物件
c.Set("Bonds", allUserBonds)         // 大量關聯資料
````

### 正確的使用方式

````go
// 在 Handler 中按需載入完整資料
func GetProfile(c *gin.Context) {
    userID, _ := utils.GetUserID(c)
    
    // 只在需要時才從資料庫載入完整資料
    user, err := userService.GetByID(userID)
    if err != nil {
        // 錯誤處理
    }
    
    c.JSON(200, user)
}
````

## 環境配置

### 開發環境 (.env.development)
````bash
ENVIRONMENT=development
PORT=8080
SUI_RPC_URL=https://fullnode.devnet.sui.io:443
````

### 生產環境 (.env.production)
````bash
ENVIRONMENT=production
PORT=80
SUI_RPC_URL=https://fullnode.mainnet.sui.io:443
````

## API 路由結構

### Public Routes（不需要驗證）
````
GET  /api/v1/health                  # 健康檢查
POST /api/v1/auth/challenge          # 取得挑戰訊息
POST /api/v1/auth/verify             # 驗證簽名
GET  /api/v1/bonds                   # 列出所有債券
GET  /api/v1/bonds/:id               # 取得債券詳情
````

### Protected Routes（需要錢包驗證）
````
GET  /api/v1/profile                 # 取得基本資料
GET  /api/v1/profile/full            # 取得完整資料
PUT  /api/v1/profile                 # 更新資料
POST /api/v1/bonds                   # 發行債券
POST /api/v1/bonds/:id/purchase      # 購買債券
GET  /api/v1/my-bonds                # 我的債券
GET  /api/v1/transactions            # 交易歷史
````

## 錯誤處理

### 統一錯誤回應格式
````json
{
  "code": 400,
  "message": "Invalid request format",
  "details": "validation error: email is required",
  "request_id": "uuid-xxx"
}
````

### 成功回應格式
````json
{
  "code": 200,
  "message": "Success",
  "data": { ... },
  "request_id": "uuid-xxx"
}
````

## 待實作功能

### 高優先級
1. [ ] 資料庫連接和初始化
2. [ ] JWT Token 生成和驗證
3. [ ] 債券相關的 Handler 和 Service
4. [ ] 完整的單元測試

### 中優先級
1. [ ] Redis 快取層（用於 challenge 儲存）
2. [ ] 更完善的日誌系統（結構化日誌）
3. [ ] Metrics 監控
4. [ ] API 文件（Swagger）

### 低優先級
1. [ ] WebSocket 支援（即時通知）
2. [ ] 多語言支援
3. [ ] 管理後台 API

## 開發建議

### 1. 錯誤處理
````go
// 使用 models.ErrorResponse
c.JSON(http.StatusBadRequest, models.ErrorResponse{
    Code:    http.StatusBadRequest,
    Message: "Invalid request",
    Details: err.Error(),
})
````

### 2. 成功回應
````go
// 使用 models.SuccessResponse
c.JSON(http.StatusOK, models.SuccessResponse(
    http.StatusOK,
    "Operation successful",
    data,
))
````

### 3. 從 Context 取值
````go
// 使用 utils 輔助函數
walletAddress, err := utils.GetWalletAddress(c)
userID, err := utils.GetUserID(c)
role, err := utils.GetUserRole(c)
````

### 4. Service 層調用
````go
// Handler 中調用 Service
user, err := h.userService.GetByID(userID)
if err != nil {
    // 錯誤處理
}
````

## 測試策略

### 單元測試
````go
// 測試 Service 層
func TestUserService_GetByID(t *testing.T) {
    // 使用 mock database
}

// 測試 Middleware
func TestWalletAuthMiddleware(t *testing.T) {
    // 使用 httptest
}
````

### 整合測試
````go
// 測試完整的 API 流程
func TestAuthFlow(t *testing.T) {
    // 1. 取得 challenge
    // 2. 簽署訊息
    // 3. 驗證簽名
    // 4. 訪問受保護的路由
}
````

## 總結

這個架構具有以下特點：

✅ **清晰的職責分離**: 每個層級有明確的責任
✅ **高安全性**: 多層防護機制
✅ **易於維護**: 模組化設計，易於擴展
✅ **輕量級 Context**: 避免記憶體浪費
✅ **統一的錯誤處理**: 一致的 API 回應格式
✅ **完善的日誌追蹤**: Request ID 貫穿整個請求生命週期
