npm run build
pkill -9 python3
nohup python3 -u -m http.server 80 > ./log_frontend.log 2>&1 &

nohup ./fund > logs/log_fund.log 2>&1 &