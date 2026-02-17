#!/bin/bash
# 审计日志性能测试脚本
# 用于模拟大量审计日志并测试系统性能

set -e

# 配置
BASE_URL="http://localhost:8100/api/audit"
ADMIN_TOKEN=""  # 需要替换为实际的管理员 token
TEST_COUNT=100  # 生成测试数据的数量
THREAD_COUNT=10  # 并发线程数
BATCH_SIZE=1000  # 批量导出数量

echo "=== 审计日志性能测试 ==="
echo "基础URL: $BASE_URL"
echo "测试数据量: $TEST_COUNT"
echo "并发线程数: $THREAD_COUNT"
echo "批量导出数量: $BATCH_SIZE"
echo ""

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 检查必要参数
if [ -z "$ADMIN_TOKEN" ]; then
  echo -e "${RED}错误: 请设置 ADMIN_TOKEN${NC}"
  echo "用法: ADMIN_TOKEN=<your_token> $0"
  exit 1
fi

# 测试1: 查询性能测试
echo -e "${GREEN}[测试1] 查询性能测试${NC}"
echo "测试场景: 查询不同页面的性能"

start_time=$(date +%s%N)

# 测试分页查询
echo "  - 测试第1页 (10条)..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?page=1&pageSize=10" > /dev/null

echo "  - 测试第10页 (100条)..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?page=10&pageSize=10" > /dev/null

echo "  - 测试第100页 (1000条)..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?page=100&pageSize=10" > /dev/null

# 测试带筛选条件的查询
echo "  - 测试按操作类型筛选..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?actionType=USER_ACTION&page=1&pageSize=50" > /dev/null

echo "  - 测试按日期范围筛选..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?startTime=2025-01-01 00:00:00&endTime=2025-01-31 23:59:59&page=1&pageSize=50" > /dev/null

end_time=$(date +%s%N)
query_duration=$((end_time - start_time))
echo "查询性能测试完成，耗时: $query_duration 秒"
echo ""

# 测试2: 批量导出性能测试
echo -e "${GREEN}[测试2] 批量导出性能测试${NC}"
echo "测试场景: 导出大量审计日志的性能"

start_time=$(date +%s%N)

echo "  - 导出前10条日志..."
time curl -s -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "[1,2,3,4,5,6,7,8,9,10]" \
  "$BASE_URL/export?format=csv" > /dev/null

echo "  - 导出前50条日志..."
time curl -s -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "[$(seq 1 50 | tr '\n' ',' | sed 's/ //g')]" \
  "$BASE_URL/export?format=json" > /dev/null

end_time=$(date +%s%N)
export_duration=$((end_time - start_time))
echo "导出性能测试完成，耗时: $export_duration 秒"
echo ""

# 测试3: 高并发压力测试
echo -e "${GREEN}[测试3] 高并发压力测试${NC}"
echo "测试场景: 模拟多个用户同时查询审计日志"

start_time=$(date +%s%N)

echo "  - 启动 $THREAD_COUNT 个并发查询..."

for i in $(seq 1 $THREAD_COUNT); do
  curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
    "$BASE_URL/list?page=1&pageSize=10" > /dev/null &
done

wait

end_time=$(date +%s%N)
concurrent_duration=$((end_time - start_time))
echo "并发测试完成，耗时: $concurrent_duration 秒"
echo ""

# 测试4: 复杂查询性能测试
echo -e "${GREEN}[测试4] 复杂查询性能测试${NC}"
echo "测试场景: 多条件组合查询"

start_time=$(date +%s%N)

echo "  - 测试多条件组合查询..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?userId=1&actionType=USER_ACTION&resourceType=USER&startTime=2025-01-01 00:00:00&endTime=2025-01-31 23:59:59" > /dev/null

echo "  - 测试按用户ID查询..."
time curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL/list?userId=1&page=1&pageSize=100" > /dev/null

end_time=$(date +%s%N)
complex_duration=$((end_time - start_time))
echo "复杂查询测试完成，耗时: $complex_duration 秒"
echo ""

# 测试5: 系统稳定性测试
echo -e "${GREEN}[测试5] 系统稳定性测试${NC}"
echo "测试场景: 持续发送请求测试系统稳定性"

start_time=$(date +%s%N)
error_count=0

echo "  - 连续发送 100 次查询请求..."
for i in $(seq 1 100); do
  response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ADMIN_TOKEN" \
    "$BASE_URL/list?page=1&pageSize=10")

  if [ "$response" != "200" ]; then
    error_count=$((error_count + 1))
  fi
done

end_time=$(date +%s%N)
stability_duration=$((end_time - start_time))
echo "稳定性测试完成，耗时: $stability_duration 秒，错误数: $error_count"
echo ""

# 性能测试总结
echo "=== 性能测试总结 ==="
echo -e "${GREEN}所有测试已完成${NC}"
echo ""
echo "性能指标:"
echo "  - 查询性能: $query_duration 秒"
echo "  - 导出性能: $export_duration 秒"
echo "  - 并发性能: $concurrent_duration 秒"
echo "  - 复杂查询性能: $complex_duration 秒"
echo "  - 稳定性测试: $stability_duration 秒"
echo ""
echo "性能基准:"
echo "  - 查询响应时间 < 500ms: 优秀"
echo "  - 查询响应时间 500-1000ms: 良好"
echo "  - 查询响应时间 1000-2000ms: 一般"
echo "  - 查询响应时间 > 2000ms: 需要优化"
echo ""
echo "建议:"
if [ $query_duration -gt 2000 ]; then
  echo -e "${YELLOW}  - 查询性能较慢，建议检查数据库索引${NC}"
else
  echo -e "${GREEN}  - 查询性能良好${NC}"
fi

if [ $error_count -gt 5 ]; then
  echo -e "${YELLOW}  - 系统稳定性不佳，发现 $error_count 个错误${NC}"
else
  echo -e "${GREEN}  - 系统稳定性良好${NC}"
fi

echo ""
echo "=== 测试结束 ==="
