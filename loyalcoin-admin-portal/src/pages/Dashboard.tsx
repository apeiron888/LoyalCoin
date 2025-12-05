import React, { useEffect, useState } from 'react';
import { Coins, CreditCard, RefreshCw } from 'lucide-react';
import { getReserveStatus, getPendingAllocations, getPendingSettlements, ReserveStatus } from '../services/api';
import { useStore } from '../store';

export const Dashboard: React.FC = () => {
    const { isDarkMode } = useStore();
    const [reserve, setReserve] = useState<ReserveStatus | null>(null);
    const [pendingAllocations, setPendingAllocations] = useState(0);
    const [pendingSettlements, setPendingSettlements] = useState(0);
    const [loading, setLoading] = useState(true);


    const fetchData = async () => {
        setLoading(true);
        try {
            const [reserveRes, allocRes, settleRes] = await Promise.all([
                getReserveStatus(),
                getPendingAllocations(1, 0),
                getPendingSettlements(1, 0),
            ]);
            setReserve(reserveRes.data);
            setPendingAllocations(allocRes.data.total || 0);
            setPendingSettlements(settleRes.data.total || 0);
        } catch (err) {
            console.error('Failed to fetch dashboard data', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const cardBg = isDarkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200';
    const textPrimary = isDarkMode ? 'text-white' : 'text-gray-900';
    const textSecondary = isDarkMode ? 'text-slate-400' : 'text-gray-500';

    return (
        <div className="space-y-6 lg:space-y-8">
            <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
                <div>
                    <h1 className={`text-xl sm:text-2xl font-bold ${textPrimary}`}>Admin Dashboard</h1>
                    <p className={`text-sm ${textSecondary}`}>System overview and reserve status</p>
                </div>
                <button
                    onClick={fetchData}
                    disabled={loading}
                    className={`flex items-center justify-center gap-2 px-4 py-2 rounded-lg border transition-colors w-full sm:w-auto ${isDarkMode
                        ? 'bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700'
                        : 'bg-white border-gray-200 text-gray-700 hover:bg-gray-50'
                        }`}
                >
                    <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    Refresh
                </button>
            </div>

            {/* Governance Wallet */}
            <div className={`rounded-xl shadow-sm border p-4 sm:p-6 ${cardBg}`}>
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-3">
                    <h2 className={`text-base sm:text-lg font-semibold ${textPrimary}`}>Governance Wallet</h2>
                    <div className={`flex items-center gap-2 px-3 py-1 rounded-full w-fit ${reserve?.health === 'ACTIVE' ? 'bg-green-500/10 text-green-500' : 'bg-gray-500/10 text-gray-500'}`}>
                        <span className="text-sm font-medium">{reserve?.health || 'Loading'}</span>
                    </div>
                </div>

                {reserve?.governance_wallet_address ? (
                    <div className="space-y-4">
                        <div>
                            <p className={`text-sm ${textSecondary} mb-1`}>Wallet Address</p>
                            <div className={`p-3 rounded-lg font-mono text-xs sm:text-sm break-all ${isDarkMode ? 'bg-slate-900' : 'bg-gray-50 text-gray-600'}`}>
                                {reserve.governance_wallet_address}
                            </div>
                        </div>

                        <div className={`p-4 sm:p-6 rounded-xl ${isDarkMode ? 'bg-gradient-to-br from-amber-900/20 to-yellow-900/10' : 'bg-gradient-to-br from-amber-50 to-yellow-50'} border ${isDarkMode ? 'border-amber-800/30' : 'border-amber-200'}`}>
                            <p className={`text-xs ${textSecondary} uppercase tracking-wider mb-2`}>Total LCN Balance</p>
                            <p className={`text-2xl sm:text-4xl font-bold ${isDarkMode ? 'text-amber-400' : 'text-amber-600'}`}>
                                {reserve.lcn_balance.toLocaleString()} LCN
                            </p>
                        </div>
                    </div>
                ) : (
                    <p className={textSecondary}>Loading wallet info...</p>
                )}
            </div>

            {/* Quick Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 lg:gap-6">
                <div className={`rounded-xl shadow-sm border p-4 sm:p-6 ${cardBg}`}>
                    <div className="flex items-center gap-3 mb-3">
                        <div className={`p-2 sm:p-3 rounded-lg ${isDarkMode ? 'bg-blue-900/30' : 'bg-blue-50'}`}>
                            <Coins className={`h-5 w-5 sm:h-6 sm:w-6 ${isDarkMode ? 'text-blue-400' : 'text-blue-600'}`} />
                        </div>
                        <div>
                            <p className={`text-xs sm:text-sm ${textSecondary}`}>Pending Allocations</p>
                            <p className={`text-xl sm:text-2xl font-bold ${isDarkMode ? 'text-amber-400' : 'text-gray-900'}`}>{pendingAllocations}</p>
                        </div>
                    </div>
                    <a
                        href="/allocations"
                        className={`block text-xs sm:text-sm text-center py-2 px-4 rounded-lg transition-colors ${isDarkMode
                            ? 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                            }`}
                    >
                        Review Allocations
                    </a>
                </div>

                <div className={`rounded-xl shadow-sm border p-4 sm:p-6 ${cardBg}`}>
                    <div className="flex items-center gap-3 mb-3">
                        <div className={`p-2 sm:p-3 rounded-lg ${isDarkMode ? 'bg-green-900/30' : 'bg-green-50'}`}>
                            <CreditCard className={`h-5 w-5 sm:h-6 sm:w-6 ${isDarkMode ? 'text-green-400' : 'text-green-600'}`} />
                        </div>
                        <div>
                            <p className={`text-xs sm:text-sm ${textSecondary}`}>Pending Settlements</p>
                            <p className={`text-xl sm:text-2xl font-bold ${isDarkMode ? 'text-amber-400' : 'text-gray-900'}`}>{pendingSettlements}</p>
                        </div>
                    </div>
                    <a
                        href="/settlements"
                        className={`block text-xs sm:text-sm text-center py-2 px-4 rounded-lg transition-colors ${isDarkMode
                            ? 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                            }`}
                    >
                        Review Settlements
                    </a>
                </div>
            </div>
        </div>
    );
};
