import React, { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { ChevronLeft, Edit2, X, Plus, Activity, Zap, TrendingUp, Save } from 'lucide-react';
import ProfitMatrixGrid from '../components/ProfitMatrix';
import PayoffChart from '../components/PayoffChart';
import Sidebar from '../components/Sidebar';

// API Call to get Heatmap Data
const fetchMatrixData = async (trade: any, currentPrice: number, vol: number) => {
    try {
        const response = await fetch('http://localhost:8081/api/simulate/matrix', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                strategy: trade,
                currentPrice: currentPrice,
                volatility: vol
            })
        });

        if (!response.ok) {
            console.error("Failed to fetch matrix data");
            return [];
        }

        const data = await response.json();
        return data; // Should match MatrixPoint[] structure from backend
    } catch (e) {
        console.error("Error fetching matrix:", e);
        return [];
    }
};

const Builder: React.FC = () => {
    const location = useLocation();
    const navigate = useNavigate();

    // Initial State: Location State -> LocalStorage -> Default
    const [trade, setTrade] = useState(() => {
        if (location.state?.trade) return location.state.trade;
        try {
            const saved = localStorage.getItem('builder_trade_scratchpad');
            return saved ? JSON.parse(saved) : { name: "Custom Strategy", legs: [], netDebit: 0 };
        } catch {
            return { name: "Custom Strategy", legs: [], netDebit: 0 };
        }
    });

    const initialPrice = location.state?.currentPrice || 100;
    const [currentPrice] = useState(initialPrice);
    const [activeTab, setActiveTab] = useState<'payoff' | 'matrix'>('payoff');
    const [matrixData, setMatrixData] = useState<any[]>([]);
    const [ivAdj, setIvAdj] = useState(0);

    // Persist to Scratchpad
    useEffect(() => {
        localStorage.setItem('builder_trade_scratchpad', JSON.stringify(trade));
    }, [trade]);

    const saveToWatchlist = () => {
        const existing = localStorage.getItem('saved_drafts');
        const drafts = existing ? JSON.parse(existing) : [];
        // Add minimal unique ID or timestamp if saving multiple times?
        // For now just push.
        drafts.push({ ...trade, savedAt: new Date().toISOString() });
        localStorage.setItem('saved_drafts', JSON.stringify(drafts));

        // Dispatch event for Sidebar to pick up
        window.dispatchEvent(new Event('drafts-updated'));

        // Simple visual feedback could be added here
        // alert('Saved to Watchlist'); 
    };

    // Simulate or Fetch Matrix Data Load
    useEffect(() => {
        if (activeTab === 'matrix') {
            const vol = 0.3 + (ivAdj / 100) * 0.3; // Base Vol 30% +/- adj
            fetchMatrixData(trade, currentPrice, vol).then(data => setMatrixData(data));
        }
    }, [trade, currentPrice, activeTab, ivAdj]);

    return (
        <div className="flex flex-col h-full overflow-y-auto w-full">
            {/* Header / Editor Bar */}
            <div className="bg-surface border-b border-border p-4 flex items-center gap-4">
                <button onClick={() => navigate(-1)} className="p-2 hover:bg-white/10 rounded-full transition-colors">
                    <ChevronLeft size={20} className="text-gray-400" />
                </button>

                {/* Ticker & Expiry Badge */}
                <div className="flex flex-col border-r border-gray-800 pr-4 mr-2">
                    <span className="text-2xl font-black text-white tracking-tighter">
                        {trade.legs?.[0]?.option?.underlying || "SPY"}
                    </span>
                    {trade.expirationDate && (
                        <div className="relative group/badge">
                            <span className="text-xs font-mono font-bold text-primary bg-primary/10 px-2 py-1 rounded cursor-pointer hover:bg-primary/20 transition-colors border border-primary/20">
                                🗓️ EXP: {new Date(trade.expirationDate).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }).toUpperCase()}
                            </span>

                            {/* Hover/Click Dropdown */}
                            <div className="hidden group-hover/badge:block absolute top-full left-0 mt-2 w-48 bg-surface border border-gray-700 rounded-lg shadow-xl p-3 z-50">
                                <div className="text-[10px] uppercase text-gray-500 font-bold mb-1">Contract Expiry</div>
                                <div className="text-white font-mono text-sm">{new Date(trade.expirationDate).toLocaleDateString()}</div>
                                <div className="h-px bg-gray-800 my-2"></div>
                                <div className="text-[10px] uppercase text-gray-500 font-bold mb-1">Target Date</div>
                                <div className="text-gray-400 font-mono text-xs">Matching Closest Expiry</div>
                            </div>
                        </div>
                    )}
                </div>

                <div className="flex-1 overflow-x-auto flex items-center gap-3">
                    {trade.legs && trade.legs.map((leg: any, idx: number) => {
                        const legDate = leg.option?.expiry ? new Date(leg.option.expiry) : null;
                        const dateStr = legDate ? legDate.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) : '';
                        const typeLabel = leg.option?.type === 'Call' ? 'CALL' : 'PUT';

                        return (
                            <div key={idx} className="flex items-center gap-3 px-4 py-2 bg-black/40 border border-gray-700 rounded-lg group hover:border-primary/50 transition-all min-w-[200px]">
                                <div className="flex flex-col w-full">
                                    <span className={`text-[10px] font-bold uppercase tracking-wider ${leg.action === 'Buy' ? 'text-green-400' : 'text-red-400'}`}>
                                        {leg.action} {leg.quantity}x
                                    </span>
                                    <div className="flex items-center justify-between w-full gap-4">
                                        <span className="font-mono text-sm font-bold text-white whitespace-nowrap">
                                            {leg.option?.strike} {typeLabel}
                                        </span>
                                        {dateStr && (
                                            <span className="text-[10px] font-mono text-gray-500">
                                                @ {dateStr}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                <div className="ml-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <button className="p-1 hover:text-primary"><Edit2 size={12} /></button>
                                    <button className="p-1 hover:text-red-500"><X size={12} /></button>
                                </div>
                            </div>
                        )
                    })}

                    <button className="flex items-center gap-2 px-4 py-2 bg-primary/10 border border-dashed border-primary/30 rounded-lg hover:bg-primary/20 text-primary transition-all">
                        <Plus size={16} />
                        <span className="text-xs font-bold uppercase">Add Leg</span>
                    </button>
                </div>

                <button
                    onClick={saveToWatchlist}
                    className="flex items-center gap-2 px-4 py-2 bg-surface/50 border border-gray-700 hover:border-primary hover:text-primary text-gray-300 rounded-lg transition-all group"
                    title="Save to Watchlist"
                >
                    <Save size={16} className="group-hover:scale-110 transition-transform" />
                    <span className="hidden lg:inline text-xs font-bold uppercase tracking-wider">Save</span>
                </button>

                <div className="w-px h-8 bg-gray-800 mx-2"></div>

                <div className="text-right">
                    <div className="text-[10px] text-gray-500 uppercase">Est. Cost</div>
                    <div className={`font-mono font-bold ${trade.netDebit > 0 ? 'text-red-400' : 'text-green-400'}`}>
                        {trade.netDebit > 0 ? `DEBIT $${trade.netDebit.toFixed(2)}` : `CREDIT $${Math.abs(trade.netDebit).toFixed(2)}`}
                    </div>
                </div>
            </div>

            {/* Middle - Content */}
            <div className="flex-1 flex flex-col min-h-0">
                {/* Tabs */}
                <div className="flex items-center gap-1 p-2 bg-black/20 border-b border-border">
                    <button
                        onClick={() => setActiveTab('payoff')}
                        className={`px-6 py-2 text-sm font-bold uppercase tracking-wide rounded-md transition-all ${activeTab === 'payoff' ? 'bg-primary text-black shadow-[0_0_15px_rgba(34,211,238,0.4)]' : 'text-gray-400 hover:text-white'}`}
                    >
                        Payoff Graph
                    </button>
                    <button
                        onClick={() => setActiveTab('matrix')}
                        className={`px-6 py-2 text-sm font-bold uppercase tracking-wide rounded-md transition-all ${activeTab === 'matrix' ? 'bg-primary text-black shadow-[0_0_15px_rgba(34,211,238,0.4)]' : 'text-gray-400 hover:text-white'}`}
                    >
                        Profit Matrix
                    </button>
                </div>

                {/* Viewport */}
                <div className="flex-1 relative overflow-hidden bg-[#050505]">
                    {activeTab === 'payoff' ? (
                        <div className="p-6 h-full w-full">
                            <PayoffChart trade={trade} currentPrice={currentPrice} />
                        </div>
                    ) : (
                        <div className="h-full w-full p-4">
                            <ProfitMatrixGrid
                                data={matrixData}
                                onIvChange={(val: number) => setIvAdj(val)}
                            />
                        </div>
                    )}
                </div>
            </div>

            {/* Bottom - Greeks */}
            <div className="h-16 bg-surface border-t border-border flex items-center justify-around px-8">
                <GreekItem label="Delta" value="+4.25" icon={Activity} />
                <GreekItem label="Gamma" value="+0.08" icon={Zap} />
                <GreekItem label="Theta" value="-12.30" icon={TrendingUp} color="text-red-400" />
                <GreekItem label="Vega" value="+8.45" icon={Activity} />
            </div>
        </div>
    );
};

const GreekItem = ({ label, value, icon: Icon, color = "text-white" }: any) => (
    <div className="flex items-center gap-3">
        <div className="p-2 bg-white/5 rounded-full text-gray-400">
            <Icon size={16} />
        </div>
        <div>
            <div className="text-[10px] text-gray-500 uppercase font-bold tracking-wider">{label}</div>
            <div className={`font-mono text-sm font-bold ${color}`}>{value}</div>
        </div>
    </div>
);

export default Builder;
