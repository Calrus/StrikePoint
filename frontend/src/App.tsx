import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import Sidebar from './components/Sidebar';
import Calculator from './pages/Calculator';
import NewsInterpreter from './pages/NewsInterpreter';
import Earnings from './pages/Earnings';
import Whales from './pages/Whales';
import Builder from './pages/Builder';

// Unified Layout with "Push" Sidebar
const MainLayout = ({ children, isBuilder = false }: { children: React.ReactNode, isBuilder?: boolean }) => {
    const [sidebarOpen, setSidebarOpen] = useState(true);

    return (
        <div className="min-h-screen bg-background text-white font-sans selection:bg-primary selection:text-white flex overflow-hidden">
            <Sidebar isOpen={sidebarOpen} toggle={() => setSidebarOpen(!sidebarOpen)} />

            {/* Main Content Area - Pushed by Sidebar */}
            <main
                className={`flex-1 flex flex-col transition-all duration-300 ease-in-out 
                ${sidebarOpen ? 'ml-64' : 'ml-0'} 
                ${isBuilder ? 'h-screen overflow-hidden' : 'min-h-screen overflow-y-auto'}`}
            >
                {isBuilder ? (
                    children
                ) : (
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 pt-20 w-full">
                        {children}
                    </div>
                )}
            </main>
        </div>
    );
};

function App() {
    return (
        <Router>
            <Routes>
                <Route path="/" element={<MainLayout><Calculator /></MainLayout>} />
                <Route path="/news" element={<MainLayout><NewsInterpreter /></MainLayout>} />
                <Route path="/earnings" element={<MainLayout><Earnings /></MainLayout>} />
                <Route path="/whales" element={<MainLayout><Whales /></MainLayout>} />
                <Route path="/builder" element={<MainLayout isBuilder={true}><Builder /></MainLayout>} />
            </Routes>
        </Router>
    );
}

export default App;
