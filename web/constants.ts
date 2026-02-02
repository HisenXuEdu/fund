import { FundBasic } from './types';

// A mock database of available funds in the market
export const MARKET_FUNDS: FundBasic[] = [
  { code: '000001', name: '华夏成长混合', type: '混合型' },
  { code: '110011', name: '易方达中小盘', type: '混合型' },
  { code: '005827', name: '易方达蓝筹精选', type: '混合型' },
  { code: '001618', name: '天弘沪深300ETF', type: '指数型' },
  { code: '003095', name: '中欧医疗健康', type: '混合型' },
  { code: '161725', name: '招商中证白酒', type: '指数型' },
  { code: '001594', name: '天弘中证500', type: '指数型' },
  { code: '000198', name: '天弘余额宝', type: '货币型' },
  { code: '005918', name: '广发双擎升级', type: '混合型' },
  { code: '008086', name: '华夏中证5G', type: '指数型' },
  { code: '012414', name: '招商新能源', type: '混合型' },
  { code: '004854', name: '广发中证传媒', type: '指数型' },
];

export const DEFAULT_MY_FUNDS = ['005827', '161725', '001618'];

export const STORAGE_KEY = 'smartfund_favorites';