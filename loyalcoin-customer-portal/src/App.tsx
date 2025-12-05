import React, { useEffect } from 'react';
import { Routes, Route, Navigate, NavLink, useLocation } from 'react-router-dom';
import { Home, QrCode, Send, Clock, User } from 'lucide-react';
import { useStore, initializeAuth } from './store';

// Pages
import { Login } from './pages/Login';
import { Signup } from './pages/Signup';
import { Dashboard } from './pages/Dashboard';
import { Receive } from './pages/Receive';
import { Spend } from './pages/Spend';
import { Transactions } from './pages/Transactions';
import { Profile } from './pages/Profile';

// Bottom Navigation Component
const BottomNav: React.FC = () => {
    const location = useLocation();

    const navItems = [
        { path: '/', icon: Home, label: 'Home' },
        { path: '/receive', icon: QrCode, label: 'Receive' },
        { path: '/spend', icon: Send, label: 'Spend' },
        { path: '/transactions', icon: Clock, label: 'History' },
        { path: '/profile', icon: User, label: 'Profile' },
    ];

    return (
        <nav className="bottom-nav">
            {navItems.map(({ path, icon: Icon, label }) => (
                <NavLink
                    key={path}
                    to={path}
                    className={`nav-item ${location.pathname === path ? 'active' : ''}`}
                >
                    <Icon />
                    <span>{label}</span>
                </NavLink>
            ))}
        </nav>
    );
};

// Protected Route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { isAuthenticated } = useStore();

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }

    return (
        <div className="app-container">
            <main className="main-content">
                {children}
            </main>
            <BottomNav />
        </div>
    );
};

// Main App Component
const App: React.FC = () => {
    const { isAuthenticated, fetchBalance, fetchTransactions } = useStore();

    // Initialize auth from localStorage
    useEffect(() => {
        initializeAuth();
    }, []);

    // Fetch data when authenticated
    useEffect(() => {
        if (isAuthenticated) {
            fetchBalance();
            fetchTransactions();
        }
    }, [isAuthenticated, fetchBalance, fetchTransactions]);

    return (
        <Routes>
            {/* Public routes */}
            <Route path="/login" element={
                isAuthenticated ? <Navigate to="/" replace /> : <Login />
            } />
            <Route path="/signup" element={
                isAuthenticated ? <Navigate to="/" replace /> : <Signup />
            } />

            {/* Protected routes */}
            <Route path="/" element={
                <ProtectedRoute><Dashboard /></ProtectedRoute>
            } />
            <Route path="/receive" element={
                <ProtectedRoute><Receive /></ProtectedRoute>
            } />
            <Route path="/spend" element={
                <ProtectedRoute><Spend /></ProtectedRoute>
            } />
            <Route path="/transactions" element={
                <ProtectedRoute><Transactions /></ProtectedRoute>
            } />
            <Route path="/profile" element={
                <ProtectedRoute><Profile /></ProtectedRoute>
            } />

            {/* Fallback */}
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
};

export default App;
