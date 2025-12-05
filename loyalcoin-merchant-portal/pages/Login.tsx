import React, { useState } from 'react';
import { useStore } from '../store';
import { Coins, AlertCircle } from 'lucide-react';

export const Login: React.FC = () => {
    const { login, signup, isLoading, error, clearError } = useStore();
    const [isLogin, setIsLogin] = useState(true);
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [businessName, setBusinessName] = useState('');
    const [localError, setLocalError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLocalError(null);
        clearError();

        try {
            if (isLogin) {
                await login(email, password);
            } else {
                if (!businessName.trim()) {
                    setLocalError('Business name is required');
                    return;
                }
                await signup(email, password, businessName);
            }
        } catch (err: any) {
            setLocalError(err.message || 'An error occurred');
        }
    };

    const displayError = localError || error;

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex flex-col items-center justify-center p-4">
            <div className="mb-8 text-center">
                <div className="mx-auto h-16 w-16 bg-gradient-to-br from-amber-400 to-yellow-600 rounded-2xl flex items-center justify-center mb-4 shadow-xl shadow-amber-500/20 transform rotate-3">
                    <Coins className="text-white h-8 w-8" />
                </div>
                <h1 className="text-3xl font-bold text-gray-900">LoyalCoin</h1>
                <p className="text-amber-500 mt-2 font-medium">Merchant Portal</p>
            </div>

            <div className="w-full max-w-md bg-white rounded-xl p-8 shadow-xl border-t-4 border-t-amber-500">
                <div className="flex border-b border-gray-100 mb-6">
                    <button
                        className={`flex-1 pb-3 text-sm font-medium transition-colors ${isLogin ? 'text-amber-500 border-b-2 border-amber-500' : 'text-gray-400'}`}
                        onClick={() => setIsLogin(true)}
                    >
                        Sign In
                    </button>
                    <button
                        className={`flex-1 pb-3 text-sm font-medium transition-colors ${!isLogin ? 'text-amber-500 border-b-2 border-amber-500' : 'text-gray-400'}`}
                        onClick={() => setIsLogin(false)}
                    >
                        Sign Up
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5">
                    {displayError && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3 flex items-start gap-2">
                            <AlertCircle className="h-4 w-4 text-red-600 mt-0.5 flex-shrink-0" />
                            <span className="text-sm text-red-700">{displayError}</span>
                        </div>
                    )}

                    {!isLogin && (
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1.5">Business Name</label>
                            <input
                                type="text"
                                placeholder="Enter your business name"
                                value={businessName}
                                onChange={(e) => setBusinessName(e.target.value)}
                                required={!isLogin}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-amber-500"
                            />
                        </div>
                    )}

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Email</label>
                        <input
                            type="email"
                            placeholder="you@example.com"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-amber-500"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Password</label>
                        <input
                            type="password"
                            placeholder="••••••••"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-amber-500"
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full py-3 bg-gradient-to-r from-amber-600 to-yellow-500 text-white font-medium rounded-lg hover:from-amber-700 hover:to-yellow-500 disabled:opacity-50 transition-all"
                    >
                        {isLoading ? 'Loading...' : (isLogin ? 'Sign In' : 'Create Account')}
                    </button>
                </form>
            </div>
        </div>
    );
};
