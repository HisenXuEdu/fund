import path from 'path';
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
    // 从当前目录加载环境变量
    const env = loadEnv(mode, process.cwd(), '');
    
    // 根据模式确定 API 地址
    const apiBaseUrl = mode === 'production' 
      ? 'http://175.27.141.110:8080/api'  // 生产环境：通过服务器IP访问后端
      : '/api';  // 开发环境：使用代理
    
    console.log(`[Vite Config] Mode: ${mode}, API Base URL: ${apiBaseUrl}`);
    
    return {
      base: '/',
      server: {
        port: 3000,
        host: '0.0.0.0',
        proxy: {
          // 开发环境代理到Go后端
          '/api': {
            target: 'http://localhost:8080',
            changeOrigin: true,
          }
        }
      },
      plugins: [react()],
      define: {
        // 直接注入 API 地址，确保构建时正确替换
        'import.meta.env.VITE_API_BASE_URL': JSON.stringify(apiBaseUrl)
      },
      resolve: {
        alias: {
          '@': path.resolve(__dirname, '.'),
        }
      }
    };
});
