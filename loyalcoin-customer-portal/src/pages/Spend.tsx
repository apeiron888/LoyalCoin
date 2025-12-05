import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, CheckCircle, AlertCircle, Camera, X } from 'lucide-react';
import { useStore } from '../store';
import { redeemLCN } from '../services/api';
import { Html5Qrcode } from 'html5-qrcode';

export const Spend: React.FC = () => {
    const navigate = useNavigate();
    const { balance, fetchBalance, fetchTransactions } = useStore();

    const [merchantAddress, setMerchantAddress] = useState('');
    const [amount, setAmount] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<{ txHash: string; amount: number } | null>(null);
    const [showScanner, setShowScanner] = useState(false);
    const scannerRef = useRef<Html5Qrcode | null>(null);

    const lcnAmount = parseFloat(amount) || 0;
    const etbEquivalent = lcnAmount / 10; // 10 LCN = 1 ETB

    // Cleanup scanner on unmount
    useEffect(() => {
        return () => {
            if (scannerRef.current) {
                scannerRef.current.stop().catch(() => { });
            }
        };
    }, []);

    const startScanner = async () => {
        setShowScanner(true);
        setError(null);

        // Wait for the DOM element to be rendered
        setTimeout(async () => {
            try {
                const element = document.getElementById('qr-reader');
                if (!element) {
                    setError('Scanner element not found. Please try again.');
                    setShowScanner(false);
                    return;
                }

                const scanner = new Html5Qrcode('qr-reader');
                scannerRef.current = scanner;

                await scanner.start(
                    { facingMode: 'environment' },
                    { fps: 10, qrbox: { width: 250, height: 250 } },
                    (decodedText) => {
                        // Check if it's a valid Cardano address
                        if (decodedText.startsWith('addr_test1') || decodedText.startsWith('addr1')) {
                            setMerchantAddress(decodedText);
                            stopScanner();
                        } else {
                            setError('Invalid address. Please scan a valid merchant QR code.');
                        }
                    },
                    () => { } // Ignore errors during scanning
                );
            } catch (err: any) {
                console.error('Failed to start scanner:', err);
                setError('Failed to access camera. Please ensure camera permissions are granted.');
                setShowScanner(false);
            }
        }, 100); // Small delay to ensure DOM is ready
    };

    const stopScanner = async () => {
        if (scannerRef.current) {
            try {
                await scannerRef.current.stop();
            } catch (err) {
                console.error('Error stopping scanner:', err);
            }
        }
        setShowScanner(false);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        // Validation
        if (!merchantAddress.startsWith('addr_test1') && !merchantAddress.startsWith('addr1')) {
            setError('Please enter a valid Cardano address');
            return;
        }

        if (lcnAmount <= 0) {
            setError('Please enter a valid amount');
            return;
        }

        if (balance && lcnAmount > balance.lcn) {
            setError(`Insufficient balance. You have ${balance.lcn.toLocaleString()} LCN`);
            return;
        }

        setLoading(true);

        try {
            const response = await redeemLCN(merchantAddress, lcnAmount);
            setSuccess({ txHash: response.data.tx_hash, amount: lcnAmount });

            // Refresh balance
            setTimeout(() => {
                fetchBalance();
                fetchTransactions();
            }, 2000);
        } catch (err: any) {
            setError(err.message || 'Transaction failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    if (success) {
        return (
            <div>
                <div className="page-header">
                    <button className="back-btn" onClick={() => navigate('/')}>
                        <ArrowLeft size={20} />
                    </button>
                    <h1 className="page-title">Transaction Sent!</h1>
                </div>

                <div className="card-glass text-center">
                    <div style={{
                        width: '80px',
                        height: '80px',
                        borderRadius: '50%',
                        background: 'rgba(16, 185, 129, 0.2)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        margin: '0 auto 1.5rem'
                    }}>
                        <CheckCircle size={40} color="#10B981" />
                    </div>

                    <h2 style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>
                        -{success.amount.toLocaleString()} LCN
                    </h2>
                    <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>
                        Successfully sent to merchant
                    </p>

                    <div className="address-display" style={{ marginBottom: '1rem' }}>
                        <span style={{ flex: 1, fontSize: '0.7rem', wordBreak: 'break-all' }}>
                            TX: {success.txHash}
                        </span>
                    </div>

                    <button
                        className="btn btn-primary btn-block"
                        onClick={() => navigate('/')}
                    >
                        Back to Home
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div>
            <div className="page-header">
                <button className="back-btn" onClick={() => navigate(-1)}>
                    <ArrowLeft size={20} />
                </button>
                <h1 className="page-title">Spend Points</h1>
            </div>

            {/* Current Balance */}
            <div className="card mb-3" style={{ textAlign: 'center' }}>
                <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>Available Balance</p>
                <p style={{ fontSize: '1.5rem', fontWeight: '700', color: 'var(--primary)' }}>
                    {balance ? balance.lcn.toLocaleString() : '0'} LCN
                </p>
            </div>

            {/* QR Scanner Modal */}
            {showScanner && (
                <div style={{
                    position: 'fixed',
                    inset: 0,
                    zIndex: 50,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'rgba(0, 0, 0, 0.85)',
                    padding: '1rem'
                }}>
                    <div style={{
                        background: 'var(--bg-secondary)',
                        borderRadius: '1rem',
                        padding: '1.5rem',
                        maxWidth: '340px',
                        width: '100%'
                    }}>
                        <div style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            marginBottom: '1rem'
                        }}>
                            <h3 style={{ fontSize: '1.1rem', fontWeight: '600' }}>Scan Merchant QR</h3>
                            <button
                                onClick={stopScanner}
                                style={{
                                    background: 'var(--bg-card)',
                                    border: 'none',
                                    borderRadius: '50%',
                                    padding: '0.5rem',
                                    cursor: 'pointer',
                                    color: 'var(--text-primary)'
                                }}
                            >
                                <X size={20} />
                            </button>
                        </div>
                        <div
                            id="qr-reader"
                            style={{
                                width: '100%',
                                borderRadius: '0.75rem',
                                overflow: 'hidden',
                                background: '#000'
                            }}
                        ></div>
                        <p style={{
                            fontSize: '0.875rem',
                            color: 'var(--text-muted)',
                            marginTop: '1rem',
                            textAlign: 'center'
                        }}>
                            Point camera at merchant's QR code
                        </p>
                    </div>
                </div>
            )}

            <form onSubmit={handleSubmit}>
                {error && (
                    <div className="alert alert-error" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <AlertCircle size={18} />
                        {error}
                    </div>
                )}

                <div className="form-group">
                    <label className="form-label">Merchant Wallet Address</label>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="addr_test1..."
                            value={merchantAddress}
                            onChange={(e) => setMerchantAddress(e.target.value)}
                            style={{ flex: 1 }}
                            required
                        />
                        <button
                            type="button"
                            onClick={startScanner}
                            className="btn btn-secondary"
                            style={{
                                padding: '0.75rem 1rem',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.5rem'
                            }}
                        >
                            <Camera size={20} />
                            <span style={{ display: 'none' }}>Scan</span>
                        </button>
                    </div>
                    <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.5rem' }}>
                        Tap the camera button to scan merchant's QR code
                    </p>
                </div>

                <div className="form-group">
                    <label className="form-label">Amount (LCN)</label>
                    <input
                        type="number"
                        className="form-input"
                        placeholder="Enter amount"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        min="1"
                        step="1"
                        required
                    />
                </div>

                {lcnAmount > 0 && (
                    <div className="card mb-3" style={{
                        background: 'rgba(245, 158, 11, 0.1)',
                        borderColor: 'rgba(245, 158, 11, 0.3)'
                    }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                            <span style={{ color: 'var(--text-secondary)' }}>You spend:</span>
                            <span style={{ fontWeight: '600' }}>{lcnAmount.toLocaleString()} LCN</span>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                            <span style={{ color: 'var(--text-secondary)' }}>Value:</span>
                            <span style={{ fontWeight: '600', color: 'var(--primary)' }}>
                                ≈ {etbEquivalent.toLocaleString()} ETB
                            </span>
                        </div>
                    </div>
                )}

                <button
                    type="submit"
                    className="btn btn-primary btn-block"
                    disabled={loading || !merchantAddress || !amount}
                >
                    {loading ? <span className="spinner" /> : 'Send Points'}
                </button>
            </form>
        </div>
    );
};
