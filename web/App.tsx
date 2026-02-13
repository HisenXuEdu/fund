import React, { useState, useEffect, useCallback } from 'react';
import { FundData, ChartDataPoint, FundBasic, TimeRange } from './types';
import { fetchFundDetails, fetchFundChartData, getSavedFundCodes, saveFundCodes, searchFunds } from './services/fundService';
import FundChart from './components/FundChart';
import FundCard from './components/FundCard';
import { Search, Plus, PieChart, AlertCircle, Clock, ArrowLeft, RefreshCw } from 'lucide-react';

const App: React.FC = () => {
  const [myFundCodes, setMyFundCodes] = useState<string[]>([]);
  const [fundsData, setFundsData] = useState<FundData[]>([]);
  const [selectedFundCode, setSelectedFundCode] = useState<string | null>(null);
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [loadingChart, setLoadingChart] = useState(false);
  const [loadingList, setLoadingList] = useState(false);
  
  // Time Travel State
  const [simulatedTime, setSimulatedTime] = useState<string>('15:00');
  
  // Time Range State
  const [timeRange, setTimeRange] = useState<TimeRange>('1D');

  // Search State
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<FundBasic[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  // Initialize
  useEffect(() => {
    const codes = getSavedFundCodes();
    setMyFundCodes(codes);
    // Only auto-select on desktop (md breakpoint) to allow list view on mobile initial load
    if (codes.length > 0 && window.innerWidth >= 768) {
      setSelectedFundCode(codes[0]);
    }
  }, []);

  // Fetch list data based on global time
  const loadFundsData = useCallback(async () => {
    if (myFundCodes.length === 0) return;
    setLoadingList(true);
    try {
      const promises = myFundCodes.map(code => fetchFundDetails(code, simulatedTime));
      const results = await Promise.all(promises);
      setFundsData(results);
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingList(false);
    }
  }, [myFundCodes, simulatedTime]);

  useEffect(() => {
    loadFundsData();
  }, [loadFundsData]);

  // Fetch chart data when selection or time changes
  useEffect(() => {
    if (!selectedFundCode) return;
    
    const loadChart = async () => {
      setLoadingChart(true);
      
      try {
        const data = await fetchFundChartData(selectedFundCode, timeRange, simulatedTime);
        setChartData(data);
      } catch (err) {
        console.error(err);
      } finally {
        setLoadingChart(false);
      }
    };

    loadChart();
  }, [selectedFundCode, fundsData, simulatedTime, timeRange]);

  const handleRefresh = () => {
    loadFundsData();
  };

  const handleAddFund = (code: string) => {
    if (!myFundCodes.includes(code)) {
      const newCodes = [...myFundCodes, code];
      setMyFundCodes(newCodes);
      saveFundCodes(newCodes);
      setSelectedFundCode(code);
      setIsSearchOpen(false);
      setSearchQuery('');
    }
  };

  const handleRemoveFund = (e: React.MouseEvent, code: string) => {
    e.stopPropagation();
    const newCodes = myFundCodes.filter(c => c !== code);
    setMyFundCodes(newCodes);
    saveFundCodes(newCodes);
    if (selectedFundCode === code) {
      // If removing the currently selected fund
      if (window.innerWidth >= 768) {
        // On desktop, select next available or null
        setSelectedFundCode(newCodes[0] || null);
      } else {
        // On mobile, go back to list
        setSelectedFundCode(null);
      }
    }
  };

  const handleSearch = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setSearchQuery(query);
    
    if (!query || query.trim().length === 0) {
      setSearchResults([]);
      setIsSearching(false);
      return;
    }
    
    // 防抖：等待用户停止输入
    setIsSearching(true);
    
    // 清除之前的定时器
    if ((window as any).searchTimeout) {
      clearTimeout((window as any).searchTimeout);
    }
    
    // 设置新的定时器
    (window as any).searchTimeout = setTimeout(async () => {
      try {
        const results = await searchFunds(query);
        setSearchResults(results);
      } catch (error) {
        console.error('搜索失败:', error);
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 300); // 300ms 防抖延迟
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSimulatedTime(e.target.value);
  };
  
  const selectedFund = fundsData.find(f => f.code === selectedFundCode);

  const ranges: { label: string; value: TimeRange }[] = [
    { label: '日', value: '1D' },
    { label: '周', value: '1W' },
    { label: '月', value: '1M' },
    { label: '季', value: '3M' },
  ];

  return (
    <div className="flex h-full bg-gray-50 w-full overflow-hidden relative touch-pan-y">
      {/* Sidebar - Fund List */}
      <div className={`
        bg-white border-r border-gray-200 flex flex-col z-10 transition-all duration-300
        ${selectedFundCode ? 'hidden md:flex md:w-80' : 'w-full md:w-80 flex'}
      `}>
        <div className="p-4 border-b border-gray-100 bg-white space-y-4 shrink-0">
          <div className="flex items-center justify-between">
            <h1 className="text-xl font-extrabold text-slate-800 flex items-center gap-2">
              <PieChart className="text-indigo-600" />
              <span>SmartFund</span>
            </h1>
            <div className="flex gap-2">
              <button 
                onClick={handleRefresh}
                className="p-2 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors text-gray-600"
                title="刷新数据"
              >
                <RefreshCw size={20} className={loadingList ? "animate-spin" : ""} />
              </button>
              <button 
                onClick={() => setIsSearchOpen(true)}
                className="p-2 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors text-gray-600"
                title="添加基金"
              >
                <Plus size={20} />
              </button>
            </div>
          </div>
          
          {/* Global Time Control */}
          <div className="bg-slate-50 p-2 rounded-lg border border-gray-200 flex items-center gap-2">
            <Clock size={16} className="text-gray-400" />
            <div className="flex-1 flex flex-col">
              <label className="text-[10px] text-gray-400 font-bold uppercase tracking-wider">模拟更新时间</label>
              <input 
                type="time" 
                value={simulatedTime}
                onChange={handleTimeChange}
                min="09:30"
                max="15:00"
                className="bg-transparent text-sm font-mono font-bold text-slate-700 outline-none w-full cursor-pointer"
              />
            </div>
          </div>

          <div className="text-xs text-gray-400 font-medium px-1">自选基金 ({fundsData.length})</div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 custom-scrollbar overscroll-contain" style={{ WebkitOverflowScrolling: 'touch' }}>
          {loadingList && fundsData.length === 0 ? (
            <div className="text-center py-10 text-gray-400">加载中...</div>
          ) : (
            fundsData.map(fund => (
              <FundCard 
                key={fund.code} 
                fund={fund} 
                isSelected={fund.code === selectedFundCode}
                onClick={() => setSelectedFundCode(fund.code)}
                onRemove={(e) => handleRemoveFund(e, fund.code)}
              />
            ))
          )}
          {fundsData.length === 0 && !loadingList && (
            <div className="text-center py-10 text-gray-400 text-sm">
              暂无自选基金<br/>点击右上角 + 添加
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className={`
         flex-col h-full overflow-hidden relative bg-slate-50
         ${selectedFundCode ? 'flex flex-1' : 'hidden md:flex md:flex-1'}
      `}>
        {selectedFund ? (
          <>
            {/* Header */}
            <div className="bg-white p-4 md:p-6 border-b border-gray-100 flex justify-between items-center shadow-sm z-10 sticky top-0 shrink-0">
              <div className="flex items-center gap-2 md:gap-0">
                {/* Mobile Back Button */}
                <button 
                  onClick={() => setSelectedFundCode(null)}
                  className="md:hidden mr-1 p-2 -ml-2 text-gray-500 hover:bg-gray-100 rounded-full active:scale-95 transition-transform"
                >
                  <ArrowLeft size={22} />
                </button>
                
                <div>
                  <h2 className="text-xl md:text-2xl font-bold text-slate-800 line-clamp-1">{selectedFund.name}</h2>
                  <div className="flex items-center gap-3 mt-1 text-sm text-gray-500">
                    <span className="bg-gray-100 px-2 py-0.5 rounded text-gray-600 text-xs md:text-sm">{selectedFund.type}</span>
                    <span className="font-mono hidden md:inline">{selectedFund.code}</span>
                    <span className="flex items-center gap-1 text-xs">
                      <Clock size={12} />
                      {simulatedTime} 更新
                    </span>
                  </div>
                </div>
              </div>
              <div className="text-right shrink-0 ml-2 flex items-center gap-3">
                <div className="flex flex-col items-end">
                  <div className={`text-2xl md:text-3xl font-bold font-mono ${selectedFund.growthRate >= 0 ? 'text-up' : 'text-down'}`}>
                    {selectedFund.currentValuation.toFixed(4)}
                  </div>
                  <div className={`text-sm font-bold ${selectedFund.growthRate >= 0 ? 'text-up' : 'text-down'}`}>
                    {selectedFund.growthRate >= 0 ? '+' : ''}{selectedFund.growthRate.toFixed(2)}%
                  </div>
                </div>
                {/* Mobile Refresh Button */}
                <button 
                  onClick={handleRefresh}
                  className="md:hidden p-2 text-gray-400 hover:text-indigo-600 hover:bg-gray-50 rounded-full transition-all"
                >
                  <RefreshCw size={20} className={loadingList ? "animate-spin" : ""} />
                </button>
              </div>
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 overflow-y-auto p-4 md:p-6 custom-scrollbar overscroll-contain" style={{ WebkitOverflowScrolling: 'touch' }}>
              <div className="max-w-6xl w-full mx-auto space-y-6">
                
                {/* Chart Section */}
                <div>
                   <div className="flex justify-between items-center mb-4">
                     <h3 className="font-bold text-gray-700">净值走势</h3>
                     <div className="flex bg-gray-100 rounded-lg p-1">
                       {ranges.map(r => (
                         <button
                           key={r.value}
                           onClick={() => setTimeRange(r.value)}
                           className={`px-3 py-1 text-xs font-medium rounded-md transition-all ${
                             timeRange === r.value 
                               ? 'bg-white text-indigo-600 shadow-sm' 
                               : 'text-gray-500 hover:text-gray-700'
                           }`}
                         >
                           {r.label}
                         </button>
                       ))}
                     </div>
                   </div>
                   
                   {loadingChart ? (
                     <div className="h-64 w-full bg-gray-100 rounded-xl animate-pulse"></div>
                   ) : (
                     <FundChart 
                        data={chartData} 
                        previousClose={selectedFund.previousClose}
                        color={selectedFund.growthRate >= 0 ? '#ef4444' : '#22c55e'}
                        timeRange={timeRange}
                     />
                   )}
                </div>

                {/* Info Grid */}
                <div className="grid grid-cols-3 gap-3 md:gap-4">
                  <div className="bg-white p-3 md:p-4 rounded-xl shadow-sm border border-gray-100">
                    <div className="text-xs text-gray-400 mb-1">昨日净值</div>
                    <div className="font-mono text-base md:text-lg font-semibold">{selectedFund.previousClose.toFixed(4)}</div>
                  </div>
                  <div className="bg-white p-3 md:p-4 rounded-xl shadow-sm border border-gray-100">
                    <div className="text-xs text-gray-400 mb-1">当前净值</div>
                    <div className={`font-mono text-base md:text-lg font-semibold ${selectedFund.growthRate >= 0 ? 'text-up' : 'text-down'}`}>
                      {selectedFund.currentValuation.toFixed(4)}
                    </div>
                  </div>
                  <div className="bg-white p-3 md:p-4 rounded-xl shadow-sm border border-gray-100">
                    <div className="text-xs text-gray-400 mb-1">日涨跌幅</div>
                    <div className={`font-mono text-base md:text-lg font-semibold ${selectedFund.growthRate >= 0 ? 'text-up' : 'text-down'}`}>
                      {selectedFund.growthRate.toFixed(2)}%
                    </div>
                  </div>
                </div>

                {/* Disclaimer */}
                <div className="mt-8 p-4 bg-yellow-50 rounded-lg border border-yellow-100 text-yellow-800 text-xs flex gap-2 items-start">
                   <AlertCircle size={14} className="mt-0.5 shrink-0" />
                   <p>
                     免责声明：本应用数据均为模拟演示数据，仅用于技术展示，不构成任何投资建议。
                     实际交易请以基金公司公告为准。
                   </p>
                </div>

              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-300">
            <PieChart size={64} className="mb-4 text-gray-200" />
            <p>请选择一只基金查看详情</p>
          </div>
        )}
      </div>

      {/* Search Modal Overlay - Moved to root level for correct z-index/stacking context */}
      {isSearchOpen && (
          <div className="absolute inset-0 z-50 bg-black/20 backdrop-blur-sm flex items-start justify-center pt-20 overflow-y-auto">
            <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl border border-gray-200 overflow-hidden transform transition-all scale-100 mx-4 my-4">
               <div className="p-4 border-b border-gray-100 flex items-center gap-3 sticky top-0 bg-white z-10">
                 <Search className="text-gray-400" size={20} />
                 <input 
                    type="text"
                    autoFocus
                    placeholder="输入代码或名称 (如: 005827)"
                    className="flex-1 outline-none text-lg text-gray-900 placeholder:text-gray-400 bg-white"
                    value={searchQuery}
                    onChange={handleSearch}
                 />
                 <button onClick={() => setIsSearchOpen(false)} className="text-sm text-gray-500 font-medium px-2 hover:text-gray-800">
                   取消
                 </button>
               </div>
               <div className="max-h-80 overflow-y-auto overscroll-contain" style={{ WebkitOverflowScrolling: 'touch' }}>
                 {isSearching && (
                   <div className="p-8 text-center text-gray-400 text-sm">搜索中...</div>
                 )}
                 {!isSearching && searchQuery && searchResults.length === 0 && (
                   <div className="p-8 text-center text-gray-400 text-sm">未找到相关基金</div>
                 )}
                 {!isSearching && searchResults.map(fund => (
                   <div 
                      key={fund.code}
                      onClick={() => handleAddFund(fund.code)}
                      className="p-4 hover:bg-gray-50 cursor-pointer border-b border-gray-50 flex justify-between items-center active:bg-gray-100"
                   >
                     <div>
                       <div className="text-gray-800 font-medium">{fund.name}</div>
                       <div className="text-xs text-gray-400 font-mono">{fund.code} - {fund.type}</div>
                     </div>
                     <Plus size={18} className="text-indigo-500" />
                   </div>
                 ))}
                 {!isSearching && !searchQuery && (
                   <div className="p-4 text-xs text-gray-400">
                     热门搜索：易方达、华夏、白酒、新能源...
                   </div>
                 )}
               </div>
            </div>
          </div>
        )}
    </div>
  );
};

export default App;