package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"
)

// TestFund123CSRF 测试获取CSRF令牌
func TestFund123CSRF(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("https://www.fund123.cn/fund")
	if err != nil {
		t.Fatalf("❌ 请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ 读取响应失败: %v", err)
	}

	// 提取CSRF令牌
	re := regexp.MustCompile(`"csrf":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		t.Fatalf("❌ 未找到CSRF令牌")
	}

	csrfToken := matches[1]
	t.Logf("✅ 获取CSRF令牌成功: %s", csrfToken[:10]+"...")
}

// TestFund123SearchFund 测试搜索基金 (使用 searchFund 接口)
func TestFund123SearchFund(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 先获取CSRF令牌
	resp, err := client.Get("https://www.fund123.cn/fund")
	if err != nil {
		t.Fatalf("❌ 获取CSRF令牌失败: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	re := regexp.MustCompile(`"csrf":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		t.Fatalf("❌ 未找到CSRF令牌")
	}
	csrfToken := matches[1]
	t.Logf("✅ CSRF令牌: %s", csrfToken[:10]+"...")

	// 测试多个基金代码
	testCodes := []string{
		"161005", // 富国天惠成长混合A
		"001186", // 富国文体健康股票
		"110022", // 易方达消费行业股票
		"025857", // 之前报错的基金代码
	}

	for _, fundCode := range testCodes {
		t.Run(fmt.Sprintf("searchFund_%s", fundCode), func(t *testing.T) {
			url := fmt.Sprintf("https://www.fund123.cn/api/fund/searchFund?_csrf=%s", csrfToken)
			
			reqBody := fmt.Sprintf(`{"fundCode":"%s"}`, fundCode)
			req, err := http.NewRequest("POST", url, bytes.NewBufferString(reqBody))
			if err != nil {
				t.Fatalf("❌ 创建请求失败: %v", err)
			}

			req.Header.Set("Accept", "json")
			req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "https://www.fund123.cn")
			req.Header.Set("Referer", "https://www.fund123.cn/fund")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
			req.Header.Set("X-API-Key", "foobar")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("❌ 搜索请求失败: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("❌ 读取响应失败: %v", err)
			}

			t.Logf("📄 响应内容: %s", string(body))

			var result struct {
				Success  bool `json:"success"`
				FundInfo struct {
					Key      string `json:"key"`
					FundName string `json:"fundName"`
				} `json:"fundInfo"`
				Message string `json:"message"`
			}

			if err := json.Unmarshal(body, &result); err != nil {
				t.Logf("⚠️  解析JSON失败: %v", err)
				return
			}

			if !result.Success {
				t.Logf("⚠️  基金 %s 查询失败: %s", fundCode, result.Message)
				return
			}

			t.Logf("✅ 找到基金: %s (fundKey: %s)", result.FundInfo.FundName, result.FundInfo.Key)
		})
	}
}

// TestFund123IntradayData 测试获取分时数据 (使用 queryFundEstimateIntraday 接口)
func TestFund123IntradayData(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 先获取CSRF令牌
	resp, err := client.Get("https://www.fund123.cn/fund")
	if err != nil {
		t.Fatalf("❌ 获取CSRF令牌失败: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	re := regexp.MustCompile(`"csrf":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		t.Fatalf("❌ 未找到CSRF令牌")
	}
	csrfToken := matches[1]

	// 测试基金代码
	fundCode := "161005"

	// 1. 先获取 fundKey
	searchURL := fmt.Sprintf("https://www.fund123.cn/api/fund/searchFund?_csrf=%s", csrfToken)
	reqBody := fmt.Sprintf(`{"fundCode":"%s"}`, fundCode)
	req, _ := http.NewRequest("POST", searchURL, bytes.NewBufferString(reqBody))
	req.Header.Set("Accept", "json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.fund123.cn")
	req.Header.Set("Referer", "https://www.fund123.cn/fund")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("X-API-Key", "foobar")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("❌ 搜索基金失败: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	var searchResult struct {
		Success  bool `json:"success"`
		FundInfo struct {
			Key      string `json:"key"`
			FundName string `json:"fundName"`
		} `json:"fundInfo"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		t.Fatalf("❌ 解析搜索结果失败: %v", err)
	}

	if !searchResult.Success {
		t.Fatalf("❌ 未找到基金 %s", fundCode)
	}

	fundKey := searchResult.FundInfo.Key
	fundName := searchResult.FundInfo.FundName
	t.Logf("✅ 找到基金: %s (%s), fundKey: %s", fundName, fundCode, fundKey)

	// 2. 获取分时数据
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	
	intradayURL := "https://www.fund123.cn/api/fund/queryFundEstimateIntraday"
	payload := fmt.Sprintf(`{"startTime":"%s","endTime":"%s","limit":200,"productId":"%s","format":true,"source":"WEALTHBFFWEB"}`, 
		today, tomorrow, fundKey)
	
	req, _ = http.NewRequest("POST", intradayURL, bytes.NewBufferString(payload))
	req.Header.Set("Accept", "json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.fund123.cn")
	req.Header.Set("Referer", "https://www.fund123.cn/fund")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("X-API-Key", "foobar")
	
	q := req.URL.Query()
	q.Add("_csrf", csrfToken)
	req.URL.RawQuery = q.Encode()

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("❌ 获取分时数据失败: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	t.Logf("📊 分时数据响应: %s", string(body)[:min(len(body), 200)]+"...")

	var intradayResult struct {
		Success bool `json:"success"`
		List    []struct {
			Time             int64   `json:"time"`
			ForecastGrowth   float64 `json:"forecastGrowth"`
			ForecastNetValue float64 `json:"forecastNetValue"`
		} `json:"list"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &intradayResult); err != nil {
		t.Fatalf("❌ 解析分时数据失败: %v", err)
	}

	if !intradayResult.Success {
		t.Logf("⚠️  获取分时数据失败: %s", intradayResult.Message)
		return
	}

	if len(intradayResult.List) == 0 {
		t.Logf("⚠️  基金 %s 暂无分时数据（可能非交易时间）", fundCode)
		return
	}

	t.Logf("✅ 获取到 %d 个分时数据点", len(intradayResult.List))
	
	// 显示最后一个数据点
	latest := intradayResult.List[len(intradayResult.List)-1]
	latestTime := time.Unix(latest.Time/1000, 0).Format("15:04")
	t.Logf("📈 最新数据: 时间=%s, 净值=%.4f, 涨幅=%.2f%%",
		latestTime,
		latest.ForecastNetValue,
		latest.ForecastGrowth*100)
}

// TestIntradayServiceGetData 测试 IntradayService 的实时获取方法
func TestIntradayServiceGetData(t *testing.T) {
	service := NewIntradayService()

	// 初始化服务
	if err := service.Start(); err != nil {
		t.Fatalf("❌ 启动服务失败: %v", err)
	}

	// 测试多个基金代码
	testCodes := []string{
		"161005", // 富国天惠成长混合A
		"001186", // 富国文体健康股票
		"110022", // 易方达消费行业股票
	}

	for _, fundCode := range testCodes {
		t.Run(fmt.Sprintf("获取分时数据_%s", fundCode), func(t *testing.T) {
			data, err := service.GetIntradayDataRealtime(fundCode)
			if err != nil {
				t.Logf("⚠️  获取基金 %s 失败: %v", fundCode, err)
				return
			}

			t.Logf("✅ 基金: %s (%s)", data.Name, data.Code)
			t.Logf("📅 日期: %s", data.Date)
			t.Logf("📊 数据点数: %d", len(data.Data))

			if len(data.Data) > 0 {
				latest := data.Data[len(data.Data)-1]
				t.Logf("📈 最新: 时间=%s, 净值=%.4f, 涨幅=%.2f%%",
					latest.Time, latest.Value, latest.Rate)
			}
		})
	}
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
