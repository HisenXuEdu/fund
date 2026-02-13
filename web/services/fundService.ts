import { FundData, ChartDataPoint, FundBasic, TimeRange } from '../types';
import { MARKET_FUNDS, STORAGE_KEY, DEFAULT_MY_FUNDS } from '../constants';

// API配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// Market hours: 9:30-11:30 (120m), 13:00-15:00 (120m) -> Total 240 minutes
const TOTAL_MINUTES = 240;

const getMinutesFromTime = (timeStr: string): number => {
  if (!timeStr) return TOTAL_MINUTES;
  const [h, m] = timeStr.split(':').map(Number);
  const minutesFromMidnight = h * 60 + m;
  
  const startMorning = 9 * 60 + 30; // 570
  const endMorning = 11 * 60 + 30;  // 690
  const startAfternoon = 13 * 60;   // 780
  const endAfternoon = 15 * 60;     // 900

  if (minutesFromMidnight < startMorning) return 0;
  if (minutesFromMidnight > endAfternoon) return TOTAL_MINUTES;

  if (minutesFromMidnight <= endMorning) {
    return minutesFromMidnight - startMorning;
  }
  
  // Lunch break returns the end of morning
  if (minutesFromMidnight < startAfternoon) {
    return 120;
  }

  return 120 + (minutesFromMidnight - startAfternoon);
};

// Deterministic pseudo-random trend generator based on fund code
const getDayTrend = (code: string, previousClose: number): number[] => {
  const seed = parseInt(code.replace(/\D/g, '')) || 12345;
  const values: number[] = [previousClose];
  
  // Simple Linear Congruential Generator
  let state = seed;
  const rand = () => {
    state = (state * 9301 + 49297) % 233280;
    return state / 233280;
  };

  let current = previousClose;
  for (let i = 0; i < TOTAL_MINUTES; i++) {
    // Volatility approx 0.3% per minute range
    const change = (rand() - 0.5) * 0.003; 
    current = current * (1 + change);
    values.push(current);
  }
  return values;
};

// Helper to get formatted time string from index
const getDisplayTime = (index: number): string => {
  let totalMins;
  if (index <= 120) {
    totalMins = 9 * 60 + 30 + index;
  } else {
    totalMins = 13 * 60 + (index - 120);
  }
  const h = Math.floor(totalMins / 60);
  const m = totalMins % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
};

export const fetchFundDetails = async (code: string, timeStr: string): Promise<FundData> => {
  try {
    // 1. 先调用后端获取基金详情（包含名称、类型等基本信息）
    const detailResponse = await fetch(`${API_BASE_URL}/fund/detail?code=${code}`);
    
    let fundBasicInfo: FundBasic = {
      code: code,
      name: '未知基金',
      type: '未知'
    };
    
    let previousClose = 1.0;
    let currentValuation = 1.0;
    let growthRate = 0;
    
    if (detailResponse.ok) {
      const detailData = await detailResponse.json();
      
      // 从详情API获取基本信息
      fundBasicInfo = {
        code: detailData.code || code,
        name: detailData.name || '未知基金',
        type: '混合型' // 详情API暂时没有返回类型，先用默认值
      };
      
      // 解析净值和涨跌幅
      previousClose = parseFloat(detailData.currentPrice) || 1.0;
      currentValuation = parseFloat(detailData.estimatePrice) || previousClose;
      growthRate = parseFloat(detailData.estimateRate) || 0;
    }
    
    // 2. 尝试获取日内数据（用于更精确的实时数据）
    try {
      const intradayResponse = await fetch(`${API_BASE_URL}/fund/intraday?code=${code}`);
      
      if (intradayResponse.ok) {
        const intradayData = await intradayResponse.json();
        
        // 如果有名称，更新基本信息
        if (intradayData.name) {
          fundBasicInfo.name = intradayData.name;
        }
        
        // 根据时间过滤数据点
        const targetMinutes = getMinutesFromTime(timeStr);
        const filteredData = intradayData.data?.filter((_: any, index: number) => index <= targetMinutes) || [];
        const latestPoint = filteredData[filteredData.length - 1] || intradayData.data?.[0];
        
        if (latestPoint) {
          currentValuation = latestPoint.value;
          previousClose = intradayData.previousClose || latestPoint.value;
          // 使用后端返回的 rate 字段
          growthRate = latestPoint.rate !== undefined ? latestPoint.rate : ((currentValuation - previousClose) / previousClose) * 100;
        }
      }
    } catch (intradayError) {
      console.log('日内数据获取失败，使用详情数据:', intradayError);
    }
    
    const tags = [fundBasicInfo.type];
    if (growthRate > 1.5) tags.push('大涨');
    else if (growthRate < -1.5) tags.push('大跌');
    else if (Math.abs(growthRate) < 0.2) tags.push('震荡');
    
    return {
      ...fundBasicInfo,
      previousClose,
      currentValuation,
      growthRate,
      updateTime: timeStr,
      tags
    };
  } catch (error) {
    console.error('获取基金数据失败，使用模拟数据:', error);
    // 回退到模拟数据
    return fetchFundDetailsFallback(code, timeStr);
  }
};

// 回退方案：使用模拟数据
const fetchFundDetailsFallback = async (code: string, timeStr: string): Promise<FundData> => {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 200));

  // 尝试从搜索接口获取基金基本信息
  let basic: FundBasic = { code, name: `基金${code}`, type: '混合型' };
  
  try {
    const searchResponse = await fetch(`${API_BASE_URL}/fund/list?keyword=${code}&pageSize=1`);
    if (searchResponse.ok) {
      const searchData = await searchResponse.json();
      if (searchData.data && searchData.data.length > 0) {
        const fund = searchData.data[0];
        basic = {
          code: fund.code || code,
          name: fund.name || `基金${code}`,
          type: fund.type || '混合型'
        };
      }
    }
  } catch (err) {
    console.log('无法从搜索接口获取基金信息，使用默认值');
  }
  
  // 如果还是找不到，尝试从本地常量查找
  const localFund = MARKET_FUNDS.find(f => f.code === code);
  if (localFund) {
    basic = localFund;
  }
  
  // Deterministic previous close
  const seed = parseInt(code.replace(/\D/g, '')) || 1;
  const previousClose = 1.0 + (seed % 50) / 10;
  
  const fullTrend = getDayTrend(code, previousClose);
  const targetIndex = getMinutesFromTime(timeStr);
  const currentIndex = Math.min(targetIndex, fullTrend.length - 1);
  
  const currentValuation = fullTrend[currentIndex];
  const growthRate = ((currentValuation - previousClose) / previousClose) * 100;
  
  const tags = [basic.type];
  if (growthRate > 1.5) tags.push('大涨');
  else if (growthRate < -1.5) tags.push('大跌');
  else if (Math.abs(growthRate) < 0.2) tags.push('震荡');
  
  return {
    ...basic,
    previousClose,
    currentValuation,
    growthRate,
    updateTime: timeStr,
    tags
  };
};

// Generate historical daily candles
const getHistoricalTrend = (code: string, days: number, currentVal: number): ChartDataPoint[] => {
  const data: ChartDataPoint[] = [];
  const seed = parseInt(code.replace(/\D/g, '')) || 1;
  
  let state = seed * 7; // Different seed from intraday
  const rand = () => {
    state = (state * 9301 + 49297) % 233280;
    return state / 233280;
  };

  // Generate backwards from current value
  let val = currentVal;
  
  // Add today first (will be reversed later)
  const today = new Date();
  
  for (let i = 0; i < days; i++) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);
    
    // Format MM-DD
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    const timeStr = `${month}-${day}`;

    data.push({
      time: timeStr,
      value: val,
    });

    // Determine previous day value (reverse volatility)
    // Daily volatility approx 1-2%
    const change = (rand() - 0.5) * 0.04;
    val = val / (1 + change);
  }

  return data.reverse();
};

export const fetchFundChartData = async (code: string, timeRange: TimeRange, timeStr: string): Promise<ChartDataPoint[]> => {
  try {
    if (timeRange === '1D') {
      // 日内数据：调用Go后端API
      const response = await fetch(`${API_BASE_URL}/fund/intraday?code=${code}`);
      
      if (!response.ok) {
        throw new Error(`API请求失败: ${response.status}`);
      }
      
      const data = await response.json();
      
      // 检查是否为无数据标记
      if (data.data && data.data.length === 1 && data.data[0].time === 'unknown') {
        // 返回空数组，前端会显示无数据提示
        return [];
      }
      
      if (!data.data || data.data.length === 0) {
        // 回退到模拟数据
        return fetchFundChartDataFallback(code, timeRange, timeStr);
      }
      
      const targetMinutes = getMinutesFromTime(timeStr);
      const chartData: ChartDataPoint[] = data.data
        .filter((_: any, index: number) => index <= targetMinutes)
        .map((point: any) => ({
          time: point.time,
          value: point.value,
          rate: point.rate,  // 传递rate字段
          average: data.previousClose
        }));
      
      if (chartData.length === 0) {
        chartData.push({ 
          time: '09:30', 
          value: data.previousClose || 1.0, 
          average: data.previousClose || 1.0 
        });
      }
      
      return chartData;
    } else {
      // 历史数据：调用Go后端趋势API
      const periodMap: Record<TimeRange, string> = {
        '1D': 'week',
        '1W': 'week',
        '1M': 'month',
        '3M': 'quarter'
      };
      
      const period = periodMap[timeRange] || 'month';
      const response = await fetch(`${API_BASE_URL}/fund/trend?code=${code}&period=${period}`);
      
      if (!response.ok) {
        throw new Error(`API请求失败: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (!data.trendData || data.trendData.length === 0) {
        // 回退到模拟数据
        return fetchFundChartDataFallback(code, timeRange, timeStr);
      }
      
      return data.trendData.map((point: any) => ({
        time: point.time || point.date,
        value: point.value || point.netWorth
      }));
    }
  } catch (error) {
    console.error('获取图表数据失败，使用模拟数据:', error);
    // 回退到模拟数据
    return fetchFundChartDataFallback(code, timeRange, timeStr);
  }
};

// 回退方案：使用模拟数据
const fetchFundChartDataFallback = async (code: string, timeRange: TimeRange, timeStr: string): Promise<ChartDataPoint[]> => {
  await new Promise(resolve => setTimeout(resolve, 300));
  
  const seed = parseInt(code.replace(/\D/g, '')) || 1;
  const previousClose = 1.0 + (seed % 50) / 10;
  
  if (timeRange === '1D') {
    const fullTrend = getDayTrend(code, previousClose);
    const targetIndex = getMinutesFromTime(timeStr);
    const currentIndex = Math.min(targetIndex, fullTrend.length - 1);

    const data: ChartDataPoint[] = [];
    // Start from 1 to skip the initial previousClose point at t=0 effectively
    for (let i = 1; i <= currentIndex; i++) {
      data.push({
        time: getDisplayTime(i),
        value: fullTrend[i],
        average: previousClose
      });
    }

    if (data.length === 0) {
        data.push({ time: '09:30', value: previousClose, average: previousClose });
    }
    return data;
  } else {
    // Calculate current value to be consistent with header
    const fullTrend = getDayTrend(code, previousClose);
    const targetIndex = getMinutesFromTime(timeStr);
    const currentIndex = Math.min(targetIndex, fullTrend.length - 1);
    const currentValuation = fullTrend[currentIndex];

    let days = 7;
    if (timeRange === '1M') days = 30;
    if (timeRange === '3M') days = 90;

    return getHistoricalTrend(code, days, currentValuation);
  }
};

export const getSavedFundCodes = (): string[] => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? JSON.parse(stored) : DEFAULT_MY_FUNDS;
  } catch (e) {
    return DEFAULT_MY_FUNDS;
  }
};

export const saveFundCodes = (codes: string[]) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(codes));
};

export const searchFunds = async (query: string): Promise<FundBasic[]> => {
  if (!query || query.trim().length === 0) return [];
  
  try {
    // 调用Go后端API搜索基金
    const response = await fetch(`${API_BASE_URL}/fund/list?keyword=${encodeURIComponent(query)}&pageSize=50`);
    
    if (!response.ok) {
      throw new Error(`API请求失败: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (!data.data || data.data.length === 0) {
      return [];
    }
    
    // 转换为 FundBasic 格式
    return data.data.map((fund: any) => ({
      code: fund.code,
      name: fund.name,
      type: fund.type
    }));
  } catch (error) {
    console.error('搜索基金失败，使用本地数据:', error);
    // 回退到本地数据
    return MARKET_FUNDS.filter(
      f => f.code.includes(query) || f.name.includes(query)
    );
  }
};