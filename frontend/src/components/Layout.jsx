import React from 'react';

const Layout = ({ children }) => {
    return (
        <div className="min-h-screen bg-background text-white font-sans selection:bg-primary selection:text-white">
            {/* Header */}
            <header className="border-b border-surface/50 backdrop-blur-md sticky top-0 z-50 bg-background/80">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-primary rounded-full animate-pulse" />
                        <h1 className="text-xl font-bold tracking-tight font-mono">
                            STRIKE<span className="text-primary">LOGIC</span>
                        </h1>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-gray-400 font-mono">
                        <span>STATUS: <span className="text-success">ONLINE</span></span>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {children}
            </main>
        </div>
    );
};

export default Layout;
