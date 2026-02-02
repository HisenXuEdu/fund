import { GoogleGenAI, Type } from "@google/genai";
import { FundData, AIAnalysisResult, ChartDataPoint } from '../types';

export const analyzeFundTrend = async (fund: FundData, chartData: ChartDataPoint[]): Promise<AIAnalysisResult> => {
  // Initialize GoogleGenAI with API key from process.env directly
  const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });
  
  // Sample the chart data to reduce token count (take every 20th point)
  const sampledTrend = chartData
    .filter((_, i) => i % 20 === 0)
    .map(p => `${p.time}:${p.value.toFixed(4)}`)
    .join(', ');

  const prompt = `
    Role: Professional Financial Analyst.
    Task: Analyze the simulated intraday performance of a Chinese Mutual Fund.
    
    Fund Information:
    - Name: ${fund.name} (${fund.code})
    - Type: ${fund.type}
    - Previous Close: ${fund.previousClose.toFixed(4)}
    - Current Valuation: ${fund.currentValuation.toFixed(4)}
    - Intraday Growth: ${fund.growthRate.toFixed(2)}%
    
    Intraday Trend Sample (Time:Value):
    [${sampledTrend}]
    
    Please provide a response in JSON format.
    
    Language: Chinese (Simplified).
  `;

  try {
    const response = await ai.models.generateContent({
      model: 'gemini-3-flash-preview',
      contents: prompt,
      config: {
        responseMimeType: 'application/json',
        responseSchema: {
          type: Type.OBJECT,
          properties: {
            summary: { type: Type.STRING, description: "A concise summary of the day's performance (max 100 characters)." },
            advice: { type: Type.STRING, description: "Strategic advice for an investor (Hold/Buy/Sell consideration)." },
            bullish: { type: Type.BOOLEAN, description: "True if the trend looks positive/stable, false if negative/volatile." },
          },
          required: ["summary", "advice", "bullish"],
        }
      }
    });

    const text = response.text;
    const result = text ? JSON.parse(text) as AIAnalysisResult : {
       summary: "数据解析失败",
       advice: "请稍后重试",
       bullish: false
    };
    return result;
  } catch (error) {
    console.error("Gemini Analysis Error:", error);
    return {
      summary: "AI服务暂时不可用，请稍后再试。",
      advice: "建议关注市场整体走势。",
      bullish: fund.growthRate >= 0
    };
  }
};