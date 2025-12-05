import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Copy, Check } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { useStore } from '../store';

export const Receive: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useStore();
    const [copied, setCopied] = useState(false);

    const walletAddress = user?.wallet_address || '';

    const copyAddress = async () => {
        try {
            await navigator.clipboard.writeText(walletAddress);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    return (
        <div>
            <div className="page-header">
                <button className="back-btn" onClick={() => navigate(-1)}>
                    <ArrowLeft size={20} />
                </button>
                <h1 className="page-title">Receive Points</h1>
            </div>

            <div className="card-glass text-center">
                <p style={{ marginBottom: '1rem', color: 'var(--text-secondary)' }}>
                    Show this QR code to a merchant to receive LCN
                </p>

                {/* QR Code */}
                <div className="qr-container">
                    <QRCodeSVG
                        value={walletAddress}
                        size={200}
                        level="H"
                        includeMargin={true}
                        bgColor="#ffffff"
                        fgColor="#1a1a2e"
                    />
                </div>

                {/* Address */}
                <div className="address-display" style={{ marginTop: '1rem' }}>
                    <span style={{
                        flex: 1,
                        wordBreak: 'break-all',
                        fontSize: '0.7rem',
                        lineHeight: '1.4',
                    }}>
                        {walletAddress}
                    </span>
                    <button className="copy-btn" onClick={copyAddress}>
                        {copied ? <Check size={18} /> : <Copy size={18} />}
                    </button>
                </div>

                {copied && (
                    <p style={{
                        color: 'var(--success)',
                        fontSize: '0.875rem',
                        marginTop: '0.75rem'
                    }}>
                        Address copied!
                    </p>
                )}
            </div>

            <div className="card mt-3" style={{ textAlign: 'center' }}>
                <h3 style={{ fontSize: '1rem', marginBottom: '0.5rem' }}>How it works</h3>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                    When you make a purchase at a participating merchant,
                    show them this QR code to receive loyalty points in your wallet.
                </p>
            </div>
        </div>
    );
};
