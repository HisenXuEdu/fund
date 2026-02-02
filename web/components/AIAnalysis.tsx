import React from 'react';
import { AIAnalysisResult } from '../types';
import { Bot, Sparkles } from 'lucide-react';

interface AIAnalysisProps {
  analysis: AIAnalysisResult | null;
  loading: boolean;
  onAnalyze: () => void;
}

const AIAnalysis: React.FC<AIAnalysisProps> = ({ analysis, loading, onAnalyze }) => {
  if (loading) {
    return (
      <div className="mt-4 p-4 bg-indigo-50 rounded-xl border border-indigo-100 flex items-center justify-center gap-3 animate-pulse">
        <Bot className="text-indigo-500 animate-bounce" />
        <span className="text-indigo-700 text-sm font-medium">Gemini 正在分析盘面数据...</span>
      </div>
    );
  }

  if (!analysis) {
    return (
      <button 
        onClick={onAnalyze}
        className="mt-4 w-full py-3 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-xl shadow-md hover:shadow-lg transition-all flex items-center justify-center gap-2 font-medium"
      >
        <Sparkles size={18} />
        生成 AI 智能解读
      </button>
    );
  }

  return (
    <div className="mt-4 bg-white rounded-xl border border-indigo-100 shadow-sm overflow-hidden">
      <div className="bg-indigo-50 px-4 py-2 border-b border-indigo-100 flex items-center gap-2">
        <Bot size={16} className="text-indigo-600" />
        <span className="text-xs font-bold text-indigo-700 uppercase tracking-wider">Gemini Insight</span>
      </div>
      <div className="p-4 space-y-3">
        <div>
          <h4 className="text-xs text-gray-500 mb-1">行情摘要</h4>
          <p className="text-sm text-gray-800 leading-relaxed">{analysis.summary}</p>
        </div>
        <div>
          <h4 className="text-xs text-gray-500 mb-1">操作建议</h4>
          <div className={`text-sm p-2 rounded-lg ${analysis.bullish ? 'bg-red-50 text-red-800' : 'bg-green-50 text-green-800'}`}>
            {analysis.advice}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AIAnalysis;