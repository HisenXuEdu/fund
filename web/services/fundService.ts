import { FundData, ChartDataPoint, FundBasic, TimeRange } from '../types';
import { STORAGE_KEY, DEFAULT_MY_FUNDS } from '../constants';

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
    console.error('获取基金数据失败:', error);
    // 不使用模拟数据,返回错误状态
    throw error;
  }
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
        // 返回空数组,不使用模拟数据
        return [];
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
      
      // 后端返回的是 FundTrend 结构: { code, name, period, data: [] }
      if (!data.data || data.data.length === 0) {
        // 返回空数组,不使用模拟数据
        return [];
      }
      
      return data.data.map((point: any) => ({
        time: point.date,  // 后端返回的是 date 字段
        value: point.value
      }));
    }
  } catch (error) {
    console.error('获取图表数据失败:', error);
    // 不使用模拟数据,返回空数组
    return [];
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
    console.error('搜索基金失败:', error);
    // 不使用本地数据,返回空数组
    return [];
  }
};