package main

import (
	"fmt"
	"fund/handler"
	"fund/router"
	"fund/service"
	"log"
	"net/http"
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

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
