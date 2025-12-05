import React, { useState } from 'react';
import { useStore } from '../store';
import { Shield, AlertCircle } from 'lucide-react';

export const Login: React.FC = () => {
    const { login, isLoading, error, clearError, isDarkMode } = useStore();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        clearError();
        try {
            await login(email, password);
        } catch {
            // Error is handled in store
        }
    };

    return (
        <div className={`min-h-screen ${isDarkMode ? 'bg-slate-900' : 'bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900'} flex flex-col items-center justify-center p-4`}>
            <div className="mb-8 text-center">
                <div className="mx-auto h-16 w-16 bg-gradient-to-br from-amber-400 to-yellow-600 rounded-2xl flex items-center justify-center mb-4 shadow-xl shadow-amber-500/20">
                    <Shield className="text-white h-8 w-8" />
                </div>
                <h1 className="text-3xl font-bold text-white">LoyalCoin Admin</h1>
                <p className="text-amber-400 mt-2 font-medium">System Administration Portal</p>
            </div>

            <div className="w-full max-w-md bg-slate-800/80 backdrop-blur rounded-xl p-8 shadow-2xl border border-slate-700">
                <h2 className="text-xl font-semibold text-white mb-6">Sign In</h2>

                {error && (
                    <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center gap-2 text-red-400 text-sm">
                        <AlertCircle className="h-4 w-4 flex-shrink-0" />
                        <span>{error}</span>
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-slate-300 mb-1">Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-transparent"
                            placeholder="admin@loyalcoin.com"
                            required
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-300 mb-1">Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 bg-slate-700 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-transparent"
                            placeholder="••••••••"
                            required
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full py-3 px-4 bg-gradient-to-r from-amber-500 to-yellow-500 hover:from-amber-600 hover:to-yellow-600 disabled:opacity-50 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-all shadow-lg shadow-amber-500/25 flex items-center justify-center gap-2"
                    >
                        {isLoading ? (
                            <>
                                <div className="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                Signing in...
                            </>
                        ) : (
                            'Sign In'
                        )}
                    </button>
                </form>

                <p className="mt-6 text-center text-xs text-slate-500">
                    Access restricted to authorized administrators only.
                </p>
            </div>
        </div>
    );
};
