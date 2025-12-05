import React from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from './store';
import { Layout } from './components/Layout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Allocations } from './pages/Allocations';
import { Settlements } from './pages/Settlements';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const isAuthenticated = useStore((state) => state.isAuthenticated);
    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }
    return <Layout>{children}</Layout>;
};

const App: React.FC = () => {
    const isAuthenticated = useStore((state) => state.isAuthenticated);

    return (
        <HashRouter>
            <Routes>
                <Route path="/login" element={isAuthenticated ? <Navigate to="/" replace /> : <Login />} />
                <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                <Route path="/allocations" element={<ProtectedRoute><Allocations /></ProtectedRoute>} />
                <Route path="/settlements" element={<ProtectedRoute><Settlements /></ProtectedRoute>} />
            </Routes>
        </HashRouter>
    );
};

export default App;
