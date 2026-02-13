#!/bin/bash

echo "🔍 验证构建配置..."
echo ""

# 检查环境配置文件
echo "📋 环境配置文件："
echo "- .env.production:"
cat .env.production
echo ""
echo "- .env.local:"
cat .env.local
echo ""

# 构建项目
echo "🏗️  开始构建..."
npm run build
echo ""

# 检查构建产物
echo "✅ 验证构建产物中的 API 地址..."
if grep -q "175.27.141.110" dist/assets/*.js; then
    echo "✅ 成功！构建产物包含生产服务器地址: 175.27.141.110"
    echo ""
    echo "📝 找到的 API 地址："
    grep -o "http://175.27.141.110:8080/api" dist/assets/*.js | head -1
else
    echo "❌ 警告！构建产物中未找到服务器地址"
fi

echo ""
echo "📦 构建产物大小："
ls -lh dist/assets/*.js

echo ""
echo "🎉 验证完成！"
