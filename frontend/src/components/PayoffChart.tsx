import React, { useMemo } from 'react';
import { AreaChart, Area, ReferenceLine, ResponsiveContainer, YAxis, XAxis, CartesianGrid, Tooltip } from 'recharts';

interface PayoffChartProps {
    trade: any; // Using any for flexibility during loose typing, can refine later
    currentPrice: number;
}

const generatePayoffData = (trade: any, currentPrice: number) => {
    const data = [];
    const legs = trade.legs || [];
    if (legs.length === 0) return data;

    // Domain: +/- 20%
    const centerPrice = currentPrice || 100;
    const minPrice = centerPrice * 0.7; // Wider range for builder
    const maxPrice = centerPrice * 1.3;

    const steps = 100;
    const stepSize = (maxPrice - minPrice) / steps;

    for (let i = 0; i <= steps; i++) {
        const price = minPrice + (i * stepSize);
        let profit = 0;

        // Initial Cost
        profit -= (trade.netDebit || 0) * 100; // Assuming netDebit is per share, but usually netDebit is total? 
        // Note: In previous files, netDebit seemed to be cost per share (e.g. $1.50).
        // Standard Option Math: Profit = (Value - Cost) * 100 * Qty.
        // Let's align with that. If trade.netDebit is total context, we adjust.
        // Looking at strategies.go, netDebit is calculated as Ask - Bid. That's per share.
        // So global debit = netDebit * 100 * Quantity (usually 1).

        // Wait, logic in TradeShowdown was: profit -= trade.netDebit.
        // But loops legs value = intrinsic * 100 * qty.
        // So TradeShowdown mixed units? netDebit usually ~1.00 (Example). Leg Value ~100.
        // If one is x100 and other isn't, chart is wrong.
        // Let's assume netDebit needs * 100.

        // CORRECTION: In TradeShowdown.jsx: `profit -= trade.netDebit` (line 32)
        // And `legValue = intrinsic * 100 * leg.quantity` (line 49)
        // This suggests TradeShowdown was bugged if netDebit was raw option price ($1.5).
        // Let's fix it here.

        const initialCost = (trade.netDebit || 0) * 100; // Fix unit scale
        profit -= initialCost;

        legs.forEach((leg: any) => {
            let valueAtExpiry = 0;
            const qty = leg.quantity || 1;

            if (leg.isStock) {
                valueAtExpiry = (price - (leg.stockPrice || price)) * qty; // Stock P/L relative to entry
                // Actually stock P/L is (CurrentPrice - EntryPrice) * Qty.
                // If we treat it as value: Value = Price * Qty.
                // Profit = Value - Cost.
                // Cost is included in netDebit?
                // TradeShowdown treated stock leg as: `(price - leg.stockPrice) * leg.quantity`.
                // This seems to be P/L directly, not Value.
                // If so, we just add it.
                profit += valueAtExpiry;
            } else {
                const strike = leg.option.strike;
                const type = leg.option.type || (leg.option.Type === 'Call' ? 'Call' : 'Put'); // Handle Go vs JS casing

                let intrinsic = 0;
                if (type === 'Call') {
                    intrinsic = Math.max(price - strike, 0);
                } else {
                    intrinsic = Math.max(strike - price, 0);
                }

                const legValue = intrinsic * 100 * qty;
                if (leg.action === 'Buy') {
                    profit += legValue;
                } else {
                    profit -= legValue; // Short option: We receive credit (in netDebit), but we owe Value.
                    // Profit = Credit - Value.
                    // Here: profit calced as -InitialCost.
                    // If Credit trade, InitialCost is negative (so -(-Credit) = +Credit).
                    // Then we substract Value at Expiry. Correct.
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

const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
        const profit = payload[0].value;
        const isProfit = profit >= 0;
        return (
            <div className="bg-zinc-900 border border-zinc-700 p-2 rounded shadow-xl backdrop-blur-md bg-opacity-90">
                <div className="text-zinc-400 text-xs mb-1">Price: <span className="text-white font-mono">${label}</span></div>
                <div className="text-xs">
                    P/L: <span className={`font-mono font-bold ${isProfit ? 'text-neon-green' : 'text-neon-red'}`}>
                        {isProfit ? '+' : ''}${Math.round(profit)}
                    </span>
                </div>
            </div>
        );
    }
    return null;
};

const PayoffChart: React.FC<PayoffChartProps> = ({ trade, currentPrice }) => {
    const data = useMemo(() => generatePayoffData(trade, currentPrice), [trade, currentPrice]);

    return (
        <div className="w-full h-full min-h-[400px] bg-[#050505] relative rounded-lg overflow-hidden border border-zinc-800">
            <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={data} margin={{ top: 20, right: 30, left: 10, bottom: 20 }}>
                    <defs>
                        <linearGradient id="neonGradient" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#0afff0" stopOpacity={0.2} />
                            <stop offset="95%" stopColor="#0afff0" stopOpacity={0} />
                        </linearGradient>
                        {/* Split gradient for + / - regions */}
                        <linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
                            {/* Recharts split gradient requires complex offset calculation based on 0-line.
                                For simplicity, we'll just use a single neon cyan theme or a conditional fill logic if easier.
                                Let's stick to "Neon Cyan Line" as requested.
                            */}
                            <stop offset="5%" stopColor="#06b6d4" stopOpacity={0.3} />
                            <stop offset="95%" stopColor="#06b6d4" stopOpacity={0} />
                        </linearGradient>
                    </defs>
                    <CartesianGrid stroke="#222" strokeDasharray="3 3" />
                    <XAxis
                        dataKey="price"
                        stroke="#444"
                        tick={{ fill: '#666', fontSize: 11, fontFamily: 'monospace' }}
                        tickFormatter={(val) => `$${val}`}
                        domain={['auto', 'auto']}
                    />
                    <YAxis
                        stroke="#444"
                        tick={{ fill: '#666', fontSize: 11, fontFamily: 'monospace' }}
                        tickFormatter={(val) => `$${val}`}
                        domain={['auto', 'auto']}
                    />
                    <Tooltip content={<CustomTooltip />} cursor={{ stroke: 'rgba(255,255,255,0.2)', strokeWidth: 1 }} />
                    <ReferenceLine y={0} stroke="#666" strokeWidth={1} />
                    <ReferenceLine x={currentPrice} stroke="#fcd34d" strokeDasharray="4 4" label={{ value: 'Curr', fill: '#fcd34d', fontSize: 10 }} />

                    <Area
                        type="monotone"
                        dataKey="profit"
                        stroke="#06b6d4" // Cyan Neon
                        strokeWidth={3}
                        fill="url(#neonGradient)"
                        filter="drop-shadow(0 0 6px rgba(6,182,212,0.5))" // Glow effect
                    />
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );
};

export default PayoffChart;
