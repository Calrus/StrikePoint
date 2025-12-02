import React, { useRef, useEffect } from 'react';

interface DateSliderProps {
    selectedDate: string;
    onChange: (date: string) => void;
}

const DateSlider: React.FC<DateSliderProps> = ({ selectedDate, onChange }) => {
    const scrollRef = useRef<HTMLDivElement>(null);

    // Generate dates for the next 12 months (Fridays only)
    const generateDates = () => {
        const dates = [];
        const today = new Date();
        const currentYear = today.getFullYear();

        // Find next Friday
        let d = new Date();
        d.setDate(d.getDate() + (5 - d.getDay() + 7) % 7);
        if (d <= today) d.setDate(d.getDate() + 7);

        // Generate for next 50 weeks
        for (let i = 0; i < 50; i++) {
            dates.push(new Date(d));
            d.setDate(d.getDate() + 7);
        }
        return dates;
    };

    const dates = generateDates();

    // Group by Month
    const groupedDates: { [key: string]: Date[] } = {};
    dates.forEach(date => {
        const monthYear = date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
        if (!groupedDates[monthYear]) {
            groupedDates[monthYear] = [];
        }
        groupedDates[monthYear].push(date);
    });

    return (
        <div className="w-full bg-surface/30 border border-gray-800 rounded-lg p-2 overflow-hidden relative">
            <div
                ref={scrollRef}
                className="flex overflow-x-auto custom-scrollbar pb-2 gap-8 px-4"
                style={{ scrollBehavior: 'smooth' }}
            >
                {Object.entries(groupedDates).map(([month, monthDates]) => (
                    <div key={month} className="flex flex-col items-center flex-shrink-0">
                        <span className="text-xs font-mono text-gray-500 mb-2 uppercase tracking-wider sticky left-0">
                            {month}
                        </span>
                        <div className="flex gap-2">
                            {monthDates.map(date => {
                                const dateStr = date.toISOString().split('T')[0];
                                const isSelected = selectedDate === dateStr;
                                const day = date.getDate();

                                return (
                                    <button
                                        key={dateStr}
                                        type="button"
                                        onClick={() => onChange(dateStr)}
                                        className={`
                                            flex flex-col items-center justify-center w-8 h-8 rounded-md text-xs font-mono font-bold transition-all
                                            ${isSelected
                                                ? 'bg-primary text-white shadow-[0_0_10px_rgba(139,92,246,0.5)] scale-110'
                                                : 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-white'
                                            }
                                        `}
                                    >
                                        {day}
                                    </button>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </div>

            {/* Fade gradients for scroll indication */}
            <div className="absolute left-0 top-0 bottom-0 w-8 bg-gradient-to-r from-black/50 to-transparent pointer-events-none" />
            <div className="absolute right-0 top-0 bottom-0 w-8 bg-gradient-to-l from-black/50 to-transparent pointer-events-none" />
        </div>
    );
};

export default DateSlider;
