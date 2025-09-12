# BlueLink 前端應用

基於 React、TypeScript 和 Vite 構建的現代化 Web 應用程序。

## 功能

- 瀏覽可持續發展項目
- 創建新項目
- 進行捐贈並獲得 NFT 憑證
- 查看個人捐贈記錄
- 錢包連接與交易簽名

## 技術棧

- React 18 + TypeScript
- Vite (構建工具)
- Tailwind CSS (樣式)
- @mysten/dapp-kit (Sui 整合)
- Axios (HTTP 客戶端)
- React Router (路由)

## 快速開始

```bash
# 安裝依賴
npm install

# 開發模式運行
npm run dev

# 構建生產版本
npm run build
```

## 項目結構

- `src/components/` - 可重用組件
- `src/pages/` - 頁面組件
- `src/services/` - API 服務
- `src/types/` - TypeScript 類型定義

## 配置

- Sui 網絡配置：`src/App.tsx`
- API 基礎 URL：`src/services/api.ts`
