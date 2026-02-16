package main

import (
	"context"
	"fmt"
	"fund/handler"
	"fund/router"
	"fund/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 配置
	host := "0.0.0.0" // 监听所有接口
	port := 8080
	serverIP := "175.27.141.110"

	// 初始化服务层
	fundService := service.NewFundService()
	intradayService := service.NewIntradayService()

	// 启动日内实时数据采集服务
	if err := intradayService.Start(); err != nil {
		log.Fatalf("❌ 启动实时数据服务失败: %v", err)
	}

	// 初始化处理器层
	fundHandler := handler.NewFundHandler(fundService, intradayService)

	// 设置路由
	mux := router.SetupRoutes(fundHandler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 创建信号通道用于优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在新的 goroutine 中启动服务器
	go func() {
		log.Printf("🚀 服务器启动成功")
		log.Printf("📍 监听地址: %s", addr)
		log.Printf("🌐 外网访问: http://%s:%d", serverIP, port)
		log.Printf("")
		log.Printf("API 端点:")
		log.Printf("📡 基金详情: http://%s:%d/api/fund/detail?code=001186", serverIP, port)
		log.Printf("📈 走势数据: http://%s:%d/api/fund/trend?code=001186&period=month", serverIP, port)
		log.Printf("📊 日内数据: http://%s:%d/api/fund/intraday?code=001186", serverIP, port)
		log.Printf("📋 基金列表: http://%s:%d/api/fund/list", serverIP, port)
		log.Printf("🔧 服务状态: http://%s:%d/api/status", serverIP, port)
		log.Printf("❤️  健康检查: http://%s:%d/health", serverIP, port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待退出信号
	<-quit

	// 创建超时上下文（30秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止日内数据采集服务
	log.Println("📊 停止实时数据采集服务...")
	intradayService.Stop()

	// 保存数据到磁盘
	log.Println("💾 保存数据到磁盘...")
	if err := intradayService.SaveData(); err != nil {
		log.Printf("⚠️  保存数据失败: %v", err)
	}

	// 关闭 HTTP 服务器
	log.Println("🌐 关闭 HTTP 服务器...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("⚠️  服务器关闭失败: %v", err)
	}
}
