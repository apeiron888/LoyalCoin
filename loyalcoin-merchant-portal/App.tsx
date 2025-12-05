import React from 'react';
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from './store';
import { Layout } from './components/Layout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { IssueLCN } from './pages/IssueLCN';
import { Receive } from './pages/Receive';
import { Transactions } from './pages/Transactions';
import { BuyLCN } from './pages/BuyLCN';
import { CashOut } from './pages/CashOut';
import { Settings } from './pages/Settings';

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
        <Route path="/login" element={
          isAuthenticated ? <Navigate to="/" replace /> : <Login />
        } />

        <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
        <Route path="/issue" element={<ProtectedRoute><IssueLCN /></ProtectedRoute>} />
        <Route path="/receive" element={<ProtectedRoute><Receive /></ProtectedRoute>} />
        <Route path="/transactions" element={<ProtectedRoute><Transactions /></ProtectedRoute>} />
        <Route path="/buy" element={<ProtectedRoute><BuyLCN /></ProtectedRoute>} />
        <Route path="/cash-out" element={<ProtectedRoute><CashOut /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
      </Routes>
    </HashRouter>
  );
};

export default App;