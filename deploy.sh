#!/bin/bash

# 前端部署
echo "================================"
echo "📦 开始前端部署..."
echo "================================"
cd web
npm run build

echo "🛑 停止旧的前端服务..."
pkill python3 || true
sleep 1
if pgrep python3 > /dev/null 2>&1; then
    echo "⚠️  前端进程未正常退出，强制终止"
    pkill -9 python3
fi

echo "🚀 启动前端服务..."
nohup python3 -u -m http.server 80 > ../log/web.log 2>&1 &

# 后端部署
echo ""
echo "================================"
echo "📦 开始后端部署..."
echo "================================"
cd ../

echo "🛑 停止旧的后端服务..."
pkill fund || true
sleep 2
if pgrep fund > /dev/null 2>&1; then
    echo "⚠️  后端进程未正常退出，强制终止"
    pkill -9 fund
fi

echo "🚀 启动后端服务..."
nohup ./fund > log/fund.log 2>&1 &

echo ""
echo "✅ 部署完成！"
echo "前端: http://http://175.27.141.110"
echo "后端: http://http://175.27.141.110:8080"