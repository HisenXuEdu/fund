package service

import (
	"encoding/json"
	"fmt"
	"fund/model"
	"io"
	"net/http"
	"regexp"
	"time"
)

// FundService 基金服务
type FundService struct {
	httpClient *http.Client
}

// NewFundService 创建基金服务实例
func NewFundService() *FundService {
	return &FundService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetFundDetail 获取基金详细信息
func (s *FundService) GetFundDetail(fundCode string) (*model.FundDetail, error) {
	// 获取基金详情
	detailData, err := s.fetchFundDetail(fundCode)
	if err != nil {
		return nil, fmt.Errorf("获取基金详情失败: %v", err)
	}

	// 获取实时估值
	realtimeData, err := s.fetchRealtimeData(fundCode)
	if err != nil {
		return nil, fmt.Errorf("获取实时估值失败: %v", err)
	}

	// 组合返回数据
	fundDetail := &model.FundDetail{
		Code:          fundCode,
		Name:          detailData["name"],
		CurrentPrice:  detailData["currentPrice"],
		EstimatePrice: realtimeData.Gsz,
		EstimateRate:  realtimeData.GsZzl,
		UpdateTime:    realtimeData.Gztime,
		DayGrowth:     detailData["dayGrowth"],
		WeekGrowth:    detailData["weekGrowth"],
		MonthGrowth:   detailData["monthGrowth"],
		ThreeMonth:    detailData["threeMonth"],
		SixMonth:      detailData["sixMonth"],
		YearGrowth:    detailData["yearGrowth"],
		TotalGrowth:   detailData["totalGrowth"],
	}

	// 如果实时数据有名称,优先使用
	if realtimeData.Name != "" {
		fundDetail.Name = realtimeData.Name
	}

	// 如果实时数据有净值,优先使用
	if realtimeData.DwJz != "" {
		fundDetail.CurrentPrice = realtimeData.DwJz
	}

	return fundDetail, nil
}

// fetchFundDetail 获取基金详情数据
func (s *FundService) fetchFundDetail(fundCode string) (map[string]string, error) {
	timestamp := time.Now().UnixNano() / 1e6
	url := fmt.Sprintf("http://fund.eastmoney.com/pingzhongdata/%s.js?v=%d", fundCode, timestamp)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.parseFundDetailJS(string(body))
}

// fetchRealtimeData 获取实时估值数据
func (s *FundService) fetchRealtimeData(fundCode string) (*model.RealtimeData, error) {
	timestamp := time.Now().UnixNano() / 1e6
	url := fmt.Sprintf("http://fundgz.1234567.com.cn/js/%s.js?rt=%d", fundCode, timestamp)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.parseRealtimeJS(string(body))
}

// parseFundDetailJS 解析东方财富基金详情JS
func (s *FundService) parseFundDetailJS(jsContent string) (map[string]string, error) {
	result := make(map[string]string)

	// 提取基金名称
	if name := s.extractPattern(jsContent, `var fS_name = "([^"]+)"`); name != "" {
		result["name"] = name
	}

	// 提取基金代码
	if code := s.extractPattern(jsContent, `var fS_code = "([^"]+)"`); code != "" {
		result["code"] = code
	}

	// 提取增长率数据 (Data_netWorthTrend)
	if trendMatch := s.extractPattern(jsContent, `var Data_netWorthTrend = (\[.*?\]);`); trendMatch != "" {
		var trendData [][]interface{}
		jsonStr := regexp.MustCompile(`'`).ReplaceAllString(trendMatch, `"`)
		if err := json.Unmarshal([]byte(jsonStr), &trendData); err == nil && len(trendData) > 0 {
			lastData := trendData[len(trendData)-1]
			if len(lastData) >= 2 {
				result["currentPrice"] = fmt.Sprintf("%.4f", lastData[1])
			}
		}
	}

	// 提取涨跌幅数据 (Data_rateInSimilarPersent)
	if rateMatch := s.extractPattern(jsContent, `var Data_rateInSimilarPersent = \[(.*?)\];`); rateMatch != "" {
		rates := regexp.MustCompile(`[",]`).Split(rateMatch, -1)
		var cleanRates []string
		for _, rate := range rates {
			if rate != "" {
				cleanRates = append(cleanRates, rate)
			}
		}
		if len(cleanRates) >= 7 {
			result["dayGrowth"] = cleanRates[0]
			result["weekGrowth"] = cleanRates[1]
			result["monthGrowth"] = cleanRates[2]
			result["threeMonth"] = cleanRates[3]
			result["sixMonth"] = cleanRates[4]
			result["yearGrowth"] = cleanRates[5]
		}
	}

	// 提取累计收益率 (Data_grandTotal)
	if totalMatch := s.extractPattern(jsContent, `var Data_grandTotal = \[(.*?)\];`); totalMatch != "" {
		var totalData [][]interface{}
		jsonStr := regexp.MustCompile(`'`).ReplaceAllString(totalMatch, `"`)
		if err := json.Unmarshal([]byte(jsonStr), &totalData); err == nil && len(totalData) > 0 {
			lastData := totalData[len(totalData)-1]
			if len(lastData) >= 2 {
				result["totalGrowth"] = fmt.Sprintf("%.2f", lastData[1])
			}
		}
	}

	return result, nil
}

// parseRealtimeJS 解析实时估值JS
func (s *FundService) parseRealtimeJS(jsContent string) (*model.RealtimeData, error) {
	// 去除jsonpgz()包裹
	re := regexp.MustCompile(`jsonpgz\((.*?)\);?$`)
	matches := re.FindStringSubmatch(jsContent)
	if len(matches) < 2 {
		return nil, fmt.Errorf("无法解析实时数据")
	}

	var realtimeData model.RealtimeData
	if err := json.Unmarshal([]byte(matches[1]), &realtimeData); err != nil {
		return nil, err
	}

	return &realtimeData, nil
}

// extractPattern 提取正则匹配的第一个分组
func (s *FundService) extractPattern(content, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetFundTrend 获取基金走势数据
func (s *FundService) GetFundTrend(fundCode, period string) (*model.FundTrend, error) {
	// 获取基金详情数据
	timestamp := time.Now().UnixNano() / 1e6
	url := fmt.Sprintf("http://fund.eastmoney.com/pingzhongdata/%s.js?v=%d", fundCode, timestamp)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取基金数据失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取基金数据失败: %v", err)
	}

	jsContent := string(body)

	// 提取基金名称
	fundName := s.extractPattern(jsContent, `var fS_name = "([^"]+)"`)

	// 提取净值走势数据
	trendData, err := s.extractNetWorthTrend(jsContent)
	if err != nil {
		return nil, fmt.Errorf("解析走势数据失败: %v", err)
	}

	// 根据周期类型过滤数据
	filteredData := s.filterByPeriod(trendData, period)

	return &model.FundTrend{
		Code:   fundCode,
		Name:   fundName,
		Period: period,
		Data:   filteredData,
	}, nil
}

// extractNetWorthTrend 提取净值走势数据
func (s *FundService) extractNetWorthTrend(jsContent string) ([]model.TrendPoint, error) {
	// 提取 Data_netWorthTrend 数组
	pattern := `var Data_netWorthTrend = (\[.*?\]);`
	trendMatch := s.extractPattern(jsContent, pattern)
	if trendMatch == "" {
		return nil, fmt.Errorf("未找到走势数据")
	}

	// 替换单引号为双引号
	jsonStr := regexp.MustCompile(`'`).ReplaceAllString(trendMatch, `"`)

	// 解析为 map 数组
	var rawData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, err
	}

	// 转换为 TrendPoint 数组
	var result []model.TrendPoint
	for _, item := range rawData {
		timestamp, ok1 := item["x"].(float64)
		value, ok2 := item["y"].(float64)
		
		if !ok1 || !ok2 {
			continue
		}

		// 转换时间戳为日期字符串
		date := time.Unix(int64(timestamp)/1000, 0).Format("2006-01-02")
		
		result = append(result, model.TrendPoint{
			Date:  date,
			Value: value,
		})
	}

	return result, nil
}

// filterByPeriod 根据周期过滤数据
func (s *FundService) filterByPeriod(data []model.TrendPoint, period string) []model.TrendPoint {
	if len(data) == 0 {
		return data
	}

	now := time.Now()
	var startTime time.Time

	switch period {
	case "week":
		// 最近一周
		startTime = now.AddDate(0, 0, -7)
	case "month":
		// 最近一个月
		startTime = now.AddDate(0, -1, 0)
	case "quarter":
		// 最近一个季度(3个月)
		startTime = now.AddDate(0, -3, 0)
	case "half_year":
		// 最近半年
		startTime = now.AddDate(0, -6, 0)
	case "year":
		// 最近一年
		startTime = now.AddDate(-1, 0, 0)
	case "three_years":
		// 最近三年
		startTime = now.AddDate(-3, 0, 0)
	case "all":
		// 全部数据
		return data
	default:
		// 默认返回最近一个月
		startTime = now.AddDate(0, -1, 0)
	}

	// 过滤数据
	var result []model.TrendPoint
	for _, point := range data {
		pointTime, err := time.Parse("2006-01-02", point.Date)
		if err != nil {
			continue
		}
		if pointTime.After(startTime) || pointTime.Equal(startTime) {
			result = append(result, point)
		}
	}

	return result
}
