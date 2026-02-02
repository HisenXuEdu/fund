#!/bin/bash

# 基金走势API测试脚本

echo "=========================================="
echo "基金走势API测试"
echo "=========================================="
echo ""

BASE_URL="http://localhost:8080"
FUND_CODE="001186"

# 检查服务是否运行
echo "1. 检查服务健康状态..."
curl -s "${BASE_URL}/health" | python3 -m json.tool
echo ""

# 测试基金详情接口
echo "2. 获取基金详情..."
curl -s "${BASE_URL}/api/fund/detail?code=${FUND_CODE}" | python3 -m json.tool | head -20
echo ""

# 测试不同周期的走势数据
echo "3. 测试不同周期的走势数据..."
echo ""

periods=("week" "month" "quarter" "year")
period_names=("最近一周" "最近一个月" "最近一个季度" "最近一年")

for i in "${!periods[@]}"; do
    period="${periods[$i]}"
    period_name="${period_names[$i]}"
    
    echo "----------------------------------------"
    echo "测试周期: ${period_name} (${period})"
    echo "----------------------------------------"
    
    response=$(curl -s "${BASE_URL}/api/fund/trend?code=${FUND_CODE}&period=${period}")
    
    # 使用Python解析JSON并计算收益率
    python3 - <<EOF
import json
data = json.loads('''${response}''')

if 'error' in data:
    print(f"错误: {data['error']}")
else:
    name = data['name']
    code = data['code']
    data_points = data['data']
    
    if len(data_points) >= 2:
        first_value = data_points[0]['value']
        last_value = data_points[-1]['value']
        first_date = data_points[0]['date']
        last_date = data_points[-1]['date']
        
        return_rate = (last_value - first_value) / first_value * 100
        
        print(f"基金名称: {name}")
        print(f"基金代码: {code}")
        print(f"数据点数: {len(data_points)}")
        print(f"起始日期: {first_date}")
        print(f"结束日期: {last_date}")
        print(f"起始净值: {first_value:.4f}")
        print(f"最新净值: {last_value:.4f}")
        print(f"期间收益: {return_rate:+.2f}%")
    else:
        print("数据点不足")
EOF
    echo ""
done

# 测试错误处理
echo "=========================================="
echo "4. 测试错误处理..."
echo "=========================================="
echo ""

echo "测试1: 无效的基金代码"
curl -s "${BASE_URL}/api/fund/trend?code=123&period=month" | python3 -m json.tool
echo ""

echo "测试2: 无效的周期参数"
curl -s "${BASE_URL}/api/fund/trend?code=${FUND_CODE}&period=invalid" | python3 -m json.tool
echo ""

echo "测试3: 缺少必填参数"
curl -s "${BASE_URL}/api/fund/trend?period=month" | python3 -m json.tool
echo ""

echo "=========================================="
echo "测试完成!"
echo "=========================================="
