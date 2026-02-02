import { FundData, ChartDataPoint, FundBasic, TimeRange } from '../types';
import { MARKET_FUNDS, STORAGE_KEY, DEFAULT_MY_FUNDS } from '../constants';

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
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 200));

  const basic = MARKET_FUNDS.find(f => f.code === code) || { code, name: '未知基金', type: '未知' };
  
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

export const searchFunds = (query: string): FundBasic[] => {
  if (!query) return [];
  return MARKET_FUNDS.filter(
    f => f.code.includes(query) || f.name.includes(query)
  );
};