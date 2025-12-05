import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, ArrowDownLeft, ArrowUpRight, Clock } from 'lucide-react';
import { useStore } from '../store';

type FilterType = 'all' | 'received' | 'spent';

export const Transactions: React.FC = () => {
    const navigate = useNavigate();
    const { transactions } = useStore();
    const [filter, setFilter] = useState<FilterType>('all');

    const filteredTransactions = transactions.filter((tx) => {
        if (filter === 'all') return true;
        if (filter === 'received') return tx.direction === 'received';
        if (filter === 'spent') return tx.direction === 'sent';
        return true;
    });

    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const formatAddress = (addr: string) => {
        if (!addr) return '';
        return `${addr.slice(0, 12)}...${addr.slice(-8)}`;
    };

    return (
        <div>
            <div className="page-header">
                <button className="back-btn" onClick={() => navigate(-1)}>
                    <ArrowLeft size={20} />
                </button>
                <h1 className="page-title">Transaction History</h1>
            </div>

            {/* Filter Tabs */}
            <div style={{
                display: 'flex',
                gap: '0.5rem',
                marginBottom: '1.5rem',
                background: 'var(--bg-card)',
                padding: '0.25rem',
                borderRadius: '0.75rem',
            }}>
                {(['all', 'received', 'spent'] as FilterType[]).map((f) => (
                    <button
                        key={f}
                        onClick={() => setFilter(f)}
                        style={{
                            flex: 1,
                            padding: '0.75rem',
                            border: 'none',
                            borderRadius: '0.5rem',
                            background: filter === f ? 'var(--primary)' : 'transparent',
                            color: filter === f ? 'white' : 'var(--text-secondary)',
                            fontWeight: '600',
                            fontSize: '0.875rem',
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                            textTransform: 'capitalize',
                        }}
                    >
                        {f}
                    </button>
                ))}
            </div>

            {/* Transactions List */}
            {filteredTransactions.length === 0 ? (
                <div className="empty-state">
                    <Clock size={48} style={{ opacity: 0.3 }} />
                    <p style={{ marginTop: '1rem' }}>No transactions yet</p>
                    <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
                        {filter === 'all'
                            ? 'Your transaction history will appear here'
                            : `No ${filter} transactions`
                        }
                    </p>
                </div>
            ) : (
                <div className="tx-list">
                    {filteredTransactions.map((tx) => (
                        <div key={tx.tx_hash} className="tx-item" style={{ flexDirection: 'column', gap: '0.75rem' }}>
                            <div style={{ display: 'flex', alignItems: 'center', width: '100%', gap: '1rem' }}>
                                <div className={`tx-icon ${tx.direction === 'received' ? 'received' : 'spent'}`}>
                                    {tx.direction === 'received' ? (
                                        <ArrowDownLeft size={20} />
                                    ) : (
                                        <ArrowUpRight size={20} />
                                    )}
                                </div>
                                <div className="tx-details">
                                    <p className="tx-title">
                                        {tx.direction === 'received' ? 'Received LCN' : 'Spent LCN'}
                                    </p>
                                    <p className="tx-date">{formatDate(tx.submitted_at)}</p>
                                </div>
                                <span className={`tx-amount ${tx.direction === 'received' ? 'positive' : 'negative'}`}>
                                    {tx.direction === 'received' ? '+' : '-'}{tx.amount_lcn.toLocaleString()} LCN
                                </span>
                            </div>

                            {/* Transaction Details */}
                            <div style={{
                                width: '100%',
                                paddingLeft: '3.5rem',
                                fontSize: '0.75rem',
                                color: 'var(--text-muted)',
                            }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                                    <span>{tx.direction === 'received' ? 'From' : 'To'}:</span>
                                    <span style={{ fontFamily: 'monospace' }}>
                                        {formatAddress(tx.direction === 'received' ? tx.from_address : tx.to_address)}
                                    </span>
                                </div>
                                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <span>Status:</span>
                                    <span style={{
                                        color: tx.status === 'CONFIRMED' ? 'var(--success)' : 'var(--warning)',
                                        fontWeight: '500',
                                    }}>
                                        {tx.status}
                                    </span>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
