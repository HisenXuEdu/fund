package service

import (
	"encoding/json"
	"fmt"
	"fund/model"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// IntradayService 日内实时数据服务
type IntradayService struct {
	httpClient    *http.Client
	fundList      []model.FundBasicInfo       // 基金列表
	intradayData  map[string]*model.FundIntradayData // 日内数据存储 key: fundCode
	dataMutex     sync.RWMutex                // 数据锁
	stopChan      chan struct{}               // 停止信号
	isRunning     bool                        // 是否正在运行
}

// NewIntradayService 创建日内服务实例
func NewIntradayService() *IntradayService {
	return &IntradayService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		intradayData: make(map[string]*model.FundIntradayData),
		stopChan:     make(chan struct{}),
	}
}

// LoadAllFunds 加载所有基金列表
func (s *IntradayService) LoadAllFunds() error {
	url := "http://fund.eastmoney.com/js/fundcode_search.js"
	
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("获取基金列表失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取基金列表失败: %v", err)
	}

	// 解析 JS 格式: var r = [["000001","HXCZHH","华夏成长混合","混合型-偏股","HUAXIACHENGZHANGHUNHE"],...]
	content := string(body)
	pattern := regexp.MustCompile(`var r = (\[\[.*?\]\]);`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return fmt.Errorf("解析基金列表失败")
	}

	// 解析 JSON
	var rawList [][]string
	if err := json.Unmarshal([]byte(matches[1]), &rawList); err != nil {
		return fmt.Errorf("解析基金数据失败: %v", err)
	}

	// 转换为基金信息列表
	s.fundList = make([]model.FundBasicInfo, 0, len(rawList))
	for _, item := range rawList {
		if len(item) >= 4 {
			s.fundList = append(s.fundList, model.FundBasicInfo{
				Code: item[0],
				Name: item[2],
				Type: item[3],
			})
		}
	}

	log.Printf("✅ 成功加载 %d 只基金", len(s.fundList))
	return nil
}

// GetFundList 获取基金列表
func (s *IntradayService) GetFundList() []model.FundBasicInfo {
	return s.fundList
}

// fetchRealtimeEstimate 获取单个基金的实时估值
func (s *IntradayService) fetchRealtimeEstimate(fundCode string) (*model.RealtimeData, error) {
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

	// 解析 jsonpgz({...})
	re := regexp.MustCompile(`jsonpgz\((.*?)\);?$`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("解析失败")
	}

	var realtimeData model.RealtimeData
	if err := json.Unmarshal([]byte(matches[1]), &realtimeData); err != nil {
		return nil, err
	}

	return &realtimeData, nil
}

// collectIntradayData 采集日内数据
func (s *IntradayService) collectIntradayData() {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	// 检查是否在交易时间内 (9:30 - 15:00)
	if !s.isTradingTime(now) {
		return
	}

	log.Printf("📊 开始采集实时数据 [%s]", currentTime)
	
	successCount := 0
	failCount := 0

	// 遍历所有基金采集数据
	for _, fund := range s.fundList {
		// 获取实时估值
		realtime, err := s.fetchRealtimeEstimate(fund.Code)
		if err != nil {
			failCount++
			continue
		}

		// 解析估算净值和涨跌幅
		value, _ := strconv.ParseFloat(realtime.Gsz, 64)
		rate, _ := strconv.ParseFloat(realtime.GsZzl, 64)

		// 存储数据
		s.dataMutex.Lock()
		
		if _, exists := s.intradayData[fund.Code]; !exists {
			// 首次创建
			s.intradayData[fund.Code] = &model.FundIntradayData{
				Code: fund.Code,
				Name: fund.Name,
				Date: today,
				Data: []model.IntradayPoint{},
			}
		}

		// 添加数据点
		s.intradayData[fund.Code].Data = append(s.intradayData[fund.Code].Data, model.IntradayPoint{
			Time:  currentTime,
			Value: value,
			Rate:  rate,
		})
		
		s.dataMutex.Unlock()
		successCount++

		// 避免请求过快,稍微延迟
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("✅ 采集完成: 成功 %d, 失败 %d", successCount, failCount)
}

// isTradingTime 判断是否在交易时间内
func (s *IntradayService) isTradingTime(t time.Time) bool {
	// 只在工作日
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 交易时间: 9:30 - 15:00
	hour := t.Hour()
	minute := t.Minute()
	
	if hour < 9 || hour > 15 {
		return false
	}
	if hour == 9 && minute < 30 {
		return false
	}
	
	return true
}

// Start 启动实时数据采集服务
func (s *IntradayService) Start() error {
	if s.isRunning {
		return fmt.Errorf("服务已在运行中")
	}

	// 加载基金列表
	log.Println("🔄 正在加载基金列表...")
	if err := s.LoadAllFunds(); err != nil {
		return err
	}

	s.isRunning = true
	log.Println("🚀 日内实时数据服务已启动")

	// 启动定时任务
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒采集一次
		defer ticker.Stop()

		// 立即执行一次
		s.collectIntradayData()

		for {
			select {
			case <-ticker.C:
				s.collectIntradayData()
				
				// 检查是否需要清理数据(每天晚上21:00清理)
				now := time.Now()
				if now.Hour() == 21 && now.Minute() == 0 {
					s.ClearTodayData()
				}

			case <-s.stopChan:
				log.Println("⏹️  停止实时数据采集服务")
				return
			}
		}
	}()

	return nil
}

// Stop 停止服务
func (s *IntradayService) Stop() {
	if s.isRunning {
		close(s.stopChan)
		s.isRunning = false
	}
}

// GetIntradayData 获取指定基金的日内数据
func (s *IntradayService) GetIntradayData(fundCode string) (*model.FundIntradayData, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	data, exists := s.intradayData[fundCode]
	if !exists {
		return nil, fmt.Errorf("暂无该基金的日内数据")
	}

	return data, nil
}

// ClearTodayData 清理当天数据
func (s *IntradayService) ClearTodayData() {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	count := len(s.intradayData)
	s.intradayData = make(map[string]*model.FundIntradayData)
	
	log.Printf("🗑️  已清理当天数据,共 %d 只基金", count)
}

// GetDataCount 获取已采集的基金数量
func (s *IntradayService) GetDataCount() int {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()
	return len(s.intradayData)
}
