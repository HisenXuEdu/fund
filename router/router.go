package router

import (
	"fund/handler"
	"fund/middleware"
	"net/http"
)

// SetupRoutes 设置路由
func SetupRoutes(fundHandler *handler.FundHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// 基金详情API
	mux.HandleFunc("/api/fund/detail", middleware.CORS(fundHandler.GetFundDetail))
	mux.HandleFunc("/api/fund/trend", middleware.CORS(fundHandler.GetFundTrend))
	
	// 日内实时数据API
	mux.HandleFunc("/api/fund/intraday", middleware.CORS(fundHandler.GetIntradayData))
	mux.HandleFunc("/api/fund/list", middleware.CORS(fundHandler.GetFundList))
	
	// 服务状态
	mux.HandleFunc("/api/status", middleware.CORS(fundHandler.GetServiceStatus))
	
	// 健康检查
	mux.HandleFunc("/health", fundHandler.Health)

	return mux
}
