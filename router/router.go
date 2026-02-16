package router

import (
	"fund/handler"
	"fund/middleware"
	"fund/stats/metrics"
	"net/http"
	// "go.uber.org/zap"
)

var (
	requestsTotal = metrics.MustNewCounterVec("router", "requests_total", []string{"method", "path", "ip"})
)

// SetupRoutes 设置路由
func SetupRoutes(fundHandler *handler.FundHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// 包装中间件：CORS + Metrics
	wrap := func(handler http.HandlerFunc) http.HandlerFunc {
		return middleware.CORS(middleware.Metrics(handler, requestsTotal))
	}

	// 基金详情API
	mux.HandleFunc("/api/fund/detail", wrap(fundHandler.GetFundDetail))
	mux.HandleFunc("/api/fund/trend", wrap(fundHandler.GetFundTrend))

	// 日内实时数据API
	mux.HandleFunc("/api/fund/intraday", wrap(fundHandler.GetIntradayData))
	mux.HandleFunc("/api/fund/list", wrap(fundHandler.GetFundList))

	// 服务状态
	mux.HandleFunc("/api/status", wrap(fundHandler.GetServiceStatus))

	// 健康检查（不需要 metrics）
	mux.HandleFunc("/health", fundHandler.Health)

	return mux
}
