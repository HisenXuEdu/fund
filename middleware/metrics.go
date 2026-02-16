package middleware

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics 中间件：记录请求指标
func Metrics(next http.HandlerFunc, requestsTotal *prometheus.CounterVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端 IP
		ip := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			ip = strings.Split(forwardedFor, ",")[0]
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}
		// 去掉端口号
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}

		// 记录请求
		requestsTotal.WithLabelValues(r.Method, r.URL.Path, ip).Inc()

		// 继续处理请求
		next(w, r)
	}
}
