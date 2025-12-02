import React from 'react';
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer, ReferenceLine, Tooltip } from 'recharts';

const generatePayoffData = (trade, type) => {
    const data = [];
    const legs = trade.legs;
    if (!legs || legs.length === 0) return data;

    // Determine key price points
    let minPrice, maxPrice, strike;

    if (type === 'degen') {
        strike = legs[0].option.strike;
        minPrice = strike * 0.8;
        maxPrice = strike * 1.4;
    } else {
        // Spread / PMCC
        const longStrike = legs[0].option.strike;
        const shortStrike = legs[1].option.strike;
        minPrice = longStrike * 0.8;
        maxPrice = shortStrike * 1.2;
    }

    const steps = 20;
    const stepSize = (maxPrice - minPrice) / steps;

    for (let i = 0; i <= steps; i++) {
        const price = minPrice + (i * stepSize);
        let profit = 0;

        if (type === 'degen') {
            // Long Call: max(S - K, 0) - Debit
            const strike = legs[0].option.strike;
            const cost = trade.netDebit;
            profit = Math.max(price - strike, 0) - cost;
        } else {
            // Spread: (Long Value - Short Value) - Debit
            // Simplified intrinsic value model
            const longStrike = legs[0].option.strike;
            const shortStrike = legs[1].option.strike;
            const cost = trade.netDebit;

            const longValue = Math.max(price - longStrike, 0);
            const shortValue = Math.max(price - shortStrike, 0);

            profit = (longValue - shortValue) - cost;
        }

        data.push({
            price: price.toFixed(2),
            profit: profit
        });
    }

    return data;
};

const TradeCard = ({ trade, type }) => {
    const isDegen = type === 'degen';
    const isLow = type === 'low';
    const isMedium = type === 'medium';

    let borderColor = 'border-gray-700';
    let shadowColor = '';
    let titleColor = 'text-gray-300';
    let badgeColor = 'bg-gray-800 text-gray-300';
    let chartColor = '#9ca3af'; // gray-400

    if (isLow) {
        borderColor = 'border-success';
        titleColor = 'text-success';
        badgeColor = 'bg-success/20 text-success';
        chartColor = '#10b981';
    } else if (isMedium) {
        borderColor = 'border-blue-500';
        titleColor = 'text-blue-500';
        badgeColor = 'bg-blue-500/20 text-blue-400';
        chartColor = '#3b82f6';
    } else if (isDegen) {
        borderColor = 'border-primary';
        shadowColor = 'shadow-[0_0_30px_rgba(139,92,246,0.3)]';
        titleColor = 'text-primary';
        badgeColor = 'bg-primary/20 text-primary';
        chartColor = '#8b5cf6';
    }

    const payoffData = generatePayoffData(trade, type);

    return (
        <div className={`
      relative bg-surface/40 backdrop-blur-sm rounded-xl p-6 border-2 ${borderColor} ${shadowColor}
      flex flex-col h-full transition-all duration-300 hover:-translate-y-1
    `}>
            {/* Header */}
            <div className="flex justify-between items-start mb-4">
                <div>
                    <h3 className={`text-xl font-bold font-mono ${titleColor} uppercase tracking-wider`}>
                        {trade.riskProfile.split('(')[1].replace(')', '')}
                    </h3>
                    <span className="text-xs text-gray-500 font-mono uppercase">{trade.riskProfile.split('(')[0]} RISK</span>
                </div>
                <div className={`px-3 py-1 rounded-full text-xs font-bold font-mono ${badgeColor}`}>
                    {isDegen ? 'MOONSHOT' : isLow ? 'INCOME' : 'BALANCED'}
                </div>
            </div>

            {/* Description */}
            <p className="text-sm text-gray-400 mb-6 min-h-[40px] leading-relaxed">
                {trade.description}
            </p>

            {/* Chart */}
            <div className="h-32 w-full mb-6 -ml-2">
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={payoffData}>
                        <XAxis
                            dataKey="price"
                            hide={true}
                        />
                        <YAxis
                            hide={true}
                            domain={['auto', 'auto']}
                        />
                        <ReferenceLine y={0} stroke="#4b5563" strokeDasharray="3 3" />
                        <Tooltip
                            contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px' }}
                            itemStyle={{ color: '#fff', fontFamily: 'monospace' }}
                            labelStyle={{ display: 'none' }}
                            formatter={(value) => [`$${parseFloat(value).toFixed(2)}`, 'Profit/Loss']}
                        />
                        <Line
                            type="monotone"
                            dataKey="profit"
                            stroke={chartColor}
                            strokeWidth={2}
                            dot={false}
                        />
                    </LineChart>
                </ResponsiveContainer>
                <div className="text-center text-[10px] text-gray-500 font-mono mt-1">PAYOFF DIAGRAM AT EXPIRY</div>
            </div>

            {/* Legs */}
            <div className="space-y-3 mb-6 flex-grow">
                {trade.legs.map((leg, idx) => (
                    <div key={idx} className="flex justify-between items-center text-sm font-mono border-b border-gray-800 pb-2 last:border-0">
                        <div className="flex items-center gap-2">
                            <span className={leg.action === 'Buy' ? 'text-green-400' : 'text-red-400'}>
                                {leg.action.toUpperCase()}
                            </span>
                            <span className="text-gray-300">{leg.option.strike} {leg.option.type.toUpperCase()}</span>
                        </div>
                        <span className="text-gray-500">{leg.option.expiry}</span>
                    </div>
                ))}
            </div>

            {/* Stats */}
            <div className="space-y-4 pt-4 border-t border-gray-800">
                <div className="flex justify-between items-center">
                    <span className="text-xs text-gray-500 font-mono">NET DEBIT</span>
                    <span className="text-xl font-bold font-mono text-white">${trade.netDebit.toFixed(2)}</span>
                </div>

                <div className="flex justify-between items-center">
                    <span className="text-xs text-gray-500 font-mono">MAX PROFIT</span>
                    <span className="text-sm font-bold font-mono text-gray-300">{trade.maxProfit}</span>
                </div>

                {/* ROI / Win Prob */}
                <div className={`
          mt-4 p-4 rounded-lg text-center
          ${isDegen ? 'bg-primary/10 border border-primary/30' : 'bg-surface/50 border border-gray-700'}
        `}>
                    <span className="block text-xs text-gray-500 font-mono mb-1">
                        {isLow ? 'EST. ANNUALIZED RETURN' : 'POTENTIAL ROI'}
                    </span>
                    <span className={`
            font-mono font-bold
            ${isDegen ? 'text-4xl text-primary animate-pulse' : 'text-2xl text-white'}
          `}>
                        {trade.roi}
                    </span>
                </div>
            </div>
        </div>
    );
};

const TradeShowdown = ({ trades }) => {
    if (!trades || trades.length === 0) return null;

    // Map backend risk profiles to our internal types for styling
    const getTradeType = (profile) => {
        if (profile.includes('Low')) return 'low';
        if (profile.includes('Medium')) return 'medium';
        if (profile.includes('Degen')) return 'degen';
        return 'medium';
    };

    return (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 w-full max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-700">
            {trades.map((trade, idx) => (
                <TradeCard key={idx} trade={trade} type={getTradeType(trade.riskProfile)} />
            ))}
        </div>
    );
};

export default TradeShowdown;
