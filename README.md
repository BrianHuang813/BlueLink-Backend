# BlueLink - Blue Bond Funding Platform

基於 Sui 區塊鏈的藍色債券募資平台，致力於為海洋保育、永續漁業等藍色經濟項目提供透明、可信的資金募集解決方案。

## 🌊 什麼是藍色債券？

藍色債券（Blue Bonds）是專門為海洋保育和永續海洋經濟項目籌集資金的金融工具。投資者購買債券支持海洋相關項目，並在到期時獲得本金和利息回報。

## 🌟 核心功能

### 投資者功能

- **錢包連接登入**: 使用 Sui 錢包簽名驗證身份，零密碼體驗
- **瀏覽債券項目**: 查看所有上架的藍色債券，包含發行者、利率、到期日等完整資訊
- **購買債券代幣**: 投資支持海洋保育項目，獲得 NFT 債券代幣
- **追蹤投資組合**: 查看個人持有的債券代幣和收益狀況
- **到期贖回**: 在到期日贖回本金和利息
- **多裝置管理**: 支援查看和管理所有登入裝置

### 發行者功能

- **項目發行**: 發行藍色債券為海洋項目募資
- **資金管理**: 管理募集資金的使用和贖回池
- **進度追蹤**: 即時監控募資進度和投資者情況
- **狀態控制**: 暫停/恢復債券銷售，存入贖回資金

### 平台特色

- **全程鏈上**: 所有交易記錄永久保存在 Sui 區塊鏈上
- **即時同步**: 區塊鏈事件監聽，數據即時同步
- **透明可追溯**: 資金流向完全公開透明
- **安全可信**: 智能合約自動執行，無需中介機構
- **高性能**: 基於 Sui 區塊鏈的高吞吐量和低延遲

## 🏗️ 技術架構

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │  Sui Blockchain │
│                 │    │                 │    │                 │
│ React + Vite    │◄──►│ Go + Gin       │◄──►│ Move Contracts  │
│ TypeScript      │    │ PostgreSQL     │    │ Event Listener  │
│ Wallet Connect  │    │ Session Auth   │    │ RPC API         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 技術棧

- **區塊鏈**: Sui (Move 智能合約)
- **後端**: Go 1.24 + Gin 框架
- **資料庫**: PostgreSQL
- **區塊鏈整合**: Sui Go SDK
- **認證**: 錢包簽名驗證 + Memory Session 管理
- **前端**: React + TypeScript + Vite (規劃中)

## 📁 專案結構

```text
BlueLink_backend/
├── README.md
├── backend/                    # Go 後端服務
│   ├── go.mod                 # Go 模組定義
│   ├── cmd/                   # 應用程式入口點
│   │   ├── main.go           # 主程式入口
│   │   └── migrate/          # 資料庫遷移工具
│   │       └── main.go
│   ├── internal/             # 內部業務邏輯
│   │   ├── blockchain/       # 區塊鏈相關
│   │   │   └── event_listener.go  # Sui 事件監聽器
│   │   ├── config/           # 配置管理
│   │   │   └── config.go
│   │   ├── database/         # 資料庫連接和遷移
│   │   │   ├── migrations.go
│   │   │   └── postgres.go
│   │   ├── handlers/         # HTTP 處理器
│   │   │   ├── auth/         # 認證相關
│   │   │   │   └── auth_handler.go
│   │   │   ├── bonds/        # 債券相關
│   │   │   │   ├── bond_handler.go
│   │   │   │   └── request.go
│   │   │   └── users/        # 用戶相關
│   │   │       ├── profile_handler.go
│   │   │       └── request.go
│   │   ├── logger/           # 日誌系統
│   │   │   └── logger.go
│   │   ├── middleware/       # 中介軟體
│   │   │   ├── cors.go
│   │   │   ├── error_handler.go
│   │   │   ├── logging.go
│   │   │   ├── rate_limit.go
│   │   │   ├── recovery.go
│   │   │   ├── request_id.go
│   │   │   ├── require_role.go
│   │   │   └── session_auth.go
│   │   ├── models/           # 資料模型
│   │   │   ├── bond.go
│   │   │   ├── response.go
│   │   │   ├── session.go
│   │   │   ├── transaction.go
│   │   │   └── user.go
│   │   ├── repository/       # 資料庫操作層
│   │   │   ├── bond_repository.go
│   │   │   ├── transaction_repository.go
│   │   │   └── user_repository.go
│   │   ├── routes/           # 路由設定
│   │   │   └── routes.go
│   │   ├── services/         # 業務邏輯層
│   │   │   ├── bond_service.go
│   │   │   └── user_service.go
│   │   ├── session/          # Session 管理
│   │   │   └── memory_session.go
│   │   └── utils/            # 工具函數
│   │       └── context.go
│   ├── tmp/                  # 編譯輸出
│   └── vendor/               # Go 依賴套件
└── tmp/                      # 臨時文件
```

## 🚀 快速開始

### 環境要求

- **Go**: 1.24+
- **PostgreSQL**: 14+
- **Sui 錢包**: 用於登入認證

### 1. 克隆專案

```bash
git clone <repository-url>
cd BlueLink_backend
```

### 2. 設定環境變數

```bash
cd backend
cp .env.example .env
```

編輯 `.env` 文件：

```env
# 環境設定
ENV=development
PORT=:8080

# 資料庫設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=bluelink
DB_SSL_MODE=disable

# Sui 區塊鏈設定
SUI_RPC_URL=https://fullnode.testnet.sui.io:443
SUI_PACKAGE_ID=your_contract_package_id

# 安全設定
JWT_SECRET=your_jwt_secret_key
SESSION_TIMEOUT=86400

# 日誌設定
LOG_LEVEL=info
ENABLE_SWAGGER=true
```

### 3. 安裝依賴

```bash
go mod download
```

### 4. 設定資料庫

確保 PostgreSQL 正在運行，然後執行遷移：

```bash
# 手動運行遷移（可選）
go run cmd/migrate/main.go

# 或者在開發環境下啟動時會自動執行遷移
```

### 5. 啟動服務

```bash
go run cmd/main.go
```

服務將在 `http://localhost:8080` 啟動。

## 📋 API 端點

### 健康檢查

- `GET /health` - 系統健康狀態檢查

### 認證 API (v1/auth)

```text
POST /api/v1/auth/challenge     # 取得登入挑戰訊息
POST /api/v1/auth/verify        # 驗證錢包簽名並登入
POST /api/v1/auth/logout        # 登出當前 Session
POST /api/v1/auth/logout-all    # 登出所有裝置
```

### 用戶 API (需要認證)

```text
GET  /api/v1/profile            # 取得基本個人資料
GET  /api/v1/profile/full       # 取得完整個人資料
PUT  /api/v1/profile            # 更新個人資料
GET  /api/v1/sessions           # 取得所有活躍 Session
DELETE /api/v1/sessions/:id     # 撤銷特定 Session
```

### 債券 API (需要認證)

```
GET /api/v1/bonds               # 取得所有債券列表
```

### 管理員 API (需要管理員權限)

管理員功能待實現

## 🔄 區塊鏈整合

### 事件監聽

系統會自動監聽以下 Sui 區塊鏈事件：

#### 債券相關事件

- `BondProjectCreated` - 債券項目創建
- `BondTokensPurchased` - 債券代幣購買
- `BondTokenRedeemed` - 債券代幣贖回

#### 資金管理事件

- `RedemptionFundsDeposited` - 贖回資金存入
- `FundsWithdrawn` - 資金提取

#### 狀態控制事件

- `SalePaused` - 銷售暫停
- `SaleResumed` - 銷售恢復

### 數據同步

- **即時監聽**: 每 15 秒查詢新事件
- **自動重試**: 失敗事件會記錄錯誤並重試
- **去重處理**: 防止重複處理相同交易
- **狀態更新**: 自動更新債券和用戶持倉狀態

## 🎯 業務流程

### 1. 用戶認證流程

```
用戶 → 請求 Challenge → 錢包簽名 → 提交驗證 → 創建 Session → 認證完成
```

### 2. 債券投資流程

```text
瀏覽債券 → 選擇投資 → 錢包交易 → 鏈上確認 → 更新持倉 → 投資完成
```

### 3. 債券發行流程

```text
創建項目 → 設定參數 → 部署合約 → 開始募資 → 管理資金 → 到期贖回
```

## 🔒 安全特性

### 認證安全

- **錢包簽名驗證**: 使用 Sui 原生簽名，無需密碼
- **Nonce 機制**: 防止重放攻擊
- **Session 管理**: HttpOnly Cookie，24 小時有效期
- **多裝置限制**: 每用戶最多 3 個同時登入

### API 安全

- **限流保護**: 防止暴力破解和 DDoS
- **CORS 設定**: 跨域請求控制
- **錯誤處理**: 統一錯誤回應格式
- **請求追蹤**: Request ID 和審計日誌

### 資料安全

- **軟刪除**: 保留資料追蹤記錄
- **事務處理**: 確保資料一致性
- **敏感資料**: 不記錄私鑰等敏感資訊

## 🧪 測試和開發

### 啟動開發模式

```bash
# 啟動時會顯示 BlueLink Logo
go run cmd/main.go
```

### 健康檢查

```bash
curl http://localhost:8080/health
```

### 查看日誌

系統會輸出詳細的結構化日誌，包括：

- 資料庫連接狀態
- Sui 客戶端狀態
- 事件監聽狀態
- API 請求追蹤

## 🌍 永續發展目標

本平台致力於支持聯合國永續發展目標（SDGs）第 14 項：

### 🐟 海洋生態保護

- 支持海洋保育項目募資
- 追蹤資金使用透明度
- 促進永續漁業發展

### 💙 藍色經濟

- 連接投資者與海洋項目
- 創造透明的資金機制
- 平衡環境保護與經濟發展
