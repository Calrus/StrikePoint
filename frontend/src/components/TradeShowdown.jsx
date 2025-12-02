import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer, ReferenceLine, Tooltip, ReferenceArea, CartesianGrid } from 'recharts';
import { TrendingUp, Shield, Zap, ArrowRight, DollarSign, Percent } from 'lucide-react';

const generatePayoffData = (trade, currentPrice) => {
    const data = [];
    const legs = trade.legs;
    if (!legs || legs.length === 0) return data;

    // Domain: +/- 20% of current price
    // If currentPrice is not provided, fallback to strike-based estimation
    let centerPrice = currentPrice;
    if (!centerPrice) {
        centerPrice = legs[0].option.strike;
    }

    const minPrice = centerPrice * 0.8;
    const maxPrice = centerPrice * 1.2;

    const steps = 50;
    const stepSize = (maxPrice - minPrice) / steps;

    for (let i = 0; i <= steps; i++) {
        const price = minPrice + (i * stepSize);
        let profit = 0;

        // Calculate profit based on strategy type (simplified)
        // Ideally we should use a more robust payoff calculator that handles all leg types
        // For now, we'll stick to the basic logic but apply it to all points

        let grossValue = 0;
        legs.forEach(leg => {
            const strike = leg.option.strike;
            const type = leg.option.type; // 'Call' or 'Put'
            const action = leg.action; // 'Buy' or 'Sell'

            let legValue = 0;
            if (type.toLowerCase() === 'call') {
                legValue = Math.max(price - strike, 0);
            } else {
                legValue = Math.max(strike - price, 0);
            }

            if (action === 'Buy') {
                grossValue += legValue;
            } else {
                grossValue -= legValue;
            }
        });

        profit = grossValue - trade.netDebit;

        data.push({
            price: parseFloat(price.toFixed(2)),
            profit: parseFloat(profit.toFixed(2))
        });
    }

    return data;
};

const MiniCard = ({ trade, type, isActive, onClick }) => {
    const isDegen = type === 'degen';
    const isLow = type === 'low';

    let borderColor = 'border-gray-800';
    let activeClass = '';
    let icon = <Shield size={16} className="text-blue-400" />;
    let badgeText = 'BALANCED';
    let badgeColor = 'bg-blue-500/10 text-blue-400';

    if (isLow) {
        icon = <TrendingUp size={16} className="text-success" />;
        badgeText = 'INCOME';
        badgeColor = 'bg-success/10 text-success';
    } else if (isDegen) {
        icon = <Zap size={16} className="text-primary" />;
        badgeText = 'AGGRESSIVE';
        badgeColor = 'bg-primary/10 text-primary';
    }

    if (isActive) {
        borderColor = isDegen ? 'border-primary' : isLow ? 'border-success' : 'border-blue-500';
        activeClass = 'bg-surface/80 shadow-lg scale-[1.02]';
    } else {
        activeClass = 'bg-surface/30 hover:bg-surface/50 opacity-70 hover:opacity-100';
    }

    return (
        <div
            onClick={onClick}
            className={`
                cursor-pointer p-4 rounded-xl border ${borderColor} ${activeClass}
                transition-all duration-200 mb-3 group
            `}
        >
            <div className="flex justify-between items-start mb-2">
                <div className="flex items-center gap-2">
                    {icon}
                    <span className="font-bold font-mono text-sm text-white">
                        {trade.riskProfile.split('(')[1].replace(')', '')}
                    </span>
                </div>
                <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full font-mono ${badgeColor}`}>
                    {badgeText}
                </span>
            </div>

            <div className="flex justify-between items-end">
                <div>
                    <span className="text-xs text-gray-500 font-mono block">MAX PROFIT</span>
                    <span className="text-sm font-bold text-gray-300 font-mono">{trade.maxProfit}</span>
                </div>
                <div className="text-right">
                    <span className="text-xs text-gray-500 font-mono block">ROI</span>
                    <span className={`text-sm font-bold font-mono ${isDegen ? 'text-primary' : 'text-white'}`}>
                        {trade.roi}
                    </span>
                </div>
            </div>
        </div>
    );
};

const DetailView = ({ trade, type, currentPrice }) => {
    const payoffData = generatePayoffData(trade, currentPrice);
    const isDegen = type === 'degen';
    const isLow = type === 'low';

    // Calculate Break-Even (approximate: find where profit crosses 0)
    // Simple scan
    let breakEvenPrice = null;
    for (let i = 0; i < payoffData.length - 1; i++) {
        if ((payoffData[i].profit < 0 && payoffData[i + 1].profit >= 0) ||
            (payoffData[i].profit >= 0 && payoffData[i + 1].profit < 0)) {
            breakEvenPrice = payoffData[i].price;
            break;
        }
    }

    // Determine chart min/max for ReferenceArea
    const profits = payoffData.map(d => d.profit);
    const maxProfitVal = Math.max(...profits, 100); // Ensure some height
    const minProfitVal = Math.min(...profits, -100);

    return (
        <div className="flex flex-col h-full animate-in fade-in duration-500">
            {/* Header Stats */}
            <div className="grid grid-cols-3 gap-4 mb-8">
                <div className="bg-surface/30 p-4 rounded-xl border border-gray-800">
                    <div className="flex items-center gap-2 mb-1 text-gray-400">
                        <DollarSign size={16} />
                        <span className="text-xs font-mono">NET DEBIT</span>
                    </div>
                    <span className="text-2xl font-bold font-mono text-white">${trade.netDebit.toFixed(2)}</span>
                </div>
                <div className="bg-surface/30 p-4 rounded-xl border border-gray-800">
                    <div className="flex items-center gap-2 mb-1 text-gray-400">
                        <TrendingUp size={16} />
                        <span className="text-xs font-mono">MAX PROFIT</span>
                    </div>
                    <span className="text-2xl font-bold font-mono text-white">{trade.maxProfit}</span>
                </div>
                <div className="bg-surface/30 p-4 rounded-xl border border-gray-800">
                    <div className="flex items-center gap-2 mb-1 text-gray-400">
                        <Percent size={16} />
                        <span className="text-xs font-mono">ROI POTENTIAL</span>
                    </div>
                    <span className={`text-2xl font-bold font-mono ${isDegen ? 'text-primary' : 'text-success'}`}>
                        {trade.roi}
                    </span>
                </div>
            </div>

            {/* Main Chart */}
            <div className="flex-grow bg-surface/20 rounded-xl p-6 border border-gray-800 mb-8 relative overflow-hidden">
                <div className="absolute top-4 left-6 z-10">
                    <h3 className="text-lg font-bold text-white font-mono flex items-center gap-2">
                        PAYOFF DIAGRAM <span className="text-xs text-gray-500 bg-black/50 px-2 py-1 rounded">AT EXPIRY</span>
                    </h3>
                </div>
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={payoffData} margin={{ top: 20, right: 30, left: 0, bottom: 0 }}>
                        <CartesianGrid stroke="#333" strokeDasharray="3 3" vertical={false} />

                        {/* Profit Zone (Green) */}
                        <ReferenceArea y1={0} y2={maxProfitVal * 1.5} fill="green" fillOpacity={0.05} />
                        {/* Loss Zone (Red) */}
                        <ReferenceArea y1={minProfitVal * 1.5} y2={0} fill="red" fillOpacity={0.05} />

                        <XAxis
                            dataKey="price"
                            stroke="#666"
                            tick={{ fill: '#666', fontSize: 12, fontFamily: 'monospace' }}
                            tickFormatter={(val) => `$${val}`}
                            domain={['auto', 'auto']}
                        />
                        <YAxis
                            stroke="#666"
                            tick={{ fill: '#666', fontSize: 12, fontFamily: 'monospace' }}
                            tickFormatter={(val) => `$${val}`}
                        />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#0a0a0a', border: '1px solid #333', borderRadius: '8px' }}
                            itemStyle={{ color: '#fff', fontFamily: 'monospace' }}
                            formatter={(value) => [`$${value}`, 'Profit/Loss']}
                            labelFormatter={(label) => `Price: $${label}`}
                        />

                        {/* Zero Line */}
                        <ReferenceLine y={0} stroke="#6b7280" strokeWidth={1} />

                        {/* Current Price Line */}
                        {currentPrice && (
                            <ReferenceLine x={currentPrice} stroke="#fff" strokeDasharray="3 3" label={{ value: 'CURRENT', position: 'top', fill: '#fff', fontSize: 10 }} />
                        )}

                        {/* Break-Even Line */}
                        {breakEvenPrice && (
                            <ReferenceLine x={breakEvenPrice} stroke="#fbbf24" strokeDasharray="3 3" label={{ value: 'BREAK-EVEN', position: 'top', fill: '#fbbf24', fontSize: 10 }} />
                        )}

                        <Line
                            type="monotone"
                            dataKey="profit"
                            stroke="#00ffcc"
                            strokeWidth={3}
                            dot={false}
                            style={{ filter: 'drop-shadow(0 0 8px rgba(0, 255, 204, 0.6))' }}
                        />
                    </LineChart>
                </ResponsiveContainer>
            </div>

            {/* Legs Detail */}
            <div className="bg-black/40 rounded-xl p-6 border border-gray-800">
                <h4 className="text-sm font-mono text-gray-400 mb-4 uppercase tracking-wider">Strategy Composition</h4>
                <div className="space-y-3">
                    {trade.legs.map((leg, idx) => (
                        <div key={idx} className="flex items-center justify-between p-3 bg-surface/30 rounded-lg border border-gray-800/50">
                            <div className="flex items-center gap-4">
                                <span className={`font-bold font-mono px-2 py-1 rounded text-xs ${leg.action === 'Buy' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                                    {leg.action.toUpperCase()}
                                </span>
                                <span className="text-white font-mono text-lg">{leg.option.strike} {leg.option.type.toUpperCase()}</span>
                            </div>
                            <div className="flex items-center gap-4 text-gray-400 font-mono text-sm">
                                <span>EXP: {leg.option.expiry}</span>
                                <ArrowRight size={14} />
                            </div>
                        </div>
                    ))}
                </div>
                <p className="mt-4 text-gray-400 text-sm leading-relaxed border-t border-gray-800 pt-4">
                    {trade.description}
                </p>
            </div>
        </div>
    );
};

const TradeShowdown = ({ trades, currentPrice }) => {
    const [activeIndex, setActiveIndex] = useState(0);

    if (!trades || trades.length === 0) return null;

    const getTradeType = (profile) => {
        if (profile.includes('Low')) return 'low';
        if (profile.includes('Medium')) return 'medium';
        if (profile.includes('Degen')) return 'degen';
        return 'medium';
    };

    const activeTrade = trades[activeIndex];
    // Fallback if currentPrice is missing (e.g. direct nav)
    const price = currentPrice || (activeTrade.legs[0]?.option.strike || 100);

    return (
        <div className="flex flex-col lg:flex-row gap-6 w-full max-w-7xl mx-auto h-[800px] animate-in fade-in slide-in-from-bottom-4 duration-700">
            {/* Sidebar (Master List) */}
            <div className="w-full lg:w-[350px] flex-shrink-0 flex flex-col bg-black/20 rounded-2xl border border-gray-800 overflow-hidden">
                <div className="p-4 border-b border-gray-800 bg-surface/30">
                    <h3 className="font-mono font-bold text-gray-400 text-xs uppercase tracking-widest">Available Strategies</h3>
                </div>
                <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
                    {trades.map((trade, idx) => (
                        <MiniCard
                            key={idx}
                            trade={trade}
                            type={getTradeType(trade.riskProfile)}
                            isActive={idx === activeIndex}
                            onClick={() => setActiveIndex(idx)}
                        />
                    ))}
                </div>
            </div>

            {/* Canvas (Detail View) */}
            <div className="flex-1 bg-black/20 rounded-2xl border border-gray-800 p-6 overflow-y-auto custom-scrollbar">
                <DetailView
                    key={activeIndex} // Force re-render for animation
                    trade={activeTrade}
                    type={getTradeType(activeTrade.riskProfile)}
                    currentPrice={price}
                />
            </div>
        </div>
    );
};

export default TradeShowdown;
