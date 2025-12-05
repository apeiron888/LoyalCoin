import React, { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import {
    LayoutDashboard,
    Send,
    QrCode,
    History,
    Coins,
    ArrowDownLeft,
    Settings,
    LogOut,
    Menu,
    X
} from 'lucide-react';
import { useStore } from '../store';

interface NavItem {
    path: string;
    icon: React.ComponentType<{ className?: string }>;
    label: string;
}

const navItems: NavItem[] = [
    { path: '/', icon: LayoutDashboard, label: 'Dashboard' },
    { path: '/issue', icon: Send, label: 'Issue LCN' },
    { path: '/receive', icon: QrCode, label: 'Receive Coins' },
    { path: '/transactions', icon: History, label: 'Transactions' },
    { path: '/buy', icon: Coins, label: 'Buy LCN' },
    { path: '/cash-out', icon: ArrowDownLeft, label: 'Cash Out' },
    { path: '/settings', icon: Settings, label: 'Settings' },
];

export const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { user, wallet, logout } = useStore();
    const location = useLocation();
    const navigate = useNavigate();
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Mobile Header */}
            <div className="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white border-b border-gray-200">
                <div className="flex items-center justify-between px-4 h-16">
                    <button
                        onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                        className="p-2 rounded-lg hover:bg-gray-100"
                    >
                        {isSidebarOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
                    </button>
                    <div className="flex items-center gap-2">
                        <Coins className="h-5 w-5 text-amber-600" />
                        <span className="font-semibold text-gray-900">LoyalCoin</span>
                    </div>
                    <div className="w-10" />
                </div>
            </div>

            {/* Overlay */}
            {isSidebarOpen && (
                <div
                    className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-30"
                    onClick={() => setIsSidebarOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside className={`
        fixed inset-y-0 left-0 z-40 w-64 bg-white border-r border-gray-200
        transform transition-transform duration-300 ease-in-out
        ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
        lg:translate-x-0
      `}>
                {/* Logo */}
                <div className="p-6 border-b border-gray-200">
                    <div className="flex items-center gap-3">
                        <div className="h-10 w-10 bg-gradient-to-br from-amber-500 to-yellow-600 rounded-lg flex items-center justify-center">
                            <Coins className="h-6 w-6 text-white" />
                        </div>
                        <div>
                            <h1 className="font-bold text-lg text-gray-900">LoyalCoin</h1>
                            <p className="text-xs text-amber-600 font-medium">Merchant Portal</p>
                        </div>
                    </div>
                </div>

                {/* Navigation */}
                <nav className="flex-1 p-4 overflow-y-auto">
                    <ul className="space-y-1">
                        {navItems.map((item) => {
                            const isActive = location.pathname === item.path;
                            return (
                                <li key={item.path}>
                                    <Link
                                        to={item.path}
                                        onClick={() => setIsSidebarOpen(false)}
                                        className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${isActive
                                            ? 'bg-gradient-to-r from-amber-500 to-yellow-500 text-white shadow-md'
                                            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                                            }`}
                                    >
                                        <item.icon className="h-5 w-5" />
                                        <span className="font-medium">{item.label}</span>
                                    </Link>
                                </li>
                            );
                        })}
                    </ul>
                </nav>

                {/* User Section */}
                <div className="p-4 border-t border-gray-200">
                    <div className="flex items-center gap-3 px-2 mb-4">
                        <div className="h-10 w-10 bg-gradient-to-br from-amber-500 to-yellow-600 rounded-full flex items-center justify-center text-sm font-bold text-white">
                            {user?.businessName?.[0]?.toUpperCase() || 'M'}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">{user?.businessName}</p>
                            <p className="text-xs text-gray-500">Merchant</p>
                        </div>
                    </div>
                    <button
                        onClick={handleLogout}
                        className="flex items-center gap-2 w-full px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    >
                        <LogOut className="h-4 w-4" />
                        <span className="text-sm font-medium">Sign Out</span>
                    </button>
                </div>
            </aside>

            {/* Main Content */}
            <main className="lg:ml-64 pt-16 lg:pt-0">
                {/* Desktop Header - Shows LCN Balance only */}
                <header className="hidden lg:flex h-16 items-center justify-between border-b border-gray-200 bg-white px-8">
                    <div className="flex items-center gap-2">
                        <Coins className="h-5 w-5 text-amber-600" />
                        <span className="text-sm font-medium text-gray-600">Balance:</span>
                        <span className="text-lg font-bold text-amber-600">{wallet.balanceLCN.toLocaleString()} LCN</span>
                    </div>
                </header>

                {/* Page Content */}
                <div className="p-4 lg:p-8">
                    {children}
                </div>
            </main>
        </div>
    );
};
