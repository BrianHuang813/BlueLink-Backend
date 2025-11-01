#!/bin/bash

# BlueLink 債券市場 API 測試腳本

# 顏色輸出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API Base URL
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
echo -e "${YELLOW}Testing API at: ${BASE_URL}${NC}\n"

# 測試計數器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 測試函數
test_api() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local expected_code=$4
    local data=$5
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${YELLOW}Test ${TOTAL_TESTS}: ${test_name}${NC}"
    
    # 構建 curl 命令
    if [ "$method" == "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}${endpoint}" \
            -H "Accept: application/json")
    elif [ "$method" == "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -H "Accept: application/json" \
            -d "$data")
    fi
    
    # 分離響應體和狀態碼
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    # 檢查狀態碼
    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ Status Code: ${http_code}${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ Status Code: ${http_code} (expected ${expected_code})${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    # 顯示響應體 (格式化 JSON)
    if command -v jq &> /dev/null; then
        echo "$body" | jq '.'
    else
        echo "$body"
    fi
    
    echo ""
}

# 測試健康檢查
echo -e "${YELLOW}=== Health Check ===${NC}\n"
test_api "Health Check" "GET" "/health" "200"

# 測試債券 API
echo -e "${YELLOW}=== Bond API Tests ===${NC}\n"

# 1. 獲取所有債券
test_api "Get All Bonds" "GET" "/api/v1/bonds" "200"

# 2. 獲取單個債券 (假設 ID 1 存在)
test_api "Get Bond by ID" "GET" "/api/v1/bonds/1" "200"

# 3. 獲取不存在的債券
test_api "Get Non-existent Bond" "GET" "/api/v1/bonds/99999" "404"

# 4. 測試同步端點 (需要認證,預期 401)
echo -e "${YELLOW}=== Sync Transaction (without auth) ===${NC}\n"
test_api "Sync Transaction Without Auth" "POST" "/api/v1/bonds/sync" "401" \
    '{"transaction_digest":"ABC123","event_type":"bond_created"}'

# 顯示測試結果摘要
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed! ✗${NC}"
    exit 1
fi
