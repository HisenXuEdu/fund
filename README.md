# 基金监控系统

一个基于 Go 后端 + React 前端的实时基金数据监控系统，支持查看基金详情、历史走势和日内实时数据。

体验地址：http://175.27.141.110 （个人服务器，请轻点使用）
![alt text](/img/Clipboard_Screenshot_1772372199.png)

## ✨ 功能特性

- 📈 **走势分析**：支持查看日/周/月/季度/年度历史走势
- 📋 **基金列表**：管理自选基金列表
- 📡 **RESTful API**：提供完整的 HTTP API 接口
- 📊 **Prometheus 监控**：内置 metrics 接口（端口 8081）

## 🛠️ 技术栈

**后端**
- Go 1.23+
- Prometheus client

**前端**
- React + TypeScript
- Vite 构建工具

## 📦 部署方法

### 前置要求

- Go 1.23 或更高版本
- Node.js 和 npm
- Python 3（用于前端静态文件服务）

### 快速部署

使用一键部署脚本：

```bash
# 赋予执行权限
chmod +x deploy.sh

# 执行部署
./deploy.sh
```

部署脚本会自动完成以下操作：

1. **前端部署**
   - 构建 React 应用（`npm run build`）
   - 停止旧的前端服务进程
   - 启动新的静态文件服务（端口 80）

2. **后端部署**
   - 编译 Go 程序（`go build -o fund`）
   - 停止旧的后端服务进程
   - 启动新的后端服务（端口 8080）

### 手动部署

#### 后端部署

```bash
# 编译
go build -o fund

# 运行
./fund
```

#### 前端部署

```bash
cd web
npm install
npm run build

# 启动静态文件服务
python3 -m http.server 80
```

## 🚀 使用说明

### 服务地址

- 前端：`http://your-ip`
- 后端 API：`http://your-ip:8080`
- Prometheus 监控：`http://your-ip:8081/metrics`

### API 接口

| 端点 | 说明 | 示例 |
|------|------|------|
| `/api/fund/detail?code={code}` | 获取基金详情 | `/api/fund/detail?code=001186` |
| `/api/fund/trend?code={code}&period={period}` | 获取走势数据 | `/api/fund/trend?code=001186&period=month` |
| `/api/fund/intraday?code={code}` | 获取日内实时数据 | `/api/fund/intraday?code=001186` |
| `/api/fund/list` | 获取自选基金列表 | `/api/fund/list` |
| `/api/status` | 服务状态 | `/api/status` |
| `/health` | 健康检查 | `/health` |

**period 参数**：`day`（日）、`week`（周）、`month`（月）、`quarter`（季）、`year`（年）

## 📝 日志

- 后端日志：`log/fund.log`
- 前端日志：`log/web.log`