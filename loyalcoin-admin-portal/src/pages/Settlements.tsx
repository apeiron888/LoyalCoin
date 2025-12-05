import React, { useEffect, useState } from 'react';
import { CheckCircle, XCircle, RefreshCw, Clock } from 'lucide-react';
import { getPendingSettlements, approveSettlement, Settlement } from '../services/api';
import { useStore } from '../store';

export const Settlements: React.FC = () => {
    const { isDarkMode } = useStore();
    const [settlements, setSettlements] = useState<Settlement[]>([]);
    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState<string | null>(null);
    const [actionError, setActionError] = useState<string | null>(null);

    const cardBg = isDarkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200';
    const textPrimary = isDarkMode ? 'text-white' : 'text-gray-900';
    const textSecondary = isDarkMode ? 'text-slate-400' : 'text-gray-500';
    const tableBg = isDarkMode ? 'bg-slate-700/50' : 'bg-gray-50';
    const rowHover = isDarkMode ? 'hover:bg-slate-700/50' : 'hover:bg-gray-50';

    const fetchSettlements = async () => {
        setLoading(true);
        setActionError(null);
        try {
            const res = await getPendingSettlements(50, 0);
            console.log('Fetched settlements:', res.data);
            setSettlements(res.data.settlements || []);
        } catch (err) {
            console.error('Failed to fetch settlements', err);
            setActionError('Failed to load settlements');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSettlements();
    }, []);

    const handleAction = async (id: string, action: 'APPROVE' | 'REJECT') => {
        console.log('Processing action:', action, 'for settlement ID:', id);
        setProcessing(id);
        setActionError(null);
        try {
            const paymentRef = action === 'APPROVE' ? `BANK-${Date.now()}` : undefined;
            const result = await approveSettlement(id, action, paymentRef, action === 'REJECT' ? 'Rejected by admin' : 'Approved by admin');
            console.log('Action result:', result);
            setSettlements((prev) => prev.filter((s) => s.id !== id));
        } catch (err: any) {
            console.error('Failed to process settlement', err);
            setActionError(err?.message || 'Failed to process settlement');
        } finally {
            setProcessing(null);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className={`text-2xl font-bold ${textPrimary}`}>Pending Settlements</h1>
                    <p className={textSecondary}>Review and approve merchant cash-out requests</p>
                </div>
                <button
                    onClick={fetchSettlements}
                    disabled={loading}
                    className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors ${isDarkMode
                            ? 'bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700'
                            : 'bg-white border-gray-200 text-gray-700 hover:bg-gray-50'
                        }`}
                >
                    <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    Refresh
                </button>
            </div>

            {actionError && (
                <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
                    {actionError}
                </div>
            )}

            <div className={`rounded-xl shadow-sm border overflow-hidden ${cardBg}`}>
                {loading ? (
                    <div className={`p-12 text-center ${textSecondary}`}>Loading settlements...</div>
                ) : settlements.length === 0 ? (
                    <div className={`p-12 text-center ${textSecondary}`}>
                        <Clock className={`h-12 w-12 mx-auto mb-3 ${isDarkMode ? 'text-slate-600' : 'text-gray-300'}`} />
                        <p>No pending settlements</p>
                    </div>
                ) : (
                    <table className="w-full text-sm">
                        <thead className={`${tableBg} uppercase tracking-wider text-xs ${textSecondary}`}>
                            <tr>
                                <th className="px-6 py-3 text-left">Date</th>
                                <th className="px-6 py-3 text-left">Merchant ID</th>
                                <th className="px-6 py-3 text-left">Amount</th>
                                <th className="px-6 py-3 text-left">Bank Account</th>
                                <th className="px-6 py-3 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className={`divide-y ${isDarkMode ? 'divide-slate-700' : 'divide-gray-100'}`}>
                            {settlements.map((settle) => (
                                <tr key={settle.id} className={rowHover}>
                                    <td className={`px-6 py-4 ${textPrimary}`}>
                                        {new Date(settle.requested_at).toLocaleDateString()}
                                    </td>
                                    <td className={`px-6 py-4 font-mono text-xs ${textSecondary}`}>
                                        {settle.merchant_id?.substring(0, 12)}...
                                    </td>
                                    <td className="px-6 py-4">
                                        <span className={`font-semibold ${textPrimary}`}>{settle.amount_lcn?.toLocaleString()} LCN</span>
                                        <span className={`ml-2 ${textSecondary}`}>({settle.amount_etb?.toLocaleString()} ETB)</span>
                                    </td>
                                    <td className={`px-6 py-4 ${textSecondary}`}>
                                        <p>{settle.bank_account?.bank_name}</p>
                                        <p className="text-xs">{settle.bank_account?.account_number} • {settle.bank_account?.account_holder}</p>
                                    </td>
                                    <td className="px-6 py-4 text-right">
                                        <div className="flex items-center justify-end gap-2">
                                            <button
                                                onClick={() => handleAction(settle.id, 'APPROVE')}
                                                disabled={processing === settle.id}
                                                className="flex items-center gap-1 px-3 py-1.5 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 disabled:opacity-50 text-white text-xs font-medium rounded-lg shadow-sm"
                                            >
                                                <CheckCircle className="h-3.5 w-3.5" />
                                                Approve
                                            </button>
                                            <button
                                                onClick={() => handleAction(settle.id, 'REJECT')}
                                                disabled={processing === settle.id}
                                                className="flex items-center gap-1 px-3 py-1.5 bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 disabled:opacity-50 text-white text-xs font-medium rounded-lg shadow-sm"
                                            >
                                                <XCircle className="h-3.5 w-3.5" />
                                                Reject
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    );
};
