import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Copy, Check, LogOut, User } from 'lucide-react';
import { useStore } from '../store';

export const Profile: React.FC = () => {
    const navigate = useNavigate();
    const { user, balance, logout } = useStore();
    const [copied, setCopied] = useState(false);

    const copyAddress = async () => {
        if (!user?.wallet_address) return;
        try {
            await navigator.clipboard.writeText(user.wallet_address);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    const handleLogout = () => {
        logout();
        localStorage.removeItem('customer_user');
        navigate('/login');
    };

    return (
        <div>
            <div className="page-header">
                <button className="back-btn" onClick={() => navigate(-1)}>
                    <ArrowLeft size={20} />
                </button>
                <h1 className="page-title">Profile</h1>
            </div>

            {/* Profile Card */}
            <div className="card-glass profile-card">
                <div className="profile-avatar">
                    <User size={36} color="white" />
                </div>
                <h2 className="profile-name">{user?.username || 'Customer'}</h2>
                <p className="profile-email">{user?.email}</p>
            </div>

            {/* Stats */}
            <div className="card mt-3">
                <h3 style={{ fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '1rem' }}>
                    WALLET STATS
                </h3>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>LCN Balance</span>
                    <span style={{ fontWeight: '600' }}>
                        {balance ? balance.lcn.toLocaleString() : '0'} LCN
                    </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>ADA Balance</span>
                    <span style={{ fontWeight: '600' }}>
                        {balance ? balance.ada.toLocaleString(undefined, { maximumFractionDigits: 2 }) : '0'} ADA
                    </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>ETB Value</span>
                    <span style={{ fontWeight: '600', color: 'var(--success)' }}>
                        ≈ {balance ? (balance.lcn / 10).toLocaleString() : '0'} ETB
                    </span>
                </div>
            </div>

            {/* Wallet Address */}
            <div className="card mt-3">
                <h3 style={{ fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
                    WALLET ADDRESS
                </h3>
                <div className="address-display">
                    <span style={{
                        flex: 1,
                        wordBreak: 'break-all',
                        fontSize: '0.7rem',
                        lineHeight: '1.4',
                    }}>
                        {user?.wallet_address || 'Loading...'}
                    </span>
                    <button className="copy-btn" onClick={copyAddress}>
                        {copied ? <Check size={18} /> : <Copy size={18} />}
                    </button>
                </div>
                {copied && (
                    <p style={{
                        color: 'var(--success)',
                        fontSize: '0.75rem',
                        marginTop: '0.5rem',
                        textAlign: 'center',
                    }}>
                        Copied to clipboard!
                    </p>
                )}
            </div>

            {/* About */}
            <div className="card mt-3">
                <h3 style={{ fontSize: '0.875rem', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
                    ABOUT
                </h3>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>App Version</span>
                    <span>1.0.0</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>Network</span>
                    <span>Cardano Preprod</span>
                </div>
            </div>

            {/* Logout Button */}
            <button
                className="btn btn-outline btn-block mt-3"
                onClick={handleLogout}
                style={{ color: 'var(--danger)', borderColor: 'var(--danger)' }}
            >
                <LogOut size={18} />
                Sign Out
            </button>
        </div>
    );
};
