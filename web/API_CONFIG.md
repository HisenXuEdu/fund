# API 配置说明

## 环境配置

项目使用 Vite 环境变量管理不同环境的 API 地址。

### 开发环境 (`.env.local`)

```bash
VITE_API_BASE_URL=/api
```

- 本地开发时使用
- 通过 Vite 代理转发到 `http://localhost:8080`
- 运行 `npm run dev` 时自动生效

### 生产环境 (`.env.production`)

```bash
VITE_API_BASE_URL=http://175.27.141.110:8080/api
```

- GitHub Pages 部署时使用
- 直接请求远程服务器 API
- 运行 `npm run build` 时自动注入

## Vite 配置 (`vite.config.ts`)

配置已优化，根据构建模式自动选择正确的 API 地址：

```typescript
const apiBaseUrl = mode === 'production' 
  ? 'http://175.27.141.110:8080/api'  // 生产环境
  : '/api';  // 开发环境
```

## 验证构建

### 方法 1：使用验证脚本

```bash
cd web
./verify-build.sh
```

### 方法 2：手动验证

```bash
cd web
npm run build

# 检查构建产物是否包含正确的 API 地址
grep -o "175.27.141.110" dist/assets/*.js
```

应该能看到输出：`175.27.141.110`

### 方法 3：本地预览

```bash
cd web
npm run build
npm run preview
```

打开浏览器，查看 Network 面板，确认 API 请求指向 `http://175.27.141.110:8080/api`

## 常见问题

### Q: 为什么本地开发时还是连不上后端？

**A:** 确保后端服务已启动：

```bash
# 在项目根目录
go run main.go
```

应该看到：`🚀 服务器启动成功，监听端口: 8080`

### Q: 生产环境还是使用 mock 数据？

**A:** 检查以下几点：

1. **后端服务是否运行**
   ```bash
   curl http://175.27.141.110:8080/health
   ```

2. **浏览器是否阻止混合内容**
   - GitHub Pages 使用 HTTPS
   - 后端使用 HTTP
   - 浏览器会阻止 HTTPS 页面请求 HTTP API

   **解决方案：为后端启用 HTTPS**

3. **检查浏览器控制台**
   - 打开 F12 开发者工具
   - 查看 Console 标签的错误信息
   - 查看 Network 标签的请求状态

### Q: 如何切换到不同的后端服务器？

**A:** 修改 `web/.env.production`：

```bash
# 修改为你的服务器地址
VITE_API_BASE_URL=http://your-server-ip:8080/api
```

然后重新构建：

```bash
cd web
npm run build
```

## 智能回退机制

前端已实现智能回退，当 API 请求失败时：

1. 首先尝试调用真实 API
2. 如果失败，自动使用本地模拟数据
3. 在控制台输出：`获取基金数据失败，使用模拟数据`

这确保了：
- **开发时**：即使后端未启动也能预览界面
- **生产时**：后端故障时前端仍可访问

## 部署流程

### 1. 本地测试

```bash
cd web
npm run dev          # 开发模式，连接本地后端
npm run build        # 生产构建
npm run preview      # 预览生产构建
```

### 2. 推送到 GitHub

```bash
git add .
git commit -m "Update API configuration"
git push origin main
```

### 3. 自动部署

GitHub Actions 会自动：
- 使用生产环境配置构建
- 部署到 GitHub Pages

### 4. 验证

访问你的 GitHub Pages 地址，打开浏览器控制台：

- ✅ 看到真实数据 → API 连接成功
- ⚠️ 看到 `使用模拟数据` → API 连接失败，检查后端

## 环境变量优先级

Vite 加载环境变量的优先级：

1. `.env.[mode].local` （最高优先级，git ignore）
2. `.env.[mode]` （如 `.env.production`）
3. `.env.local` （所有模式，git ignore）
4. `.env` （所有模式）

**建议：**
- 开发环境配置放在 `.env.local`（不提交到 git）
- 生产环境配置放在 `.env.production`（可提交）
