package handler

import (
	"encoding/json"
	"fund/service"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// FundHandler 基金处理器
type FundHandler struct {
	fundService     *service.FundService
	intradayService *service.IntradayService
}

// NewFundHandler 创建基金处理器实例
func NewFundHandler(fundService *service.FundService, intradayService *service.IntradayService) *FundHandler {
	return &FundHandler{
		fundService:     fundService,
		intradayService: intradayService,
	}
}

// GetFundDetail 获取基金详情接口
func (h *FundHandler) GetFundDetail(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 获取基金代码参数
	fundCode := r.URL.Query().Get("code")
	if fundCode == "" {
		h.responseError(w, http.StatusBadRequest, "请提供基金代码参数 code")
		return
	}

	// 验证基金代码格式(6位数字)
	if !h.isValidFundCode(fundCode) {
		h.responseError(w, http.StatusBadRequest, "基金代码格式错误,应为6位数字")
		return
	}

	// 获取基金详情
	fundDetail, err := h.fundService.GetFundDetail(fundCode)
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回成功响应
	h.responseSuccess(w, fundDetail)
}

// GetFundTrend 获取基金走势接口
func (h *FundHandler) GetFundTrend(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 获取基金代码参数
	fundCode := r.URL.Query().Get("code")
	if fundCode == "" {
		h.responseError(w, http.StatusBadRequest, "请提供基金代码参数 code")
		return
	}

	// 验证基金代码格式
	if !h.isValidFundCode(fundCode) {
		h.responseError(w, http.StatusBadRequest, "基金代码格式错误,应为6位数字")
		return
	}

	// 获取周期参数
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // 默认一个月
	}

	// 验证周期参数
	validPeriods := map[string]bool{
		"week":        true,
		"month":       true,
		"quarter":     true,
		"half_year":   true,
		"year":        true,
		"three_years": true,
		"all":         true,
	}
	if !validPeriods[period] {
		h.responseError(w, http.StatusBadRequest, "周期参数无效,可选值: week/month/quarter/half_year/year/three_years/all")
		return
	}

	// 获取基金走势
	fundTrend, err := h.fundService.GetFundTrend(fundCode, period)
	if err != nil {
		h.responseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回成功响应
	h.responseSuccess(w, fundTrend)
}

// GetIntradayData 获取基金日内实时数据接口
func (h *FundHandler) GetIntradayData(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 获取基金代码参数
	fundCode := r.URL.Query().Get("code")
	if fundCode == "" {
		h.responseError(w, http.StatusBadRequest, "请提供基金代码参数 code")
		return
	}

	// 验证基金代码格式
	if !h.isValidFundCode(fundCode) {
		h.responseError(w, http.StatusBadRequest, "基金代码格式错误,应为6位数字")
		return
	}

	// 获取日内数据
	intradayData, err := h.intradayService.GetIntradayData(fundCode)
	if err != nil {
		h.responseError(w, http.StatusNotFound, err.Error())
		return
	}

	// 返回成功响应
	h.responseSuccess(w, intradayData)
}

// GetFundList 获取基金列表接口
func (h *FundHandler) GetFundList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	fundList := h.intradayService.GetFundList()
	
	// 支持分页
	page := 1
	pageSize := 100
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("pageSize"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 1000 {
			pageSize = s
		}
	}

	// 计算分页
	total := len(fundList)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		start = 0
		end = 0
	} else if end > total {
		end = total
	}

	response := map[string]interface{}{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     fundList[start:end],
	}

	h.responseSuccess(w, response)
}

// GetServiceStatus 获取服务状态接口
func (h *FundHandler) GetServiceStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	
	status := map[string]interface{}{
		"status":      "running",
		"fundCount":   len(h.intradayService.GetFundList()),
		"dataCount":   h.intradayService.GetDataCount(),
		"currentTime": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	h.responseSuccess(w, status)
}

// Health 健康检查接口
func (h *FundHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// isValidFundCode 验证基金代码格式
func (h *FundHandler) isValidFundCode(code string) bool {
	matched, _ := regexp.MatchString(`^\d{6}$`, code)
	return matched
}

// responseError 返回错误响应
func (h *FundHandler) responseError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// responseSuccess 返回成功响应
func (h *FundHandler) responseSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
