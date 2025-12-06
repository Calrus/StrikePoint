import React, { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import { Menu, Calculator, Calendar, Radar, Twitter, X, FileText, Trash2 } from 'lucide-react';

const SavedDraftsList = () => {
    const [drafts, setDrafts] = useState<any[]>([]);

    const loadDrafts = () => {
        try {
            const saved = localStorage.getItem('saved_drafts');
            if (saved) {
                setDrafts(JSON.parse(saved));
            }
        } catch (e) {
            console.error(e);
        }
    };

    useEffect(() => {
        loadDrafts();
        window.addEventListener('drafts-updated', loadDrafts);
        return () => window.removeEventListener('drafts-updated', loadDrafts);
    }, []);

    const deleteDraft = (e: React.MouseEvent, index: number) => {
        e.preventDefault();
        e.stopPropagation();
        const newDrafts = [...drafts];
        newDrafts.splice(index, 1);
        setDrafts(newDrafts);
        localStorage.setItem('saved_drafts', JSON.stringify(newDrafts));
    };

    if (drafts.length === 0) return <div className="text-gray-600 text-xs px-2 italic">No drafts saved</div>;

    return (
        <>
            {drafts.map((draft, idx) => (
                <NavLink
                    key={idx}
                    to="/builder"
                    state={{ trade: draft }}
                    className="flex items-center justify-between px-3 py-2 rounded-md text-sm transition-colors group text-gray-400 hover:text-white hover:bg-white/5"
                >
                    <div className="flex items-center gap-2 overflow-hidden">
                        <FileText size={14} />
                        <span className="truncate max-w-[120px] font-mono text-xs">{draft.name || 'Untitled'}</span>
                    </div>
                    <button
                        onClick={(e) => deleteDraft(e, idx)}
                        className="opacity-0 group-hover:opacity-100 hover:text-red-400 transition-opacity p-1"
                    >
                        <Trash2 size={12} />
                    </button>
                </NavLink>
            ))}
        </>
    );
};

interface SidebarProps {
    isOpen: boolean;
    toggle: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen, toggle }) => {
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
                onClick={toggle}
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

                    <div className="mt-6 mb-2 px-4">
                        <h3 className="text-xs font-bold text-gray-500 uppercase tracking-widest mb-3">Saved Drafts</h3>
                        <div className="space-y-1">
                            <SavedDraftsList />
                        </div>
                    </div>

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
