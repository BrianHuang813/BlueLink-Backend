# BlueLink Sui 智能合約

使用 Move 語言編寫的智能合約，實現透明的捐贈平台功能。

## 合約結構

### 對象定義

#### Project
- `id`: 項目唯一 ID
- `creator`: 創建者地址
- `name`: 項目名稱
- `description`: 項目描述
- `funding_goal`: 籌款目標 (MIST)
- `total_raised`: 已籌款額 (Balance<SUI>)
- `donor_count`: 捐贈者數量

#### DonationReceipt (NFT)
- `id`: NFT 唯一 ID
- `project_id`: 對應項目 ID
- `donor`: 捐贈者地址
- `amount`: 捐贈金額 (MIST)

### 入口函數

#### `create_project`
創建新的可持續發展項目

#### `donate`
向項目捐贈 SUI 並獲得 NFT 憑證

#### `withdraw`
項目創建者提取已籌集的資金

## 部署

```bash
# 構建合約
sui move build

# 部署到測試網
sui client publish --gas-budget 20000000
```

## 事件

- `ProjectCreated`: 項目創建事件
- `DonationMade`: 捐贈完成事件  
- `FundsWithdrawn`: 資金提取事件

所有事件都會在區塊鏈上永久記錄，確保完全透明。
