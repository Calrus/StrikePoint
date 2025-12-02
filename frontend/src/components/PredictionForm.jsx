import React, { useState, useEffect } from 'react';

const PredictionForm = ({ onCalculate, loading }) => {
    const [formData, setFormData] = useState({
        ticker: 'AAPL',
        currentPrice: '',
        targetPrice: 160,
        date: '2026-01-01'
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
                            currentPrice: data.price
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
        onCalculate(formData);
    };

    return (
        <form onSubmit={handleSubmit} className="w-full max-w-4xl mx-auto mb-12">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                {/* Ticker */}
                <div className="bg-surface/50 p-4 rounded-lg border border-gray-800 focus-within:border-primary transition-colors">
                    <label className="block text-xs text-gray-400 font-mono mb-1">TICKER</label>
                    <input
                        type="text"
                        name="ticker"
                        value={formData.ticker}
                        onChange={handleChange}
                        className="w-full bg-transparent text-2xl font-bold font-mono focus:outline-none uppercase"
                        placeholder="SPY"
                    />
                </div>

                {/* Current Price */}
                <div className="bg-surface/50 p-4 rounded-lg border border-gray-800 focus-within:border-primary transition-colors opacity-75">
                    <label className="block text-xs text-gray-400 font-mono mb-1">CURRENT PRICE</label>
                    <div className="flex items-center">
                        <span className="text-gray-500 mr-1">$</span>
                        <input
                            type="text"
                            disabled
                            value={fetchingPrice ? "LOADING..." : formData.currentPrice || "AUTO"}
                            className="w-full bg-transparent text-2xl font-bold font-mono focus:outline-none text-gray-400 cursor-not-allowed"
                        />
                    </div>
                </div>

                {/* Target Price */}
                <div className="bg-surface/50 p-4 rounded-lg border border-gray-800 focus-within:border-primary transition-colors">
                    <label className="block text-xs text-gray-400 font-mono mb-1">TARGET PRICE</label>
                    <div className="flex items-center">
                        <span className="text-gray-500 mr-1">$</span>
                        <input
                            type="number"
                            name="targetPrice"
                            value={formData.targetPrice}
                            onChange={handleChange}
                            className="w-full bg-transparent text-2xl font-bold font-mono focus:outline-none"
                            placeholder="0.00"
                        />
                    </div>
                </div>

                {/* Target Date */}
                <div className="bg-surface/50 p-4 rounded-lg border border-gray-800 focus-within:border-primary transition-colors">
                    <label className="block text-xs text-gray-400 font-mono mb-1">TARGET DATE</label>
                    <input
                        type="date"
                        name="date"
                        value={formData.date}
                        onChange={handleChange}
                        className="w-full bg-transparent text-lg font-bold font-mono focus:outline-none mt-1"
                    />
                </div>
            </div>

            <button
                type="submit"
                disabled={loading}
                className={`
          w-full py-4 rounded-lg font-bold text-xl tracking-widest font-mono
          transition-all duration-300 transform hover:scale-[1.01] active:scale-[0.99]
          ${loading
                        ? 'bg-gray-700 cursor-not-allowed text-gray-500'
                        : 'bg-primary hover:bg-primary/90 text-white shadow-[0_0_20px_rgba(139,92,246,0.5)] hover:shadow-[0_0_30px_rgba(139,92,246,0.7)]'
                    }
        `}
            >
                {loading ? 'CALCULATING...' : 'CALCULATE OPTIMAL TRADES'}
            </button>
        </form>
    );
};

export default PredictionForm;
