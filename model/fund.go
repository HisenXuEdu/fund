package model

// FundDetail 基金详细信息
type FundDetail struct {
	Code          string `json:"code"`          // 基金代码
	Name          string `json:"name"`          // 基金名称
	CurrentPrice  string `json:"currentPrice"`  // 当前净值
	EstimatePrice string `json:"estimatePrice"` // 估算净值
	EstimateRate  string `json:"estimateRate"`  // 估算增长率
	UpdateTime    string `json:"updateTime"`    // 更新时间
	DayGrowth     string `json:"dayGrowth"`     // 日增长率
	WeekGrowth    string `json:"weekGrowth"`    // 周增长率
	MonthGrowth   string `json:"monthGrowth"`   // 月增长率
	ThreeMonth    string `json:"threeMonth"`    // 近3月增长率
	SixMonth      string `json:"sixMonth"`      // 近6月增长率
	YearGrowth    string `json:"yearGrowth"`    // 近1年增长率
	TotalGrowth   string `json:"totalGrowth"`   // 成立以来增长率
}

// RealtimeData 实时估值数据
type RealtimeData struct {
	FundCode string `json:"fundcode"` // 基金代码
	Name     string `json:"name"`     // 基金名称
	JzRiqi   string `json:"jzrq"`     // 净值日期
	DwJz     string `json:"dwjz"`     // 当日净值
	GsZzl    string `json:"gszzl"`    // 估算增长率
	Gsz      string `json:"gsz"`      // 估算净值
	Gztime   string `json:"gztime"`   // 估值时间
}

// TrendPoint 走势数据点
type TrendPoint struct {
	Date  string  `json:"date"`  // 日期
	Value float64 `json:"value"` // 净值
}

// FundTrend 基金走势数据
type FundTrend struct {
	Code   string        `json:"code"`   // 基金代码
	Name   string        `json:"name"`   // 基金名称
	Period string        `json:"period"` // 周期类型: week/month/quarter
	Data   []TrendPoint  `json:"data"`   // 走势数据点
}

// FundBasicInfo 基金基本信息
type FundBasicInfo struct {
	Code string `json:"code"` // 基金代码
	Name string `json:"name"` // 基金名称
	Type string `json:"type"` // 基金类型
}

// IntradayPoint 日内数据点
type IntradayPoint struct {
	Time  string  `json:"time"`  // 时间 HH:MM
	Value float64 `json:"value"` // 估算净值
	Rate  float64 `json:"rate"`  // 估算涨跌幅
}

// FundIntradayData 基金日内实时数据
type FundIntradayData struct {
	Code string          `json:"code"` // 基金代码
	Name string          `json:"name"` // 基金名称
	Date string          `json:"date"` // 日期
	Data []IntradayPoint `json:"data"` // 日内数据点
}
