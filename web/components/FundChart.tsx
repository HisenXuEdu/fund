import React from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';
import { ChartDataPoint, TimeRange } from '../types';

interface FundChartProps {
  data: ChartDataPoint[];
  previousClose: number;
  color: string;
  timeRange: TimeRange;
}

const FundChart: React.FC<FundChartProps> = ({ data, previousClose, color, timeRange }) => {
  // 检查是否为无数据状态
  if (data.length === 0) {
    return (
      <div className="w-full h-64 bg-white rounded-xl shadow-sm border border-gray-100 p-2 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl font-bold text-gray-300 mb-2">Unknown</div>
          <div className="text-sm text-gray-400">当日暂无数据</div>
        </div>
      </div>
    );
  }

  const isIntraday = timeRange === '1D';
  
  // 检查数据是否包含rate字段
  const hasRateData = data.length > 0 && data[0].rate !== undefined;

  // 获取基准值：日内用previousClose，周月季用第一天的净值
  const baseValue = isIntraday 
    ? previousClose 
    : (data.length > 0 ? data[0].value : previousClose);

  // 转换数据: 统一转换为涨跌幅百分比
  const transformedData = hasRateData 
    ? data.map(d => ({ ...d, displayValue: d.rate! }))
    : data.map(d => ({ 
        ...d, 
        displayValue: ((d.value - baseValue) / baseValue * 100)  // 基于起始值计算涨跌幅
      }));

  // Calculate min/max domain
  const displayValues = transformedData.map(d => d.displayValue);
  
  // Y轴范围应该是涨跌幅百分比,基准是0
  const allValues = [...displayValues, 0];  // 基准都是0%
  
  const minVal = Math.min(...allValues);
  const maxVal = Math.max(...allValues);
  const padding = Math.max(Math.abs(maxVal), Math.abs(minVal)) * 0.1;  // 对称padding
  
  const min = minVal - padding;
  const max = maxVal + padding;

  // 格式化Y轴显示涨跌幅 - 统一显示百分比
  const formatYAxis = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  // 生成完整的交易时间轴，但只填充已有数据的点，未到达的时间点设为null
  const generateChartDataWithFullAxis = () => {
    if (!isIntraday) return transformedData;

    const fullData: any[] = [];
    const dataMap = new Map(transformedData.map((d: any) => [d.time, d]));
    
    // 找出最后一个有数据的时间点
    const lastDataTime = transformedData.length > 0 ? transformedData[transformedData.length - 1].time : null;
    let reachedLastData = false;

    // 生成 9:30-11:30 (上午盘)
    for (let h = 9; h <= 11; h++) {
      const startMin = h === 9 ? 30 : 0;
      const endMin = h === 11 ? 30 : 59;
      for (let m = startMin; m <= endMin; m++) {
        const time = `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
        
        if (time === lastDataTime) {
          reachedLastData = true;
        }
        
        const existingData = dataMap.get(time);
        if (existingData) {
          fullData.push(existingData);
        } else if (!reachedLastData) {
          // 数据开始前，用基准值填充
          const baseValue = hasRateData ? 0 : previousClose;
          fullData.push({ time, value: previousClose, displayValue: baseValue, rate: 0 });
        } else {
          // 数据结束后，用null表示无数据（不画线）
          fullData.push({ time, value: null, displayValue: null });
        }
      }
    }

    // 生成 13:00-15:00 (下午盘)
    for (let h = 13; h <= 15; h++) {
      const endMin = h === 15 ? 0 : 59;
      for (let m = 0; m <= endMin; m++) {
        const time = `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
        
        if (time === lastDataTime) {
          reachedLastData = true;
        }
        
        const existingData = dataMap.get(time);
        if (existingData) {
          fullData.push(existingData);
        } else if (!reachedLastData) {
          // 数据开始前，用0%填充(涨跌幅为0)
          fullData.push({ time, value: previousClose, displayValue: 0, rate: 0 });
        } else {
          // 数据结束后，用null表示无数据（不画线）
          fullData.push({ time, value: null, displayValue: null });
        }
      }
    }

    return fullData;
  };

  const chartData = generateChartDataWithFullAxis();

  return (
    <div className="w-full h-64 bg-white rounded-xl shadow-sm border border-gray-100 p-2">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart
          data={chartData}
          margin={{
            top: 10,
            right: 10,
            left: 10,
            bottom: 0,
          }}
        >
          <defs>
            <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.2} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
          <XAxis 
            dataKey="time" 
            tick={{ fontSize: 10, fill: '#94a3b8' }} 
            ticks={isIntraday ? ['09:30', '10:30', '11:30', '13:00', '14:00', '15:00'] : undefined}
            interval={isIntraday ? 'preserveStartEnd' : 'preserveStartEnd'}
            axisLine={false}
            tickLine={false}
          />
          <YAxis 
            domain={[min, max]} 
            tick={{ fontSize: 11, fill: '#64748b' }}
            tickFormatter={formatYAxis}
            axisLine={false}
            tickLine={false}
            width={55}
          />
          <Tooltip 
            contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
            labelStyle={{ color: '#64748b', fontSize: '12px' }}
            formatter={(value: number, name: string) => {
              if (name === 'displayValue') {
                // 统一显示涨跌幅百分比
                return [`${value >= 0 ? '+' : ''}${value.toFixed(2)}%`, '涨跌幅'];
              }
              return [value, name];
            }}
          />
          {isIntraday && (
            <ReferenceLine y={0} stroke="#94a3b8" strokeDasharray="3 3" />
          )}
          <Area 
            type="monotone" 
            dataKey="displayValue" 
            stroke={color} 
            fillOpacity={1} 
            fill="url(#colorValue)" 
            strokeWidth={2}
            connectNulls={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};

export default FundChart;