# BlueLink - 透明的可持續發展捐贈平台

BlueLink 是一個基於 Sui 區塊鏈的去中心化捐贈平台，旨在解決可持續發展項目資金流向不透明的問題。通過區塊鏈技術，捐贈者可以追蹤其資金流向，項目創建者（如 NPO、NGO）可以獲得可驗證的資金支持。

## 🌟 核心功能

- **透明的資金追蹤**: 所有捐贈和提款都記錄在 Sui 區塊鏈上
- **NFT 捐贈憑證**: 每筆捐贈都會生成一個 NFT 作為永久證明
- **項目創建與管理**: 任何人都可以創建可持續發展項目並進行募資
- **實時資金監控**: 查看項目的實時籌款進度和捐贈統計

## 🏗️ 技術架構

### 區塊鏈層
- **區塊鏈**: Sui
- **智能合約語言**: Move
- **網絡**: Sui Testnet (可配置)

### 後端層
- **語言**: Golang
- **框架**: Gin
- **功能**: Sui RPC 客戶端，REST API 服務

### 前端層
- **框架**: React + TypeScript
- **構建工具**: Vite
- **區塊鏈整合**: @mysten/dapp-kit
- **樣式**: Tailwind CSS

## 📁 項目結構

```
BlueLink/
├── sui-contracts/          # Sui 智能合約
│   ├── sources/
│   │   └── bluelink.move   # 主要合約文件
│   └── Move.toml           # Move 項目配置
├── backend-go/             # Go 後端服務
│   ├── main.go             # 主要服務器文件
│   ├── go.mod              # Go 模組配置
│   └── .env.example        # 環境變數示例
└── frontend-react/         # React 前端應用
    ├── src/
    │   ├── components/     # 可重用組件
    │   ├── pages/          # 頁面組件
    │   ├── services/       # API 服務
    │   └── types/          # TypeScript 類型
    ├── package.json        # Node.js 依賴
    └── vite.config.ts      # Vite 配置
```

## 🚀 快速開始

### 前置要求

- [Node.js](https://nodejs.org/) (v18+)
- [Go](https://golang.org/) (v1.21+)
- [Sui CLI](https://docs.sui.io/build/install)
- Sui 錢包瀏覽器擴充功能

### 1. 克隆項目

```bash
git clone <repository-url>
cd BlueLink
```

### 2. 部署智能合約

```bash
cd sui-contracts

# 構建合約
sui move build

# 部署到測試網
sui client publish --gas-budget 20000000
```

記錄部署後的包地址，需要更新前端和後端配置。

### 3. 啟動後端服務

```bash
cd ../backend-go

# 安裝依賴
go mod tidy

# 複製並配置環境變數
cp .env.example .env

# 啟動服務器
go run main.go
```

後端將在 `http://localhost:8080` 上運行。

### 4. 啟動前端應用

```bash
cd ../frontend-react

# 安裝依賴
npm install

# 啟動開發服務器
npm run dev
```

前端將在 `http://localhost:3000` 上運行。

## 🔧 配置說明

### 智能合約配置

部署合約後，需要更新以下文件中的包地址：

1. `backend-go/main.go` - 更新 `objectType` 變數
2. `frontend-react/src/pages/ProjectDetailPage.tsx` - 更新 `target` 地址
3. `frontend-react/src/pages/CreateProjectPage.tsx` - 更新 `target` 地址

### 後端配置

創建 `backend-go/.env` 文件：

```env
PORT=8080
SUI_RPC_URL=https://fullnode.testnet.sui.io:443
```

### 前端配置

如果需要連接到不同的網絡或後端，請修改：
- `src/App.tsx` - 網絡配置
- `src/services/api.ts` - API 基礎 URL

## 📋 API 端點

### GET /api/projects
獲取所有項目列表

### GET /api/projects/:id  
獲取特定項目詳情

### GET /api/donors/:address
獲取指定地址的捐贈記錄

### GET /health
健康檢查端點

## 🎯 使用流程

### 創建項目
1. 連接 Sui 錢包
2. 訪問 "創建項目" 頁面
3. 填寫項目信息和籌款目標
4. 提交交易並等待確認

### 進行捐贈
1. 瀏覽項目列表
2. 選擇感興趣的項目
3. 在項目詳情頁面輸入捐贈金額
4. 確認交易並獲得 NFT 憑證

### 查看記錄
1. 訪問 "我的儀表板"
2. 查看所有捐贈記錄和 NFT 憑證
3. 跟蹤支持的項目狀態

## 🔒 安全考慮

- 所有資金操作都通過智能合約執行
- 私鑰永遠不會離開用戶的錢包
- 所有交易都需要用戶簽名確認
- 資金流向完全透明且不可篡改

## 🌱 可持續發展目標

BlueLink 支持聯合國可持續發展目標 (SDGs)，特別關注：
- 清潔能源和氣候行動
- 負責任的消費和生產
- 減少不平等
- 海洋和陸地生態保護

## 🤝 貢獻指南

1. Fork 項目
2. 創建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 開啟 Pull Request

## 📄 許可證

本項目採用 MIT 許可證 - 詳見 [LICENSE](LICENSE) 文件。

## 🆘 支持

如果您遇到任何問題或有疑問，請：
1. 查看 [Issues](../../issues) 頁面
2. 創建新的 Issue 詳細描述問題
3. 加入我們的社群討論

## 🚧 開發狀態

這是一個 MVP（最小可行產品）版本，未來將添加：
- 項目類別和標籤系統
- 高級搜索和過濾功能
- 項目更新和里程碑追蹤
- 多語言支持
- 行動端優化

---

**讓我們一起建設一個更透明、更可持續的未來！** 🌍✨
