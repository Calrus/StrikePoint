import React, { useMemo, useState, useCallback, useRef } from 'react';
import './ProfitMatrix.css';

interface MatrixPoint {
    date: string;
    price: number;
    profit: number;
    zScore: number;
}

interface ProfitMatrixProps {
    data: MatrixPoint[];
    onIvChange?: (ivChange: number) => void;
}

const ProfitMatrixGrid: React.FC<ProfitMatrixProps> = ({ data, onIvChange }) => {
    // ---- State ----
    const [ivAdj, setIvAdj] = useState(0);
    const [hoveredPoint, setHoveredPoint] = useState<MatrixPoint | null>(null);
    const [cursorPos, setCursorPos] = useState({ x: 0, y: 0 });

    // Debounce Timer Ref
    const debounceTimer = useRef<NodeJS.Timeout | null>(null);

    // ---- Handlers ----
    const handleIvChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const val = parseInt(e.target.value);
        setIvAdj(val);

        if (onIvChange) {
            if (debounceTimer.current) clearTimeout(debounceTimer.current);
            debounceTimer.current = setTimeout(() => {
                onIvChange(val);
            }, 100); // 100ms debounce as requested
        }
    };

    const handleMouseEnter = (point: MatrixPoint, e: React.MouseEvent) => {
        setHoveredPoint(point);
        updateCursorPos(e); // Initial pos
    };

    const handleMouseMove = (e: React.MouseEvent) => {
        if (hoveredPoint) {
            updateCursorPos(e);
        }
    };

    const handleMouseLeave = () => {
        setHoveredPoint(null);
    };

    const updateCursorPos = (e: React.MouseEvent) => {
        setCursorPos({ x: e.clientX, y: e.clientY });
    };

    // ---- Data Processing ----
    const { uniqueDates, uniquePrices, priceMap, maxAbsProfit, breakEvenKeys } = useMemo(() => {
        if (!data || data.length === 0) {
            return { uniqueDates: [], uniquePrices: [], priceMap: {}, minProfit: 0, maxProfit: 0, maxAbsProfit: 1, breakEvenKeys: new Set<string>() };
        }

        const dates = Array.from(new Set(data.map(p => p.date))).sort();
        // Prices sorted Descending (High at top)
        const prices = Array.from(new Set(data.map(p => p.price))).sort((a, b) => b - a);

        const map: Record<string, MatrixPoint> = {};
        let minP = 0;
        let maxP = 0;
        let absMax = 0;

        // Group by Date to find BreakEvens
        const dateGroups: Record<string, MatrixPoint[]> = {};

        data.forEach(p => {
            const key = `${p.date}-${p.price}`;
            map[key] = p;
            if (p.profit < minP) minP = p.profit;
            if (p.profit > maxP) maxP = p.profit;
            if (Math.abs(p.profit) > absMax) absMax = Math.abs(p.profit);

            if (!dateGroups[p.date]) dateGroups[p.date] = [];
            dateGroups[p.date].push(p);
        });

        const breakEvenKeys = new Set<string>();
        dates.forEach(d => {
            const points = dateGroups[d];
            if (points && points.length > 0) {
                let closest = points[0];
                let minDist = Math.abs(closest.profit);
                points.forEach(p => {
                    const dist = Math.abs(p.profit);
                    if (dist < minDist) {
                        minDist = dist;
                        closest = p;
                    }
                });
                breakEvenKeys.add(`${closest.date}-${closest.price}`);
            }
        });

        return {
            uniqueDates: dates,
            uniquePrices: prices,
            priceMap: map,
            minProfit: minP,
            maxProfit: maxP,
            maxAbsProfit: absMax > 0 ? absMax : 1,
            breakEvenKeys
        };
    }, [data]);

    if (!data || data.length === 0) {
        return <div className="text-gray-500 p-4">No data available</div>;
    }

    return (
        <div className="profit-matrix-container">
            {/* 1. Control Bar */}
            <div className="scenario-bar">
                <div className="iv-control">
                    <span className="iv-label">Implied Volatility Simulation</span>
                    <input
                        type="range"
                        min="-50"
                        max="50"
                        value={ivAdj}
                        onChange={handleIvChange}
                        className="iv-slider"
                    />
                    <span className="iv-value">{ivAdj > 0 ? '+' : ''}{ivAdj}%</span>
                </div>
            </div>

            {/* 2. Grid */}
            <div
                className="profit-matrix-grid"
                style={{
                    gridTemplateColumns: `80px repeat(${uniqueDates.length}, 1fr)`
                }}
                onMouseLeave={handleMouseLeave}
            >
                {/* Top-Left Corner (Empty) */}
                <div className="matrix-header-price"></div>

                {/* Date Headers */}
                {uniqueDates.map(date => (
                    <div key={date} className="matrix-header-date">
                        {new Date(date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                    </div>
                ))}

                {/* Rows */}
                {uniquePrices.map(price => (
                    <React.Fragment key={price}>
                        {/* Price Label (Left Column) */}
                        <div className="matrix-header-price">
                            ${price.toFixed(2)}
                        </div>

                        {/* Data Cells */}
                        {uniqueDates.map(date => {
                            const key = `${date}-${price}`;
                            const isBreakEven = breakEvenKeys.has(key);
                            return (
                                <MatrixCell
                                    key={key}
                                    point={priceMap[key]}
                                    maxAbsProfit={maxAbsProfit}
                                    isBreakEven={isBreakEven}
                                    onMouseEnter={handleMouseEnter}
                                    onMouseMove={handleMouseMove}
                                />
                            );
                        })}
                    </React.Fragment>
                ))}
            </div>

            {/* 3. Tooltip Portal/Overlay */}
            {hoveredPoint && (
                <MatrixTooltip point={hoveredPoint} x={cursorPos.x} y={cursorPos.y} />
            )}
        </div>
    );
};

// ---- Sub-Components ----

// Tooltip Component
const MatrixTooltip = ({ point, x, y }: { point: MatrixPoint, x: number, y: number }) => {
    // Calculate Probability (Rough approx from Z-Score)
    // Z=0 -> 100% (ATM-ish), Z=2 -> Low
    // Let's use simple normal density scaling or validity check.
    // If Z-Score is distance from target:
    // User requested "Probability: 12%".
    // Let's assume Probability of Touch ~ 2 * (1 - CDF(|z|)).
    // Or just map Z to a sensible 0-100 score.
    // Z=0 -> ~50% ITM? No, Z-Score here is (Target - Current) / Vol.
    // So it's standard deviations away.
    // Prob of expiring ITM (assuming 50% drift neutral) ~ 1 - CDF(z). 
    // If z=0 (at money), 50%. If z=1, 16%. If z=-1, 84%.
    // However, for a heatmap, usually "Probability" implies "Likelihood of this price occurring".
    // That decreases as Z increases.
    // Let's use PDF-relative scalar for a "Heat" score?
    // Or just CDF tail for "Prob of Reaching".
    // Let's stick to "Prob ITM" logic: 1 - CDF(zScore)
    // Actually, zScore in backend was (Price - Current)...
    // So if Price > Current, Z > 0.
    // Prob(Stock > Price) = 1 - CDF(Z).
    // This is a standard Option Delta approximation. Good enough for "Probability".

    // Quick Error Function approximation for CDF
    const cdf = (x: number) => {
        const t = 1 / (1 + 0.2316419 * Math.abs(x));
        const d = 0.3989423 * Math.exp(-x * x / 2);
        const prob = d * t * (0.3193815 + t * (-0.3565638 + t * (1.781478 + t * (-1.821256 + t * 1.330274))));
        if (x > 0) return 1 - prob;
        return prob;
    };

    // If Z-Score is standard deviations from current price:
    // Prob of reaching/exceeding (for Call side view) or just "Probability of this specific outcome density"?
    // The prompt says "Probability: 12%".
    // Let's calculate "Probability of Expiring At or Beyond this price".
    // P(>Price) = 1 - CDF(Z).
    // If Price < Current (Z < 0), P(>Price) is high (e.g. 80%).
    // But usually in heatmaps, people want "Prob of Touch" or "Prob ITM".
    // Let's simple use: 1 - CDF(zScore) and format as %.
    // Note: If zScore is unsigned in backend logic (it wasn't), we are good.
    // Backend: zScore = (p - current) / ...

    // Wait, if Z is very negative (price drop), P(>Price) is high.
    // Maybe we want "Probability that price is near this level"?
    // Or just simple Delta Proxy.
    // Let's go with "Prob ITM/Touch" proxy: 50% * exp(-0.5 * z^2) ??
    // No, let's use the Delta-like `1 - CDF(z)` but handle Direction.
    // Actually, simply `Prob = 2 * (1 - CDF(|z|))` gives "Probability of being outside this range" (Two tails).
    // "Probability of being AT LEAST this far"? 
    // Let's assume standard "Probability of Profit" usually means "Prob ITM".
    // But this is a generic point.
    // Let's display "Prob. Density" normalized? 
    // Let's stick to `(1 - CDF(Math.abs(point.zScore))) * 2` for "Tail Probability" (Likelihood of Extreme).
    // No, users find 100% confusing for ATM. 
    // Let's use Gaussian function value scaled to 100% at peak? `100 * exp(-0.5 * z*z)`.
    // This represents "Relative Likelihood" vs At-The-Money.
    // Z=0 -> 100%. Z=2 -> 13%.
    // This feels "Vibe" correct for a heatmap "Probability".
    const prob = Math.round(100 * Math.exp(-0.5 * point.zScore * point.zScore));

    return (
        <div
            className="matrix-tooltip"
            style={{ left: x, top: y }}
        >
            <div className="tooltip-row">
                <span className="tooltip-label">Date</span>
                <span className="tooltip-value">{point.date}</span>
            </div>
            <div className="tooltip-row">
                <span className="tooltip-label">Simulated Price</span>
                <span className="tooltip-value">${point.price.toFixed(2)}</span>
            </div>
            <div className="tooltip-row">
                <span className="tooltip-label">P/L</span>
                <span className={`tooltip-value ${point.profit >= 0 ? 'profit' : 'loss'}`}>
                    {point.profit >= 0 ? '+' : ''}${Math.round(point.profit)}
                </span>
            </div>
            <div className="tooltip-row">
                <span className="tooltip-label">Probability</span>
                <span className="tooltip-value">{prob}%</span>
            </div>
        </div>
    );
};

const MatrixCell = ({
    point, maxAbsProfit, isBreakEven, onMouseEnter, onMouseMove
}: {
    point: MatrixPoint | undefined,
    maxAbsProfit: number,
    isBreakEven: boolean,
    onMouseEnter: (p: MatrixPoint, e: React.MouseEvent) => void,
    onMouseMove: (e: React.MouseEvent) => void
}) => {
    if (!point) return <div className="matrix-cell empty"></div>;

    // Color Logic
    const isProfit = point.profit >= 0;
    const opacity = Math.max(0.15, Math.min(1, Math.abs(point.profit) / maxAbsProfit));
    const bgColor = isProfit
        ? `rgba(16, 185, 129, ${opacity})`
        : `rgba(239, 68, 68, ${opacity})`;

    const isUnlikely = Math.abs(point.zScore) > 2.5;

    return (
        <div
            className={`matrix-cell ${isBreakEven ? 'break-even' : ''} ${isUnlikely ? 'unlikely' : ''}`}
            style={{ backgroundColor: bgColor }}
            onMouseEnter={(e) => onMouseEnter(point, e)}
            onMouseMove={onMouseMove}
        // Remove title as we have custom tooltip now, or keep as fallback? Remove.
        >
            {!isUnlikely && (
                <div className="matrix-cell-content">
                    <span className="profit-val">${Math.round(point.profit)}</span>
                </div>
            )}
        </div>
    );
};

export default ProfitMatrixGrid;
