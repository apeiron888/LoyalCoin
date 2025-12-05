import React, { useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Coins, CreditCard, LogOut, Shield, Sun, Moon, Menu, X } from 'lucide-react';
import { useStore } from '../store';

const navItems = [
    { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
    { to: '/allocations', icon: Coins, label: 'Allocations' },
    { to: '/settlements', icon: CreditCard, label: 'Settlements' },
];

export const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { user, logout, isDarkMode, toggleTheme } = useStore();
    const navigate = useNavigate();
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const closeMobileMenu = () => {
        setIsMobileMenuOpen(false);
    };

    return (
        <div className={`flex min-h-screen ${isDarkMode ? 'bg-slate-900' : 'bg-gray-100'}`}>
            {/* Mobile Header */}
            <div className={`lg:hidden fixed top-0 left-0 right-0 z-40 ${isDarkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'} border-b`}>
                <div className="flex items-center justify-between p-4">
                    <div className="flex items-center gap-3">
                        <div className="h-8 w-8 bg-gradient-to-br from-amber-400 to-yellow-600 rounded-lg flex items-center justify-center shadow-lg">
                            <Shield className="h-5 w-5 text-white" />
                        </div>
                        <div>
                            <h1 className={`font-bold text-base ${isDarkMode ? 'text-white' : 'text-gray-900'}`}>LoyalCoin</h1>
                            <p className="text-xs text-amber-500 font-medium">Admin</p>
                        </div>
                    </div>
                    <button
                        onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                        className={`p-2 rounded-lg ${isDarkMode ? 'hover:bg-slate-700 text-white' : 'hover:bg-gray-100 text-gray-900'}`}
                    >
                        {isMobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
                    </button>
                </div>
            </div>

            {/* Mobile Menu Overlay */}
            {isMobileMenuOpen && (
                <div
                    className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-40 pt-[73px]"
                    onClick={closeMobileMenu}
                />
            )}

            {/* Sidebar */}
            <aside className={`
                fixed lg:static inset-y-0 left-0 z-50 w-64
                ${isDarkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'} 
                border-r flex flex-col transform transition-transform duration-300 ease-in-out
                ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
            `}>
                {/* Logo - shown on both mobile and desktop */}
                <div className={`p-6 border-b ${isDarkMode ? 'border-slate-700' : 'border-gray-200'}`}>
                    <div className="flex items-center gap-3">
                        <div className="h-10 w-10 bg-gradient-to-br from-amber-400 to-yellow-600 rounded-lg flex items-center justify-center shadow-lg">
                            <Shield className="h-6 w-6 text-white" />
                        </div>
                        <div>
                            <h1 className={`font-bold text-lg ${isDarkMode ? 'text-white' : 'text-gray-900'}`}>LoyalCoin</h1>
                            <p className="text-xs text-amber-500 font-medium">Admin Portal</p>
                        </div>
                    </div>
                </div>

                <nav className="flex-1 p-4 overflow-y-auto">
                    <ul className="space-y-2">
                        {navItems.map((item) => (
                            <li key={item.to}>
                                <NavLink
                                    to={item.to}
                                    onClick={closeMobileMenu}
                                    className={({ isActive }) =>
                                        `flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${isActive
                                            ? 'bg-gradient-to-r from-amber-500 to-yellow-500 text-white shadow-md'
                                            : isDarkMode
                                                ? 'text-slate-300 hover:bg-slate-700 hover:text-white'
                                                : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                                        }`
                                    }
                                >
                                    <item.icon className="h-5 w-5" />
                                    {item.label}
                                </NavLink>
                            </li>
                        ))}
                    </ul>
                </nav>

                <div className={`p-4 border-t ${isDarkMode ? 'border-slate-700' : 'border-gray-200'}`}>
                    {/* Theme Toggle */}
                    <button
                        onClick={toggleTheme}
                        className={`flex items-center gap-3 w-full px-4 py-2 mb-3 rounded-lg transition-colors ${isDarkMode ? 'text-slate-300 hover:bg-slate-700' : 'text-gray-600 hover:bg-gray-100'
                            }`}
                    >
                        {isDarkMode ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
                        {isDarkMode ? 'Light Mode' : 'Dark Mode'}
                    </button>

                    <div className={`flex items-center gap-3 px-2 mb-4 ${isDarkMode ? 'text-white' : 'text-gray-900'}`}>
                        <div className="h-8 w-8 bg-gradient-to-br from-amber-400 to-yellow-600 rounded-full flex items-center justify-center text-sm font-medium text-white">
                            {user?.email?.[0]?.toUpperCase()}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate">{user?.email}</p>
                            <p className={`text-xs ${isDarkMode ? 'text-slate-400' : 'text-gray-500'}`}>Administrator</p>
                        </div>
                    </div>
                    <button
                        onClick={handleLogout}
                        className={`flex items-center gap-2 w-full px-4 py-2 rounded-lg transition-colors ${isDarkMode ? 'text-red-400 hover:bg-red-900/30' : 'text-red-600 hover:bg-red-50'
                            }`}
                    >
                        <LogOut className="h-4 w-4" />
                        Sign Out
                    </button>
                </div>
            </aside>

            {/* Main Content */}
            <main className={`flex-1 overflow-auto ${isDarkMode ? 'text-white' : 'text-gray-900'} lg:p-8 p-4 pt-[89px] lg:pt-8`}>
                {children}
            </main>
        </div>
    );
};
