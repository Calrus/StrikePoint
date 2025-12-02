import React, { useState } from 'react';
import { NavLink } from 'react-router-dom';
import { Menu, Calculator, Calendar, Radar, Twitter, X } from 'lucide-react';

const Sidebar = () => {
    const [isOpen, setIsOpen] = useState(true);

    const toggleSidebar = () => setIsOpen(!isOpen);

    const navItems = [
        { name: 'Calculator', icon: Calculator, path: '/' },
        { name: 'News Interpreter', icon: Twitter, path: '/news' },
        { name: 'Earnings Plays', icon: Calendar, path: '/earnings' },
        { name: 'Whale Scanner', icon: Radar, path: '/whales' },
    ];

    return (
        <>
            {/* Toggle Button */}
            <button
                onClick={toggleSidebar}
                className="fixed top-4 left-4 z-50 p-2 rounded-lg bg-black/90 text-primary border border-primary/20 hover:bg-primary/10 transition-colors backdrop-blur-md shadow-[0_0_10px_rgba(0,0,0,0.5)]"
                aria-label="Toggle Menu"
            >
                {isOpen ? <X size={24} /> : <Menu size={24} />}
            </button>

            {/* Sidebar */}
            <div
                className={`fixed top-0 left-0 h-full bg-black/90 backdrop-blur-xl border-r border-primary/20 transition-all duration-300 z-40 ${isOpen ? 'w-64 translate-x-0' : 'w-64 -translate-x-full'
                    } shadow-[0_0_30px_rgba(0,255,0,0.05)]`}
            >
                <div className="flex flex-col h-full p-4 pt-20">
                    <div className="mb-8 px-2">
                        <h1 className="text-xl font-bold tracking-tight font-mono text-white whitespace-nowrap">
                            STRIKE<span className="text-primary">LOGIC</span>
                        </h1>
                    </div>

                    <nav className="flex-1 space-y-2">
                        {navItems.map((item) => (
                            <NavLink
                                key={item.path}
                                to={item.path}
                                className={({ isActive }) =>
                                    `flex items-center gap-3 px-4 py-3 rounded-lg transition-all duration-200 group whitespace-nowrap ${isActive
                                        ? 'bg-primary/20 text-primary border border-primary/30 shadow-[0_0_15px_rgba(0,255,0,0.15)]'
                                        : 'text-gray-400 hover:text-white hover:bg-white/5'
                                    }`
                                }
                            >
                                <item.icon size={20} className="min-w-[20px]" />
                                <span className="font-mono text-sm tracking-wide">{item.name}</span>
                            </NavLink>
                        ))}
                    </nav>

                    <div className="mt-auto px-4 py-4 border-t border-white/10">
                        <div className="flex items-center gap-2 text-xs text-gray-500 font-mono">
                            <div className="w-2 h-2 bg-success rounded-full animate-pulse" />
                            <span>SYSTEM ONLINE</span>
                        </div>
                    </div>
                </div>
            </div>
        </>
    );
};

export default Sidebar;
