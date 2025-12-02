import React, { useState } from 'react';
import PredictionForm from '../components/PredictionForm';
import TradeShowdown from '../components/TradeShowdown';

const Calculator = () => {
    const [results, setResults] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [currentPrice, setCurrentPrice] = useState<number | null>(null);

    const handleCalculate = async (formData: any) => {
        setLoading(true);
        setError(null);
        setResults(null);
        setCurrentPrice(parseFloat(formData.currentPrice));

        try {
            const response = await fetch('http://localhost:8081/api/calculate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ticker: formData.ticker,
                    // currentPrice: parseFloat(formData.currentPrice), // Fetched by backend
                    targetPrice: parseFloat(formData.targetPrice),
                    date: formData.date,
                    sentiment: formData.sentiment
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to calculate trades');
            }

            const data = await response.json();
            setResults(data);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col items-center w-full max-w-[1600px] mx-auto px-4">


            <PredictionForm onCalculate={handleCalculate} loading={loading} />

            {error && (
                <div className="w-full max-w-4xl bg-red-500/10 border border-red-500/50 text-red-500 p-4 rounded-lg mb-8 text-center font-mono">
                    ERROR: {error}
                </div>
            )}

            {results && <TradeShowdown trades={results} currentPrice={currentPrice} />}

            {!results && !loading && !error && (
                <div className="text-center py-10 opacity-50">
                    <div className="w-16 h-16 border-2 border-gray-700 rounded-full mx-auto mb-4 border-t-primary animate-spin" />
                    <p className="text-sm font-mono text-gray-500">SYSTEM IDLE // AWAITING INPUT</p>
                </div>
            )}
        </div>
    );
};

export default Calculator;
