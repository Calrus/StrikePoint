import React, { useMemo } from 'react';
import { AreaChart, Area, ReferenceLine, ResponsiveContainer, YAxis, XAxis, CartesianGrid } from 'recharts';
import { ArrowRight, DollarSign, AlertTriangle } from 'lucide-react';

// --- Helper Functions ---

const generatePayoffData = (trade, currentPrice) => {
    const data = [];
    const legs = trade.legs;
    if (!legs || legs.length === 0) return data;

    // Domain: +/- 15% of current price for mini chart
    let centerPrice = currentPrice;
    if (!centerPrice && legs[0].option) {
        centerPrice = legs[0].option.strike;
    }
    if (!centerPrice && legs[0].stockPrice) {
        centerPrice = legs[0].stockPrice;
    }
    if (!centerPrice) centerPrice = 100;

    const minPrice = centerPrice * 0.85;
    const maxPrice = centerPrice * 1.15;

    const steps = 40; // Fewer steps for mini chart
    const stepSize = (maxPrice - minPrice) / steps;

    for (let i = 0; i <= steps; i++) {
        const price = minPrice + (i * stepSize);
        let profit = 0;

        profit -= trade.netDebit;

        legs.forEach(leg => {
            let valueAtExpiry = 0;
            if (leg.isStock) {
                valueAtExpiry = (price - leg.stockPrice) * leg.quantity;
            } else {
                const strike = leg.option.strike;
                const type = leg.option.type; // 'Call' or 'Put'

                let intrinsic = 0;
                if (type === 'Call') {
                    intrinsic = Math.max(price - strike, 0);
                } else {
                    intrinsic = Math.max(strike - price, 0);
                }

                const legValue = intrinsic * 100 * leg.quantity;
                if (leg.action === 'Buy') {
                    profit += legValue;
                } else {
                    profit -= legValue;
                }
            }
        });

        data.push({
            price: parseFloat(price.toFixed(2)),
            profit: parseFloat(profit.toFixed(2))
        });
    }

    return data;
};

// --- Components ---

const StrategyCard = ({ trade, currentPrice }) => {
    const payoffData = useMemo(() => generatePayoffData(trade, currentPrice), [trade, currentPrice]);

    // Calculate ROI (approximate for display)
    // If Debit > 0: ROI = MaxProfit / Debit. If MaxProfit is infinite, show "Unlimited"
    // If Credit (Debit < 0): ROI = Credit / Margin (Margin not fully calc'd yet, use MaxRisk)
    let roiDisplay = "N/A";
    let roiColor = "text-gray-400";

    if (trade.netDebit > 0) {
        // Debit Trade
        if (trade.maxProfit > 1000000) { // Arbitrary large number for "Unlimited"
            roiDisplay = "∞%";
            roiColor = "text-green-400";
        } else {
            const roi = (trade.maxProfit / trade.netDebit) * 100;
            roiDisplay = `${roi.toFixed(0)}%`;
            roiColor = roi > 0 ? "text-green-400" : "text-red-400";
        }
    } else {
        // Credit Trade
        // ROI = Credit / Max Risk
        if (trade.maxRisk > 0) {
            const credit = Math.abs(trade.netDebit);
            const roi = (credit / trade.maxRisk) * 100;
            roiDisplay = `${roi.toFixed(0)}%`;
            roiColor = "text-green-400";
        }
    }

    // Format legs summary
    const legsSummary = trade.legs.map(leg => {
        if (leg.isStock) return `${leg.action} Stock`;
        const type = leg.option.type === 'Call' ? 'C' : 'P';
        return `${leg.action} ${leg.option.strike}${type}`;
    }).join(', ');

    return (
        <div className="bg-surface/30 border border-gray-800 rounded-xl overflow-hidden hover:border-primary/50 transition-all duration-300 flex flex-col h-[320px] group">
            {/* Header */}
            <div className="p-4 border-b border-gray-800/50 bg-black/20 text-center relative">
                <h3 className="font-mono font-bold text-white text-sm truncate">{trade.name}</h3>
                <div className="text-[10px] font-mono text-gray-400 mt-1 truncate">{legsSummary}</div>
                <div className="absolute top-0 right-0 p-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <ArrowRight size={14} className="text-gray-500" />
                </div>
            </div>

            {/* Metrics Row */}
            <div className="grid grid-cols-2 gap-4 p-4 border-b border-gray-800/30">
                <div className="flex flex-col items-center">
                    <span className="text-[10px] font-mono text-gray-500 uppercase">Max Profit</span>
                    <span className="text-lg font-bold font-mono text-green-400">
                        {trade.maxProfit > 1000000 ? "Unlimited" : `$${trade.maxProfit.toFixed(0)}`}
                    </span>
                </div>
                <div className="flex flex-col items-center">
                    <span className="text-[10px] font-mono text-gray-500 uppercase">Max Risk</span>
                    <span className="text-lg font-bold font-mono text-gray-400">${trade.maxRisk.toFixed(0)}</span>
                </div>
            </div>

            {/* Mini Chart */}
            <div className="flex-grow relative w-full bg-[#050505]">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={payoffData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                        <defs>
                            <linearGradient id={`gradient-${trade.name}`} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                                <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                            </linearGradient>
                            <linearGradient id={`splitColor-${trade.name}`} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="0" stopColor="#10b981" stopOpacity={0.4} />
                                <stop offset="1" stopColor="#ef4444" stopOpacity={0.4} />
                            </linearGradient>
                        </defs>

                        <CartesianGrid stroke="#222" strokeDasharray="3 3" vertical={true} horizontal={true} />

                        <XAxis
                            dataKey="price"
                            type="number"
                            domain={['auto', 'auto']}
                            tick={{ fill: '#666', fontSize: 10, fontFamily: 'monospace' }}
                            tickFormatter={(val) => `$${val}`}
                            interval="preserveStartEnd"
                            minTickGap={20}
                            tickCount={6}
                        />
                        <YAxis
                            tick={{ fill: '#666', fontSize: 10, fontFamily: 'monospace' }}
                            tickFormatter={(val) => `$${val}`}
                            width={45}
                        />

                        <ReferenceLine y={0} stroke="#444" strokeWidth={1} />
                        {currentPrice && (
                            <ReferenceLine x={currentPrice} stroke="#666" strokeDasharray="2 2" />
                        )}

                        <Area
                            type="monotone"
                            dataKey="profit"
                            stroke="#10b981" // Default green stroke
                            fill={`url(#splitColor-${trade.name})`} // Simple vertical gradient for now
                            strokeWidth={2}
                        />
                    </AreaChart>
                </ResponsiveContainer>

                {/* Cost Overlay */}
                <div className="absolute bottom-2 right-2 bg-black/60 px-2 py-1 rounded text-[10px] font-mono text-gray-400 border border-gray-800">
                    {trade.netDebit > 0 ? `Debit: $${trade.netDebit.toFixed(0)}` : `Credit: $${Math.abs(trade.netDebit).toFixed(0)}`}
                </div>
            </div>

            {/* Footer Button */}
            <button className="w-full py-3 bg-primary/10 hover:bg-primary hover:text-white text-primary font-mono text-xs font-bold uppercase tracking-wider transition-all duration-200 border-t border-gray-800">
                Open in Builder
            </button>
        </div>
    );
};

const TradeShowdown = ({ trades, currentPrice }) => {
    // We can still filter if needed, but the request implies showing "all strategies" (filtered by backend sentiment).
    // The backend already filters by sentiment if provided.

    if (!trades || trades.length === 0) return null;

    // Use the first trade's strike or current price for chart centering if currentPrice is missing
    const price = currentPrice || (trades[0]?.legs[0]?.option?.strike || 100);

    return (
        <div className="w-full max-w-[1600px] mx-auto animate-in fade-in slide-in-from-bottom-4 duration-700 pb-20">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {trades.map((trade, idx) => (
                    <StrategyCard
                        key={idx}
                        trade={trade}
                        currentPrice={price}
                    />
                ))}
            </div>
        </div>
    );
};

export default TradeShowdown;
