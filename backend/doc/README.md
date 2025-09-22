# BlueLink 後端服務

基於 Go 和 Gin 框架的 REST API 服務，作為前端與 Sui 區塊鏈之間的橋樑。

## 功能

- 查詢所有項目
- 獲取特定項目詳情  
- 查詢用戶的捐贈記錄
- 健康檢查端點

## 快速開始

```bash
# 安裝依賴
go mod tidy

# 配置環境變數
cp .env.example .env

# 運行服務
go run main.go
```

服務將在 <http://localhost:8080> 上運行。

## 環境變數

- `PORT`: 服務器端口 (預設: 8080)
- `SUI_RPC_URL`: Sui RPC 節點地址

## API 端點

- `GET /api/projects` - 獲取所有項目
- `GET /api/projects/:id` - 獲取特定項目
- `GET /api/donors/:address` - 獲取捐贈記錄
- `GET /health` - 健康檢查
