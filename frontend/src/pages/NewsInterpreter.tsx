import React, { useEffect, useState } from 'react';

interface Signal {
    id: number;
    ticker: string;
    title: string;
    link: string;
    published_at: string;
    sentiment_score: number;
    sentiment: string;
    confidence: number;
    reasoning: string;
}

const NewsInterpreter = () => {
    const [signals, setSignals] = useState<Signal[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedTickers, setSelectedTickers] = useState<string[]>(['TSLA', 'NVDA', 'SPY']);
    const [customTickers, setCustomTickers] = useState<string[]>([]);
    const [newTicker, setNewTicker] = useState('');
    const [isAdding, setIsAdding] = useState(false);

    const fetchSignals = async () => {
        setLoading(true);
        try {
            // Fetch all news
            const response = await fetch(`http://localhost:8081/api/news?ticker=&limit=100`);
            const data = await response.json();
            setSignals(data || []);
        } catch (error) {
            console.error('Error fetching signals:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSignals();
        // Poll every minute
        const interval = setInterval(fetchSignals, 60000);
        return () => clearInterval(interval);
    }, []);

    const toggleTicker = (t: string) => {
        if (selectedTickers.includes(t)) {
            setSelectedTickers(selectedTickers.filter(ticker => ticker !== t));
        } else {
            setSelectedTickers([...selectedTickers, t]);
        }
    };

    const handleAddTicker = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!newTicker) return;

        const tickerUpper = newTicker.toUpperCase();
        if (customTickers.includes(tickerUpper) || ['TSLA', 'NVDA', 'SPY'].includes(tickerUpper)) {
            setNewTicker('');
            setIsAdding(false);
            return;
        }

        try {
            await fetch(`http://localhost:8081/api/news/track?ticker=${tickerUpper}`, { method: 'POST' });
            setCustomTickers([...customTickers, tickerUpper]);
            setSelectedTickers([...selectedTickers, tickerUpper]);
            setNewTicker('');
            setIsAdding(false);
            // Trigger a refresh after a short delay to allow backend to start fetching
            setTimeout(fetchSignals, 2000);
        } catch (error) {
            console.error('Error adding ticker:', error);
        }
    };

    const getTradeSuggestion = (sentiment: string) => {
        if (!sentiment) return 'Wait / Neutral';
        if (sentiment.toUpperCase() === 'BULLISH') return 'Medium Risk Call';
        if (sentiment.toUpperCase() === 'BEARISH') return 'Medium Risk Put';
        return 'Wait / Neutral';
    };

    const filteredSignals = signals.filter(s => selectedTickers.includes(s.ticker));

    return (
        <div className="space-y-8">
            <header className="mb-8 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-4xl font-bold font-mono tracking-tight mb-2">
                        NEWS <span className="text-primary">INTERPRETER</span>
                    </h1>
                    <p className="text-gray-400">
                        Real-time AI analysis of market-moving headlines.
                    </p>
                </div>
                <div className="flex flex-wrap gap-2 items-center">
                    {['TSLA', 'NVDA', 'SPY', ...customTickers].map(t => (
                        <button
                            key={t}
                            onClick={() => toggleTicker(t)}
                            className={`px-4 py-2 rounded-lg font-mono text-sm transition-colors border ${selectedTickers.includes(t)
                                ? 'bg-primary border-primary text-white'
                                : 'bg-gray-800 border-gray-700 text-gray-400 hover:bg-gray-700'
                                }`}
                        >
                            {t}
                        </button>
                    ))}

                    {isAdding ? (
                        <form onSubmit={handleAddTicker} className="flex items-center gap-2">
                            <input
                                type="text"
                                value={newTicker}
                                onChange={(e) => setNewTicker(e.target.value)}
                                placeholder="TICKER"
                                className="w-24 px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-sm font-mono text-white focus:outline-none focus:border-primary uppercase"
                                autoFocus
                                onBlur={() => !newTicker && setIsAdding(false)}
                            />
                            <button type="submit" className="text-primary hover:text-white transition-colors">
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z" clipRule="evenodd" />
                                </svg>
                            </button>
                        </form>
                    ) : (
                        <button
                            onClick={() => setIsAdding(true)}
                            className="px-4 py-2 rounded-lg font-mono text-sm border border-dashed border-gray-600 text-gray-400 hover:border-primary hover:text-primary transition-colors flex items-center gap-2"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                                <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
                            </svg>
                            ADD
                        </button>
                    )}
                </div>
            </header>

            {loading && signals.length === 0 ? (
                <div className="text-center py-12">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
                    <p className="text-gray-400 font-mono">Analyzing market chatter...</p>
                </div>
            ) : (
                <div className="grid gap-6">
                    {filteredSignals.map((signal, index) => (
                        <div key={index} className="bg-gray-800/50 border border-gray-700 rounded-xl p-6 hover:border-primary/50 transition-colors">
                            <div className="flex justify-between items-start mb-4">
                                <div className="flex items-center gap-3">
                                    <span className="bg-primary/10 text-primary px-3 py-1 rounded-full text-sm font-bold font-mono">
                                        {signal.ticker}
                                    </span>
                                    <span className={`px-3 py-1 rounded-full text-sm font-bold font-mono ${signal.sentiment?.toUpperCase() === 'BULLISH' ? 'bg-green-500/10 text-green-500' :
                                        signal.sentiment?.toUpperCase() === 'BEARISH' ? 'bg-red-500/10 text-red-500' :
                                            'bg-gray-500/10 text-gray-500'
                                        }`}>
                                        {signal.sentiment?.toUpperCase() || 'NEUTRAL'}
                                    </span>
                                    <span className="text-xs text-gray-500 font-mono">
                                        {((signal.confidence || 0) * 100).toFixed(0)}% CONFIDENCE
                                    </span>
                                </div>
                                <span className="text-xs text-gray-500 font-mono">
                                    {new Date(signal.published_at).toLocaleTimeString()}
                                </span>
                            </div>

                            <h3 className="text-xl font-medium mb-4 text-white">
                                <a href={signal.link} target="_blank" rel="noopener noreferrer" className="hover:text-primary transition-colors">
                                    "{signal.title}"
                                </a>
                            </h3>

                            <div className="bg-black/30 rounded-lg p-4 mb-6 border border-gray-700/50">
                                <p className="text-gray-300 text-sm leading-relaxed">
                                    <span className="text-primary font-mono text-xs uppercase tracking-wider block mb-2">AI Reasoning</span>
                                    {signal.reasoning || 'Analysis pending...'}
                                </p>
                            </div>

                            <button className="w-full sm:w-auto bg-primary hover:bg-primary/90 text-background font-bold py-3 px-6 rounded-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] font-mono uppercase tracking-wide">
                                Suggested Trade: {getTradeSuggestion(signal.sentiment)}
                            </button>
                        </div>
                    ))}
                    {filteredSignals.length === 0 && (
                        <div className="text-center py-12 text-gray-500">
                            {signals.length > 0 ? 'No news for selected tickers.' : 'No signals detected yet.'}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default NewsInterpreter;
