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
	port := 8080

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
	log.Printf("🚀 服务器启动成功")
	log.Printf("📍 监听端口: %d", port)
	log.Printf("📡 基金详情: http://localhost:%d/api/fund/detail?code=001186", port)
	log.Printf("📈 走势数据: http://localhost:%d/api/fund/trend?code=001186&period=month", port)
	log.Printf("📊 日内数据: http://localhost:%d/api/fund/intraday?code=001186", port)
	log.Printf("📋 基金列表: http://localhost:%d/api/fund/list", port)
	log.Printf("🔧 服务状态: http://localhost:%d/api/status", port)
	log.Printf("❤️  健康检查: http://localhost:%d/health", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
