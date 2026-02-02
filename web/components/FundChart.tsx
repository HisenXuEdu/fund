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
  if (data.length === 0) return null;

  // Calculate min/max domain to make the chart look dynamic
  const values = data.map(d => d.value);
  
  // Use previousClose as anchor only for 1D chart
  const allValues = timeRange === '1D' ? [...values, previousClose] : values;
  
  const minVal = Math.min(...allValues);
  const maxVal = Math.max(...allValues);
  const padding = (maxVal - minVal) * 0.1; // 10% padding
  
  const min = minVal - padding;
  const max = maxVal + padding;

  const isIntraday = timeRange === '1D';

  return (
    <div className="w-full h-64 bg-white rounded-xl shadow-sm border border-gray-100 p-2">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart
          data={data}
          margin={{
            top: 10,
            right: 0,
            left: 0,
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
            interval={isIntraday ? 59 : 'preserveStartEnd'}
            axisLine={false}
            tickLine={false}
          />
          <YAxis 
            domain={[min, max]} 
            hide={true} 
          />
          <Tooltip 
            contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
            labelStyle={{ color: '#64748b', fontSize: '12px' }}
            formatter={(value: number) => [value.toFixed(4), '净值']}
          />
          {isIntraday && (
            <ReferenceLine y={previousClose} stroke="#94a3b8" strokeDasharray="3 3" />
          )}
          <Area 
            type="monotone" 
            dataKey="value" 
            stroke={color} 
            fillOpacity={1} 
            fill="url(#colorValue)" 
            strokeWidth={2}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};

export default FundChart;