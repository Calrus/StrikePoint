import React, { useState, useEffect } from 'react';
import DateSlider from './DateSlider';
import { ArrowDown, ArrowDownRight, ArrowRight, Share2, ArrowUpRight, ArrowUp, RefreshCw } from 'lucide-react';

const PredictionForm = ({ onCalculate, loading }) => {
    const [formData, setFormData] = useState({
        ticker: 'SPY',
        currentPrice: '',
        priceChange: '+0.00%',
        targetPrice: '',
        budget: '',
        date: new Date().toISOString().split('T')[0],
        sentiment: 'bullish'
    });
    const [fetchingPrice, setFetchingPrice] = useState(false);

    // Debounce ticker change to fetch price
    useEffect(() => {
        const fetchPrice = async () => {
            if (!formData.ticker || formData.ticker.length < 2) return;

            setFetchingPrice(true);
            try {
                const response = await fetch(`http://localhost:8081/api/quote?ticker=${formData.ticker}`);
                if (response.ok) {
                    const data = await response.json();
                    if (data.price) {
                        setFormData(prev => ({
                            ...prev,
                            currentPrice: data.price,
                            // Mock price change for now as API might not return it
                            priceChange: Math.random() > 0.5 ? '+0.45%' : '-0.23%'
                        }));
                    }
                }
            } catch (error) {
                console.error("Error fetching price:", error);
            } finally {
                setFetchingPrice(false);
            }
        };

        const timeoutId = setTimeout(fetchPrice, 1000);
        return () => clearTimeout(timeoutId);
    }, [formData.ticker]);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        // If target price is empty, set a default based on sentiment for demo purposes
        let finalData = { ...formData };
        if (!finalData.targetPrice && finalData.currentPrice) {
            const price = parseFloat(finalData.currentPrice);
            if (formData.sentiment.includes('bullish')) finalData.targetPrice = (price * 1.05).toFixed(2);
            else if (formData.sentiment.includes('bearish')) finalData.targetPrice = (price * 0.95).toFixed(2);
            else finalData.targetPrice = price.toFixed(2);
        }
        onCalculate(finalData);
    };

    const sentiments = [
        { id: 'very_bearish', label: 'Very Bearish', icon: ArrowDown, color: 'text-red-500', bg: 'bg-red-500/10 border-red-500/50' },
        { id: 'bearish', label: 'Bearish', icon: ArrowDownRight, color: 'text-orange-500', bg: 'bg-orange-500/10 border-orange-500/50' },
        { id: 'neutral', label: 'Neutral', icon: ArrowRight, color: 'text-gray-400', bg: 'bg-gray-500/10 border-gray-500/50' },
        { id: 'directional', label: 'Directional', icon: Share2, color: 'text-purple-500', bg: 'bg-purple-500/10 border-purple-500/50' },
        { id: 'bullish', label: 'Bullish', icon: ArrowUpRight, color: 'text-green-500', bg: 'bg-green-500/10 border-green-500/50' },
        { id: 'very_bullish', label: 'Very Bullish', icon: ArrowUp, color: 'text-emerald-400', bg: 'bg-emerald-500/10 border-emerald-500/50' },
    ];

    return (
        <form onSubmit={handleSubmit} className="w-full max-w-5xl mx-auto mb-8 bg-black/40 p-6 rounded-2xl border border-gray-800 backdrop-blur-sm">

            {/* Line 1: Symbol & Current Price */}
            <div className="flex items-center gap-4 mb-6">
                <div className="flex-1 flex items-center gap-4">
                    <div className="relative">
                        <label className="absolute -top-2.5 left-3 bg-black px-1 text-[10px] font-mono text-gray-500">SYMBOL</label>
                        <input
                            type="text"
                            name="ticker"
                            value={formData.ticker}
                            onChange={handleChange}
                            className="w-32 bg-transparent border border-gray-700 rounded-lg px-3 py-2 text-2xl font-bold font-mono text-white uppercase focus:border-primary focus:outline-none"
                            placeholder="SPY"
                        />
                    </div>

                    <div className="flex items-baseline gap-2">
                        <span className="text-3xl font-bold text-white font-mono">
                            ${formData.currentPrice || '---.--'}
                        </span>
                        <span className={`text-sm font-mono ${formData.priceChange.startsWith('+') ? 'text-green-500' : 'text-red-500'}`}>
                            {formData.priceChange}
                        </span>
                        {fetchingPrice && <RefreshCw className="animate-spin text-gray-500 ml-2" size={14} />}
                    </div>
                </div>

                <div className="hidden md:flex items-center gap-2 text-xs text-gray-500 font-mono">
                    <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                    MARKET OPEN
                </div>
            </div>

            {/* Line 2: Sentiment Buttons */}
            <div className="mb-6">
                <div className="flex justify-between gap-2 overflow-x-auto pb-2 custom-scrollbar">
                    {sentiments.map((s) => (
                        <button
                            key={s.id}
                            type="button"
                            onClick={() => setFormData(prev => ({ ...prev, sentiment: s.id }))}
                            className={`
                                flex flex-col items-center justify-center w-24 h-24 rounded-full border-2 transition-all duration-200 flex-shrink-0
                                ${formData.sentiment === s.id
                                    ? `${s.bg} scale-105 shadow-[0_0_15px_rgba(0,0,0,0.5)]`
                                    : 'border-gray-800 bg-surface/20 hover:border-gray-600 hover:bg-surface/40'
                                }
                            `}
                        >
                            <div className={`p-2 rounded-full mb-1 ${formData.sentiment === s.id ? 'bg-black/20' : ''}`}>
                                <s.icon size={24} className={formData.sentiment === s.id ? s.color : 'text-gray-500'} />
                            </div>
                            <span className={`text-[10px] font-bold uppercase tracking-wider ${formData.sentiment === s.id ? 'text-white' : 'text-gray-500'}`}>
                                {s.label}
                            </span>
                        </button>
                    ))}
                </div>
            </div>

            {/* Line 3: Target & Budget */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div className="relative">
                    <label className="absolute -top-2.5 left-3 bg-black px-1 text-[10px] font-mono text-gray-500">TARGET PRICE</label>
                    <div className="flex items-center bg-surface/30 border border-gray-700 rounded-lg px-3 py-2 focus-within:border-primary transition-colors">
                        <span className="text-gray-500 mr-2">$</span>
                        <input
                            type="number"
                            name="targetPrice"
                            value={formData.targetPrice}
                            onChange={handleChange}
                            className="w-full bg-transparent text-lg font-bold font-mono text-white focus:outline-none"
                            placeholder={formData.currentPrice || "0.00"}
                        />
                        <span className="text-xs text-gray-500 font-mono ml-2">(+3.8%)</span>
                    </div>
                </div>

                <div className="relative">
                    <label className="absolute -top-2.5 left-3 bg-black px-1 text-[10px] font-mono text-gray-500">BUDGET</label>
                    <div className="flex items-center bg-surface/30 border border-gray-700 rounded-lg px-3 py-2 focus-within:border-primary transition-colors">
                        <span className="text-gray-500 mr-2">$</span>
                        <input
                            type="number"
                            name="budget"
                            value={formData.budget}
                            onChange={handleChange}
                            className="w-full bg-transparent text-lg font-bold font-mono text-white focus:outline-none"
                            placeholder="None"
                        />
                    </div>
                </div>
            </div>

            {/* Line 4: Date Slider */}
            <div className="mb-8">
                <DateSlider
                    selectedDate={formData.date}
                    onChange={(date) => setFormData(prev => ({ ...prev, date }))}
                />
            </div>

            {/* Calculate Button */}
            <button
                type="submit"
                disabled={loading}
                className={`
                    w-full py-3 rounded-lg font-bold text-lg tracking-widest font-mono uppercase
                    transition-all duration-300 transform hover:scale-[1.01] active:scale-[0.99]
                    ${loading
                        ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                        : 'bg-primary hover:bg-primary/90 text-white shadow-lg shadow-primary/20'
                    }
                `}
            >
                {loading ? 'Analyzing Market Data...' : 'Build Strategy'}
            </button>
        </form>
    );
};

export default PredictionForm;
