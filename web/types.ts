export interface FundBasic {
  code: string;
  name: string;
  type: string;
}

export interface FundData extends FundBasic {
  currentValuation: number; // Current estimated Net Worth
  previousClose: number;    // Yesterday's Net Worth
  growthRate: number;       // Percentage change
  updateTime: string;
  tags: string[];
}

export interface ChartDataPoint {
  time: string;
  value: number;
  rate?: number;    // 后端返回的涨跌幅百分比
  average?: number; // Optional now as it's only for 1D
}

export type TimeRange = '1D' | '1W' | '1M' | '3M';



export enum ViewState {
  DASHBOARD = 'DASHBOARD',
  SEARCH = 'SEARCH',
}