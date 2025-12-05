import React, { useEffect, useState } from 'react';
import { CheckCircle, XCircle, RefreshCw, Clock } from 'lucide-react';
import { getPendingAllocations, approveAllocation, Allocation } from '../services/api';
import { useStore } from '../store';

export const Allocations: React.FC = () => {
    const { isDarkMode } = useStore();
    const [allocations, setAllocations] = useState<Allocation[]>([]);
    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState<string | null>(null);
    const [actionError, setActionError] = useState<string | null>(null);

    const cardBg = isDarkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200';
    const textPrimary = isDarkMode ? 'text-white' : 'text-gray-900';
    const textSecondary = isDarkMode ? 'text-slate-400' : 'text-gray-500';
    const tableBg = isDarkMode ? 'bg-slate-700/50' : 'bg-gray-50';
    const rowHover = isDarkMode ? 'hover:bg-slate-700/50' : 'hover:bg-gray-50';

    const fetchAllocations = async () => {
        setLoading(true);
        setActionError(null);
        try {
            const res = await getPendingAllocations(50, 0);
            console.log('Fetched allocations:', res.data);
            setAllocations(res.data.allocations || []);
        } catch (err) {
            console.error('Failed to fetch allocations', err);
            setActionError('Failed to load allocations');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchAllocations();
    }, []);

    const handleAction = async (id: string, action: 'APPROVE' | 'REJECT') => {
        console.log('Processing action:', action, 'for allocation ID:', id);
        setProcessing(id);
        setActionError(null);
        try {
            const result = await approveAllocation(id, action, action === 'REJECT' ? 'Rejected by admin' : 'Approved by admin');
            console.log('Action result:', result);
            // Remove from list on success
            setAllocations((prev) => prev.filter((a) => a.id !== id));
        } catch (err: any) {
            console.error('Failed to process allocation', err);
            setActionError(err?.message || 'Failed to process allocation');
        } finally {
            setProcessing(null);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className={`text-2xl font-bold ${textPrimary}`}>Pending Allocations</h1>
                    <p className={textSecondary}>Review and approve merchant LCN purchase requests</p>
                </div>
                <button
                    onClick={fetchAllocations}
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
                    <div className={`p-12 text-center ${textSecondary}`}>Loading allocations...</div>
                ) : allocations.length === 0 ? (
                    <div className={`p-12 text-center ${textSecondary}`}>
                        <Clock className={`h-12 w-12 mx-auto mb-3 ${isDarkMode ? 'text-slate-600' : 'text-gray-300'}`} />
                        <p>No pending allocations</p>
                    </div>
                ) : (
                    <table className="w-full text-sm">
                        <thead className={`${tableBg} uppercase tracking-wider text-xs ${textSecondary}`}>
                            <tr>
                                <th className="px-6 py-3 text-left">Date</th>
                                <th className="px-6 py-3 text-left">Merchant ID</th>
                                <th className="px-6 py-3 text-left">Amount</th>
                                <th className="px-6 py-3 text-left">Payment</th>
                                <th className="px-6 py-3 text-left">Reference</th>
                                <th className="px-6 py-3 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className={`divide-y ${isDarkMode ? 'divide-slate-700' : 'divide-gray-100'}`}>
                            {allocations.map((alloc) => (
                                <tr key={alloc.id} className={rowHover}>
                                    <td className={`px-6 py-4 ${textPrimary}`}>
                                        {new Date(alloc.purchased_at).toLocaleDateString()}
                                    </td>
                                    <td className={`px-6 py-4 font-mono text-xs ${textSecondary}`}>
                                        {alloc.merchant_id?.substring(0, 12)}...
                                    </td>
                                    <td className="px-6 py-4">
                                        <span className={`font-semibold ${textPrimary}`}>{alloc.amount_lcn?.toLocaleString()} LCN</span>
                                        <span className={`ml-2 ${textSecondary}`}>({alloc.amount_etb_paid?.toLocaleString()} ETB)</span>
                                    </td>
                                    <td className={`px-6 py-4 ${textSecondary}`}>{alloc.payment_method}</td>
                                    <td className={`px-6 py-4 font-mono text-xs ${textSecondary}`}>{alloc.payment_reference}</td>
                                    <td className="px-6 py-4 text-right">
                                        <div className="flex items-center justify-end gap-2">
                                            <button
                                                onClick={() => handleAction(alloc.id, 'APPROVE')}
                                                disabled={processing === alloc.id}
                                                className="flex items-center gap-1 px-3 py-1.5 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 disabled:opacity-50 text-white text-xs font-medium rounded-lg shadow-sm"
                                            >
                                                <CheckCircle className="h-3.5 w-3.5" />
                                                Approve
                                            </button>
                                            <button
                                                onClick={() => handleAction(alloc.id, 'REJECT')}
                                                disabled={processing === alloc.id}
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
