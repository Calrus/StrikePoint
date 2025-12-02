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
    const [ticker, setTicker] = useState('TSLA');

    useEffect(() => {
        const fetchSignals = async () => {
            setLoading(true);
            try {
                const response = await fetch(`http://localhost:8081/api/news?ticker=${ticker}`);
                const data = await response.json();
                // Map backend Article to frontend Signal if needed, but fields mostly match
                // Backend sends: id, ticker, title, link, published_at, sentiment_score, sentiment, confidence, reasoning
                setSignals(data || []);
            } catch (error) {
                console.error('Error fetching signals:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchSignals();
    }, [ticker]);

    const getTradeSuggestion = (sentiment: string) => {
        if (!sentiment) return 'Wait / Neutral';
        if (sentiment.toUpperCase() === 'BULLISH') return 'Medium Risk Call';
        if (sentiment.toUpperCase() === 'BEARISH') return 'Medium Risk Put';
        return 'Wait / Neutral';
    };

    return (
        <div className="space-y-8">
            <header className="mb-8 flex justify-between items-center">
                <div>
                    <h1 className="text-4xl font-bold font-mono tracking-tight mb-2">
                        NEWS <span className="text-primary">INTERPRETER</span>
                    </h1>
                    <p className="text-gray-400">
                        Real-time AI analysis of market-moving headlines.
                    </p>
                </div>
                <div className="flex gap-2">
                    {['TSLA', 'NVDA', 'SPY'].map(t => (
                        <button
                            key={t}
                            onClick={() => setTicker(t)}
                            className={`px-4 py-2 rounded-lg font-mono text-sm transition-colors ${ticker === t
                                    ? 'bg-primary text-white'
                                    : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                                }`}
                        >
                            {t}
                        </button>
                    ))}
                </div>
            </header>

            {loading ? (
                <div className="text-center py-12">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
                    <p className="text-gray-400 font-mono">Analyzing market chatter...</p>
                </div>
            ) : (
                <div className="grid gap-6">
                    {signals.map((signal, index) => (
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
                    {signals.length === 0 && (
                        <div className="text-center py-12 text-gray-500">
                            No signals detected yet.
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default NewsInterpreter;
