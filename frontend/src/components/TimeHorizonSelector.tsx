import React, { useState, useEffect } from 'react';
import { Zap, Calendar, Gem } from 'lucide-react';

interface TimeHorizonSelectorProps {
    selectedDate: string;
    onChange: (date: string) => void;
}

type HorizonType = 'SCALP' | 'SWING' | 'INVEST' | 'CUSTOM';

const TimeHorizonSelector: React.FC<TimeHorizonSelectorProps> = ({ selectedDate, onChange }) => {
    const [activeType, setActiveType] = useState<HorizonType>('CUSTOM');

    // Helper to find next Friday from a given date
    const getNextFriday = (startDate: Date): Date => {
        const date = new Date(startDate);
        const day = date.getDay();
        const diff = 5 - day; // 5 is Friday
        if (diff <= 0) {
            date.setDate(date.getDate() + diff + 7);
        } else {
            date.setDate(date.getDate() + diff);
        }
        return date;
    };

    // Calculate dates for each horizon
    const getHorizonDate = (type: HorizonType): string => {
        const today = new Date();
        let targetDate = new Date();

        switch (type) {
            case 'SCALP':
                // Nearest Friday
                targetDate = getNextFriday(today);
                break;
            case 'SWING':
                // ~30-45 days out (let's say 35 days), then find nearest Friday
                targetDate.setDate(today.getDate() + 35);
                targetDate = getNextFriday(targetDate);
                break;
            case 'INVEST':
                // ~6 months out (180 days), then find nearest Friday
                targetDate.setDate(today.getDate() + 180);
                targetDate = getNextFriday(targetDate);
                break;
            default:
                return selectedDate;
        }
        return targetDate.toISOString().split('T')[0];
    };

    const handleSelect = (type: HorizonType) => {
        setActiveType(type);
        const newDate = getHorizonDate(type);
        onChange(newDate);
    };

    // Check if selectedDate matches one of our presets to highlight it
    useEffect(() => {
        const scalpDate = getHorizonDate('SCALP');
        const swingDate = getHorizonDate('SWING');
        const investDate = getHorizonDate('INVEST');

        if (selectedDate === scalpDate) setActiveType('SCALP');
        else if (selectedDate === swingDate) setActiveType('SWING');
        else if (selectedDate === investDate) setActiveType('INVEST');
        else setActiveType('CUSTOM');
    }, [selectedDate]);

    return (
        <div className="flex items-center gap-4 w-full">
            <div className="flex-1 flex bg-surface/50 rounded-lg p-1 border border-gray-800">
                <button
                    type="button"
                    onClick={() => handleSelect('SCALP')}
                    className={`flex-1 flex flex-col items-center justify-center py-2 px-4 rounded-md transition-all duration-200 ${activeType === 'SCALP'
                            ? 'bg-primary/20 text-primary shadow-[0_0_10px_rgba(139,92,246,0.3)] border border-primary/50'
                            : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
                        }`}
                >
                    <Zap size={18} className="mb-1" />
                    <span className="text-xs font-bold font-mono tracking-wider">SCALP</span>
                </button>

                <button
                    type="button"
                    onClick={() => handleSelect('SWING')}
                    className={`flex-1 flex flex-col items-center justify-center py-2 px-4 rounded-md transition-all duration-200 ${activeType === 'SWING'
                            ? 'bg-primary/20 text-primary shadow-[0_0_10px_rgba(139,92,246,0.3)] border border-primary/50'
                            : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
                        }`}
                >
                    <Calendar size={18} className="mb-1" />
                    <span className="text-xs font-bold font-mono tracking-wider">SWING</span>
                </button>

                <button
                    type="button"
                    onClick={() => handleSelect('INVEST')}
                    className={`flex-1 flex flex-col items-center justify-center py-2 px-4 rounded-md transition-all duration-200 ${activeType === 'INVEST'
                            ? 'bg-primary/20 text-primary shadow-[0_0_10px_rgba(139,92,246,0.3)] border border-primary/50'
                            : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
                        }`}
                >
                    <Gem size={18} className="mb-1" />
                    <span className="text-xs font-bold font-mono tracking-wider">INVEST</span>
                </button>
            </div>

            <div className="relative">
                <input
                    type="date"
                    value={selectedDate}
                    onChange={(e) => {
                        onChange(e.target.value);
                        setActiveType('CUSTOM');
                    }}
                    className="absolute inset-0 opacity-0 cursor-pointer w-full h-full z-10"
                />
                <span className="text-xs text-primary hover:text-primary/80 underline cursor-pointer font-mono whitespace-nowrap z-0">
                    Edit Date
                </span>
            </div>
        </div>
    );
};

export default TimeHorizonSelector;
