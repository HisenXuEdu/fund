import React from 'react';
import { FundData } from '../types';
import { TrendingUp, TrendingDown, X } from 'lucide-react';

interface FundCardProps {
  fund: FundData;
  isSelected: boolean;
  onClick: () => void;
  onRemove: (e: React.MouseEvent) => void;
}

const FundCard: React.FC<FundCardProps> = ({ fund, isSelected, onClick, onRemove }) => {
  const isUp = fund.growthRate >= 0;
  const ColorIcon = isUp ? TrendingUp : TrendingDown;
  const textColor = isUp ? 'text-up' : 'text-down';
  const bgColor = isSelected ? 'bg-blue-50 border-blue-200' : 'bg-white border-gray-100 hover:border-gray-200';

  return (
    <div 
      onClick={onClick}
      className={`relative p-4 mb-3 rounded-xl border cursor-pointer transition-all duration-200 ${bgColor}`}
    >
      <div className="flex justify-between items-start gap-2">
        <div className="flex-1 min-w-0">
          <h3 className="font-bold text-gray-800 text-base truncate">{fund.name}</h3>
          <span className="text-xs text-gray-400 font-mono">{fund.code}</span>
        </div>
        
        <div className="flex items-start gap-3">
          <div className={`text-right ${textColor} shrink-0`}>
            <div className="text-lg font-bold flex items-center justify-end gap-1">
              <ColorIcon size={16} />
              {isUp ? '+' : ''}{fund.growthRate.toFixed(2)}%
            </div>
            <div className="text-xs opacity-80">
              估算 {fund.currentValuation.toFixed(4)}
            </div>
          </div>

          <button
            onClick={(e) => {
                e.stopPropagation();
                onRemove(e);
            }}
            className="text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-full p-1 transition-colors -mt-1 -mr-2"
            title="移除自选"
          >
            <X size={18} />
          </button>
        </div>
      </div>
    </div>
  );
};

export default FundCard;