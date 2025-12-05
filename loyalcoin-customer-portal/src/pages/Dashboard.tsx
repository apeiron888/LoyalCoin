import React, { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { QrCode, Send, ArrowDownLeft, ArrowUpRight, RefreshCw } from 'lucide-react';
import { useStore } from '../store';

export const Dashboard: React.FC = () => {
    const { user, balance, transactions, fetchBalance, fetchTransactions, isLoading } = useStore();

    // Pull to refresh
    const handleRefresh = () => {
        fetchBalance();
        fetchTransactions();
    };

    // Get recent transactions (last 5)
    const recentTransactions = transactions.slice(0, 5);

    // Format date
    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    };

    // Format address
    const formatAddress = (addr: string) => {
        if (!addr) return '';
        return `${addr.slice(0, 10)}...${addr.slice(-6)}`;
    };

    return (
        <div>
            {/* Greeting */}
            <div style={{ marginBottom: '1rem' }}>
                <h1 style={{ fontSize: '1.25rem', fontWeight: '600' }}>
                    Welcome back{user?.username ? `, ${user.username}` : ''} 👋
                </h1>
            </div>

            {/* Balance Card */}
            <div className="balance-card">
                <p className="balance-label">Your Balance</p>
                <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'center' }}>
                    <span className="balance-amount">
                        {balance ? balance.lcn.toLocaleString(undefined, { maximumFractionDigits: 2 }) : '0'}
                    </span>
                    <span className="balance-currency">LCN</span>
                </div>
                <p className="balance-ada">
                    ≈ {balance ? balance.ada.toLocaleString(undefined, { maximumFractionDigits: 2 }) : '0'} ETB
                </p>
                <button
                    onClick={handleRefresh}
                    disabled={isLoading}
                    style={{
                        marginTop: '1rem',
                        background: 'rgba(0,0,0,0.2)',
                        border: 'none',
                        borderRadius: '0.5rem',
                        padding: '0.5rem 1rem',
                        color: 'white',
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        fontSize: '0.875rem',
                    }}
                >
                    <RefreshCw size={16} className={isLoading ? 'spinning' : ''} />
                    Refresh
                </button>
            </div>

            {/* Quick Actions */}
            <div className="quick-actions">
                <Link to="/receive" className="action-btn">
                    <div className="icon receive">
                        <QrCode size={24} />
                    </div>
                    <span>Receive Points</span>
                </Link>
                <Link to="/spend" className="action-btn">
                    <div className="icon spend">
                        <Send size={24} />
                    </div>
                    <span>Spend Points</span>
                </Link>
            </div>

            {/* Recent Transactions */}
            <div className="section-header">
                <h2 className="section-title">Recent Activity</h2>
                <Link to="/transactions" className="section-link">View all</Link>
            </div>

            {recentTransactions.length === 0 ? (
                <div className="empty-state">
                    <p>No transactions yet</p>
                    <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
                        Start earning points at participating merchants!
                    </p>
                </div>
            ) : (
                <div className="tx-list">
                    {recentTransactions.map((tx) => (
                        <div key={tx.tx_hash} className="tx-item">
                            <div className={`tx-icon ${tx.direction === 'received' ? 'received' : 'spent'}`}>
                                {tx.direction === 'received' ? (
                                    <ArrowDownLeft size={20} />
                                ) : (
                                    <ArrowUpRight size={20} />
                                )}
                            </div>
                            <div className="tx-details">
                                <p className="tx-title">
                                    {tx.direction === 'received' ? 'Received' : 'Spent'}
                                </p>
                                <p className="tx-date">
                                    {formatDate(tx.submitted_at)} • {formatAddress(tx.direction === 'received' ? tx.from_address : tx.to_address)}
                                </p>
                            </div>
                            <span className={`tx-amount ${tx.direction === 'received' ? 'positive' : 'negative'}`}>
                                {tx.direction === 'received' ? '+' : '-'}{tx.amount_lcn.toLocaleString()} LCN
                            </span>
                        </div>
                    ))}
                </div>
            )}

            {/* Wallet Address */}
            <div style={{ marginTop: '1.5rem' }}>
                <p className="wallet-label">Your Wallet Address</p>
                <div className="address-display">
                    <span style={{ flex: 1 }}>{user?.wallet_address ? formatAddress(user.wallet_address) : 'Loading...'}</span>
                </div>
            </div>
        </div>
    );
};
