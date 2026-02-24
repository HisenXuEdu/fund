package service

import (
	"context"
	"encoding/json"
	"fmt"
	"fund/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// WatchConfig 监控配置
type WatchConfig struct {
	WatchList     []string `json:"watch_list"`     // 监控的基金代码列表
	FetchInterval int      `json:"fetch_interval"` // 采集周期（秒）
}

// IntradayService 日内实时数据服务
type IntradayService struct {
	httpClient   *http.Client
	fundList     []model.FundBasicInfo              // 基金列表
	intradayData map[string]*model.FundIntradayData // 日内数据存储 key: fundCode
	dataMutex    sync.RWMutex                       // 数据锁
	stopChan     chan struct{}                      // 停止信号
	isRunning    bool                               // 是否正在运行
	dataDir      string                             // 数据存储目录
	watchConfig  *WatchConfig                       // 监控配置
	configFile   string                             // 配置文件路径
	fundService  *FundService                       // 基金服务（用于批量获取）
	lastDataDate string                             // 最后一次采集数据的日期 (YYYY-MM-DD)
	csrfToken    string                             // fund123.cn的CSRF令牌
	fundKeyCache map[string]string                  // 基金代码到fundKey的缓存
}

// NewIntradayService 创建日内服务实例
func NewIntradayService() *IntradayService {
	return &IntradayService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		intradayData: make(map[string]*model.FundIntradayData),
		stopChan:     make(chan struct{}),
		dataDir:      "./data",             // 数据存储目录
		configFile:   "./watch_funds.json", // 配置文件路径
		fundService:  NewFundService(),     // 初始化基金服务
		fundKeyCache: make(map[string]string),
	}
}

// LoadWatchConfig 加载监控配置
func (s *IntradayService) LoadWatchConfig() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(s.configFile); os.IsNotExist(err) {
		log.Printf("⚠️  配置文件不存在: %s, 将采集全量基金", s.configFile)
		return nil
	}

	// 读取配置文件
	file, err := os.Open(s.configFile)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer file.Close()

	// 解析JSON
	var config WatchConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if len(config.WatchList) == 0 {
		log.Printf("⚠️  配置文件中监控列表为空，将采集全量基金")
		return nil
	}
	if config.FetchInterval <= 0 {
		config.FetchInterval = 30 // 默认30秒
	}

	s.watchConfig = &config
	log.Printf("✅ 加载监控配置: %d 只基金, 采集周期 %d 秒", len(config.WatchList), config.FetchInterval)
	log.Printf("📋 监控基金: %v", config.WatchList)

	return nil
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

// LoadFromDisk 从硬盘加载实时数据到内存
func (s *IntradayService) LoadFromDisk() error {
	// 确保数据目录存在
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	dataFile := filepath.Join(s.dataDir, "intraday_data.json")

	// 检查文件是否存在
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		log.Println("💾 未找到持久化数据文件，将使用空数据")
		return nil
	}

	// 读取文件
	file, err := os.Open(dataFile)
	if err != nil {
		return fmt.Errorf("打开数据文件失败: %v", err)
	}
	defer file.Close()

	// 解析JSON
	var diskData map[string]*model.FundIntradayData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&diskData); err != nil {
		return fmt.Errorf("解析数据文件失败: %v", err)
	}

	// 加载到内存
	s.dataMutex.Lock()
	s.intradayData = diskData
	s.dataMutex.Unlock()

	log.Printf("✅ 从硬盘加载了 %d 只基金的实时数据", len(diskData))
	return nil
}

// SaveToDisk 将内存中的实时数据保存到硬盘
func (s *IntradayService) SaveToDisk() error {
	// 确保数据目录存在
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	dataFile := filepath.Join(s.dataDir, "intraday_data.json")

	// 读取内存数据
	s.dataMutex.RLock()
	dataCopy := make(map[string]*model.FundIntradayData, len(s.intradayData))
	for k, v := range s.intradayData {
		dataCopy[k] = v
	}
	s.dataMutex.RUnlock()

	// 写入临时文件
	tmpFile := dataFile + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}

	// 编码JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dataCopy); err != nil {
		file.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("编码数据失败: %v", err)
	}
	file.Close()

	// 原子性替换文件
	if err := os.Rename(tmpFile, dataFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("替换数据文件失败: %v", err)
	}

	log.Printf("💾 已保存 %d 只基金的实时数据到硬盘", len(dataCopy))
	return nil
}

// GetFundList 获取基金列表
func (s *IntradayService) GetFundList() []interface{} {
	result := make([]interface{}, len(s.fundList))
	for i, fund := range s.fundList {
		result[i] = map[string]interface{}{
			"code": fund.Code,
			"name": fund.Name,
			"type": fund.Type,
		}
	}
	return result
}

// Fund123IntradayPoint fund123接口返回的分时数据点
type Fund123IntradayPoint struct {
	Time     string  `json:"x"` // 时间 HH:MM
	Growth   float64 `json:"y"` // 涨跌幅
	NetValue float64 `json:"equityReturn"` // 净值
}

// getCSRFToken 获取fund123.cn的CSRF令牌
func (s *IntradayService) getCSRFToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.fund123.cn/fund", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 提取CSRF令牌
	re := regexp.MustCompile(`"csrf":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return fmt.Errorf("未找到CSRF令牌")
	}

	s.csrfToken = matches[1]
	log.Printf("✅ 获取CSRF令牌成功: %s", s.csrfToken[:8]+"...")
	return nil
}

// getFundKey 根据基金代码获取fundKey
func (s *IntradayService) getFundKey(fundCode string) (string, string, error) {
	// 检查缓存
	s.dataMutex.RLock()
	if fundKey, exists := s.fundKeyCache[fundCode]; exists {
		// 获取基金名称
		fundName := fundCode
		if data, ok := s.intradayData[fundCode]; ok {
			fundName = data.Name
		}
		s.dataMutex.RUnlock()
		return fundKey, fundName, nil
	}
	s.dataMutex.RUnlock()

	// 如果没有CSRF令牌，先获取
	if s.csrfToken == "" {
		if err := s.getCSRFToken(); err != nil {
			return "", "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://www.fund123.cn/api/fund/search?keyword=%s&_csrf=%s", fundCode, s.csrfToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.fund123.cn/fund")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var result struct {
		Data []struct {
			FundKey  string `json:"fundKey"`
			FundName string `json:"fundName"`
			FundCode string `json:"fundCode"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("解析JSON失败: %v", err)
	}

	if len(result.Data) == 0 {
		return "", "", fmt.Errorf("未找到基金: %s", fundCode)
	}

	fundKey := result.Data[0].FundKey
	fundName := result.Data[0].FundName

	// 缓存fundKey
	s.dataMutex.Lock()
	s.fundKeyCache[fundCode] = fundKey
	s.dataMutex.Unlock()

	return fundKey, fundName, nil
}

// fetchFund123IntradayData 从fund123.cn获取分时数据
func (s *IntradayService) fetchFund123IntradayData(fundCode string) ([]Fund123IntradayPoint, string, error) {
	// 获取fundKey
	fundKey, fundName, err := s.getFundKey(fundCode)
	if err != nil {
		return nil, "", err
	}

	// 如果没有CSRF令牌，先获取
	if s.csrfToken == "" {
		if err := s.getCSRFToken(); err != nil {
			return nil, "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://www.fund123.cn/api/fund/trend-data?fundKey=%s&_csrf=%s", fundKey, s.csrfToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.fund123.cn/fund")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON
	var result struct {
		Data []Fund123IntradayPoint `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", fmt.Errorf("解析JSON失败: %v", err)
	}

	return result.Data, fundName, nil
}

// fetchRealtimeEstimate 获取单个基金的实时估值（使用fund123接口）
func (s *IntradayService) fetchRealtimeEstimate(fundCode string) (*model.RealtimeData, error) {
	// 使用fund123接口获取分时数据
	intradayPoints, fundName, err := s.fetchFund123IntradayData(fundCode)
	if err != nil {
		return nil, err
	}

	if len(intradayPoints) == 0 {
		return nil, fmt.Errorf("无分时数据")
	}

	// 获取最新的数据点
	latest := intradayPoints[len(intradayPoints)-1]

	// 转换为RealtimeData格式（保持兼容性）
	return &model.RealtimeData{
		FundCode: fundCode,
		Name:     fundName,
		Gsz:      fmt.Sprintf("%.4f", latest.NetValue),
		GsZzl:    fmt.Sprintf("%.2f", latest.Growth),
		Gztime:   latest.Time,
	}, nil
}

// fetchAllFundsRealtimeBatch 使用批量接口获取全量基金实时数据
func (s *IntradayService) fetchAllFundsRealtimeBatch() {
	now := time.Now()

	// 判断是否在交易时间
	if !s.isTradingTime(now) {
		log.Printf("⏸️  非交易时间 [%s], 跳过本次采集", now.Format("15:04"))
		return
	}

	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	// 检查日期变化,如果是新的一天则清理旧数据
	s.checkAndClearOldData(today)

	log.Printf("📊 开始使用批量接口获取全量基金实时数据 [%s]", currentTime)

	startTime := time.Now()

	// 先获取第一页以获取总数
	firstPageData, err := s.fundService.FetchBatchFundsForRealtime(1, 200)
	if err != nil {
		log.Printf("❌ 获取第一页失败: %v", err)
		return
	}

	// 假设总基金数（可以从之前加载的基金列表获取）
	totalFunds := len(s.fundList)
	pageSize := 200
	totalPages := (totalFunds + pageSize - 1) / pageSize

	log.Printf("📈 预计总页数: %d, 基金总数: %d", totalPages, totalFunds)

	var successCount, failCount int

	// 处理第一页数据
	s.processBatchFundsData(firstPageData, today, currentTime)
	successCount += len(firstPageData)

	log.Printf("✅ 第 1/%d 页完成, 获取 %d 只基金", totalPages, len(firstPageData))

	// 获取剩余页面
	for page := 2; page <= totalPages; page++ {
		// 请求间隔，避免触发反爬虫
		time.Sleep(200 * time.Millisecond)

		pageData, err := s.fundService.FetchBatchFundsForRealtime(page, pageSize)
		if err != nil {
			log.Printf("⚠️  获取第 %d 页失败: %v", page, err)
			failCount++
			continue
		}

		s.processBatchFundsData(pageData, today, currentTime)
		successCount += len(pageData)

		log.Printf("✅ 第 %d/%d 页完成, 获取 %d 只基金, 累计: %d",
			page, totalPages, len(pageData), successCount)
	}

	elapsed := time.Since(startTime)
	log.Printf("✅ 批量采集完成: 成功 %d 只基金, 失败 %d 页, 耗时 %v",
		successCount, failCount, elapsed)
}

// processBatchFundsData 处理批量基金数据
func (s *IntradayService) processBatchFundsData(fundsData map[string]map[string]interface{}, today, currentTime string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	for fundCode, data := range fundsData {
		// 获取基金名称
		fundName := ""
		if name, ok := data["name"].(string); ok {
			fundName = name
		}

		// 解析净值和涨跌幅
		var value, rate float64
		if netValue, ok := data["netValue"].(string); ok && netValue != "" && netValue != "---" {
			value, _ = strconv.ParseFloat(netValue, 64)
		}
		if dayGrowth, ok := data["dayGrowth"].(string); ok && dayGrowth != "" && dayGrowth != "---" {
			rate, _ = strconv.ParseFloat(dayGrowth, 64)
		}

		// 如果没有有效数据，跳过
		if value == 0 {
			continue
		}

		// 存储数据
		if _, exists := s.intradayData[fundCode]; !exists {
			// 首次创建
			s.intradayData[fundCode] = &model.FundIntradayData{
				Code: fundCode,
				Name: fundName,
				Date: today,
				Data: []model.IntradayPoint{},
			}
		}

		// 更新或添加最新数据点
		fundData := s.intradayData[fundCode]

		// 检查日期是否需要清空（新的一天）
		if fundData.Date != today {
			fundData.Date = today
			fundData.Data = []model.IntradayPoint{}
		}

		// 添加或更新当前时间点的数据
		found := false
		for i := range fundData.Data {
			if fundData.Data[i].Time == currentTime {
				fundData.Data[i].Value = value
				fundData.Data[i].Rate = rate
				found = true
				break
			}
		}
		if !found {
			fundData.Data = append(fundData.Data, model.IntradayPoint{
				Time:  currentTime,
				Value: value,
				Rate:  rate,
			})
		}
	}
}

// fetchWatchListRealtime 获取监控列表中基金的实时数据（使用fund123完整分时数据）
func (s *IntradayService) fetchWatchListRealtime() {
	if s.watchConfig == nil || len(s.watchConfig.WatchList) == 0 {
		log.Printf("⚠️  监控列表为空，跳过采集")
		return
	}

	now := time.Now()

	// 判断是否在交易时间
	if !s.isTradingTime(now) {
		log.Printf("⏸️  非交易时间 [%s], 跳过本次采集", now.Format("15:04"))
		return
	}

	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	// 检查日期变化,如果是新的一天则清理旧数据
	s.checkAndClearOldData(today)

	watchList := s.watchConfig.WatchList
	totalFunds := len(watchList)
	fetchInterval := s.watchConfig.FetchInterval

	log.Printf("📊 开始获取监控列表基金分时数据 [%s], 基金数: %d, 周期: %d秒",
		currentTime, totalFunds, fetchInterval)

	// 计算每只基金的请求间隔（均匀分布）
	intervalPerFund := time.Duration(fetchInterval*1000/totalFunds) * time.Millisecond
	log.Printf("⏱️  每只基金间隔: %v", intervalPerFund)

	startTime := time.Now()
	var successCount, failCount int

	for i, fundCode := range watchList {
		// 获取完整分时数据
		intradayPoints, fundName, err := s.fetchFund123IntradayData(fundCode)
		if err != nil {
			log.Printf("❌ [%d/%d] %s 获取失败: %v", i+1, totalFunds, fundCode, err)
			failCount++
		} else if len(intradayPoints) == 0 {
			log.Printf("⚠️  [%d/%d] %s (%s) 无分时数据", i+1, totalFunds, fundName, fundCode)
			failCount++
		} else {
			// 存储完整分时数据
			s.dataMutex.Lock()

			// 创建或更新基金数据
			s.intradayData[fundCode] = &model.FundIntradayData{
				Code: fundCode,
				Name: fundName,
				Date: today,
				Data: make([]model.IntradayPoint, 0, len(intradayPoints)),
			}

			// 转换数据格式
			for _, point := range intradayPoints {
				s.intradayData[fundCode].Data = append(s.intradayData[fundCode].Data, model.IntradayPoint{
					Time:  point.Time,
					Value: point.NetValue,
					Rate:  point.Growth,
				})
			}

			s.dataMutex.Unlock()

			latest := intradayPoints[len(intradayPoints)-1]
			log.Printf("✅ [%d/%d] %s (%s) 最新: %.4f (%.2f%%), 共 %d 个数据点",
				i+1, totalFunds, fundName, fundCode, latest.NetValue, latest.Growth, len(intradayPoints))
			successCount++
		}

		// 均匀分布请求（最后一只基金不需要等待）
		if i < totalFunds-1 {
			time.Sleep(intervalPerFund)
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("✅ 监控列表采集完成: 成功 %d, 失败 %d, 耗时 %v",
		successCount, failCount, elapsed)
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

	// 加载监控配置
	log.Println("📝 正在加载监控配置...")
	if err := s.LoadWatchConfig(); err != nil {
		log.Printf("⚠️  加载配置失败: %v, 将采集全量基金", err)
	}

	// 加载基金列表
	log.Println("🔄 正在加载基金列表...")
	if err := s.LoadAllFunds(); err != nil {
		return err
	}

	// 从硬盘加载历史数据
	log.Println("💾 正在从硬盘加载历史数据...")
	if err := s.LoadFromDisk(); err != nil {
		log.Printf("⚠️  加载历史数据失败: %v", err)
	}

	s.isRunning = true

	// 根据配置决定采集模式
	if s.watchConfig != nil && len(s.watchConfig.WatchList) > 0 {
		log.Printf("🚀 日内实时数据服务已启动 (监控模式: %d 只基金, %d秒周期)",
			len(s.watchConfig.WatchList), s.watchConfig.FetchInterval)

		// 启动定时任务：按配置周期获取监控列表基金实时数据
		go func() {
			ticker := time.NewTicker(time.Duration(s.watchConfig.FetchInterval) * time.Second)
			defer ticker.Stop()

			// 立即执行一次
			s.fetchWatchListRealtime()

			for {
				select {
				case <-ticker.C:
					s.fetchWatchListRealtime()

				case <-s.stopChan:
					log.Println("⏹️  停止实时数据采集服务")
					return
				}
			}
		}()
	} else {
		log.Println("🚀 日内实时数据服务已启动 (全量模式 - 使用批量接口)")

		// 启动定时任务：每1分钟获取全量基金实时数据（使用批量接口）
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()

			// 立即执行一次
			s.fetchAllFundsRealtimeBatch()

			for {
				select {
				case <-ticker.C:
					s.fetchAllFundsRealtimeBatch()

				case <-s.stopChan:
					log.Println("⏹️  停止实时数据采集服务")
					return
				}
			}
		}()
	}

	// 启动定时任务：每5分钟保存数据到硬盘
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.SaveToDisk(); err != nil {
					log.Printf("❌ 保存数据到硬盘失败: %v", err)
				}

			case <-s.stopChan:
				// 服务停止前最后保存一次
				log.Println("💾 服务停止前保存数据...")
				if err := s.SaveToDisk(); err != nil {
					log.Printf("❌ 保存数据失败: %v", err)
				}
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

	// 检查是否有当日数据
	if len(data.Data) == 0 {
		// 返回标记数据表示无当日数据
		return &model.FundIntradayData{
			Code: fundCode,
			Name: data.Name,
			Date: data.Date,
			Data: []model.IntradayPoint{
				{
					Time:  "unknown",
					Value: 0,
					Rate:  0,
				},
			},
		}, nil
	}

	return data, nil
}

// ClearTodayData 清理当天数据(包括内存和磁盘)
func (s *IntradayService) ClearTodayData() {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	count := len(s.intradayData)
	s.intradayData = make(map[string]*model.FundIntradayData)

	// 清理磁盘数据文件
	dataFile := filepath.Join(s.dataDir, "intraday_data.json")
	if err := os.Remove(dataFile); err != nil && !os.IsNotExist(err) {
		log.Printf("⚠️  清理磁盘数据文件失败: %v", err)
	}

	log.Printf("🗑️  已清理当天数据(内存+磁盘),共 %d 只基金", count)
}

// checkAndClearOldData 检查日期变化,如果是新的一天则清理旧数据
func (s *IntradayService) checkAndClearOldData(currentDate string) {
	// 如果没有记录日期,直接更新为当前日期
	if s.lastDataDate == "" {
		s.lastDataDate = currentDate
		return
	}

	// 如果日期发生变化(新的一天),清理旧数据
	if s.lastDataDate != currentDate {
		log.Printf("📅 检测到日期变化: %s -> %s, 清理前一天的数据", s.lastDataDate, currentDate)
		s.ClearTodayData()
		s.lastDataDate = currentDate
	}
}

// GetDataCount 获取已采集的基金数量
func (s *IntradayService) GetDataCount() int {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()
	return len(s.intradayData)
}

// SaveData 保存数据到磁盘(供外部调用)
func (s *IntradayService) SaveData() error {
	return s.SaveToDisk()
}
