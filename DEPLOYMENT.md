# BlueLink 部署指南

本指南將幫助您在本地或測試網環境中部署 BlueLink 平台。

## 部署步驟

### 1. 準備環境

確保已安裝以下工具：
- Node.js (v18+)
- Go (v1.21+)
- Sui CLI
- Git

### 2. 部署智能合約

```bash
cd sui-contracts

# 檢查 Sui 客戶端配置
sui client active-env

# 如果需要，切換到測試網
sui client switch --env testnet

# 獲取測試代幣
sui client faucet

# 構建合約
sui move build

# 部署合約 (記錄包地址)
sui client publish --gas-budget 20000000
```

### 3. 配置後端

```bash
cd ../backend-go

# 安裝依賴
go mod tidy

# 配置環境變數
cp .env.example .env
# 編輯 .env 設定 PORT 和 SUI_RPC_URL

# 更新 main.go 中的包地址
# 找到並替換 "0x0::bluelink::Project" 和 "0x0::bluelink::DonationReceipt"
# 改為實際的包地址
```

### 4. 配置前端

```bash
cd ../frontend-react

# 安裝依賴
npm install

# 更新智能合約地址
# 編輯以下文件中的合約地址：
# - src/pages/ProjectDetailPage.tsx
# - src/pages/CreateProjectPage.tsx
# 將 "0x0::bluelink::donate" 和 "0x0::bluelink::create_project" 
# 改為實際的包地址
```

### 5. 啟動服務

終端1 - 後端：
```bash
cd backend-go
go run main.go
```

終端2 - 前端：
```bash
cd frontend-react  
npm run dev
```

### 6. 測試平台

1. 在瀏覽器中訪問 http://localhost:3000
2. 安裝 Sui 錢包擴充程序
3. 連接錢包並切換到測試網
4. 獲取測試 SUI 代幣
5. 測試創建項目和捐贈功能

## 生產部署建議

### 後端
- 使用 Docker 容器化
- 配置負載均衡器
- 設置監控和日志記錄
- 使用 HTTPS

### 前端
- 構建生產版本：`npm run build`
- 部署到 CDN (如 Vercel, Netlify)
- 配置域名和 SSL
- 優化 SEO 設置

### 智能合約
- 在主網部署前進行充分測試
- 考慮合約升級策略
- 設置多重簽名管理
- 進行安全審計

## 故障排除

### 常見問題

1. **合約部署失敗**
   - 檢查 gas 預算是否充足
   - 確認錢包有足夠的 SUI 代幣
   - 檢查 Move.toml 配置

2. **後端無法獲取數據**
   - 確認 RPC 端點可訪問
   - 檢查包地址是否正確
   - 驗證網絡配置

3. **前端錢包連接問題**
   - 確保錢包擴充程序已安裝
   - 檢查網絡設置是否匹配
   - 清除瀏覽器快取

### 日志檢查

- Sui 客戶端：`sui client --help`
- 後端服務：檢查控制台輸出
- 前端應用：使用瀏覽器開發者工具

## 監控和維護

- 定期檢查合約狀態
- 監控後端服務健康狀況
- 跟蹤前端性能指標
- 備份重要配置和數據

## 支持

如遇到部署問題，請：
1. 檢查本指南的故障排除部分
2. 查看項目 Issues 頁面
3. 創建新 Issue 並提供詳細信息
