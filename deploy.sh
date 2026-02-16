# 前端部署
cd web
npm run build
pkill -9 python3
nohup python3 -u -m http.server 80 > ../log/web.log 2>&1 &

# 后端部署
pkill -9 fund
nohup ./fund > log/fund.log 2>&1 &