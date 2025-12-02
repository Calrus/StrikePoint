import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Sidebar from './components/Sidebar';
import Calculator from './pages/Calculator';
import NewsInterpreter from './pages/NewsInterpreter';
import Earnings from './pages/Earnings';
import Whales from './pages/Whales';

function App() {
    return (
        <Router>
            <div className="min-h-screen bg-background text-white font-sans selection:bg-primary selection:text-white">
                <Sidebar />
                <main className="min-h-screen w-full">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 pt-20">
                        <Routes>
                            <Route path="/" element={<Calculator />} />
                            <Route path="/news" element={<NewsInterpreter />} />
                            <Route path="/earnings" element={<Earnings />} />
                            <Route path="/whales" element={<Whales />} />
                        </Routes>
                    </div>
                </main>
            </div>
        </Router>
    );
}

export default App;
