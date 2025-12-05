import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button, Input } from '../components/UIComponents';
import { Coins, ArrowLeft, CheckCircle, AlertCircle, Camera, X } from 'lucide-react';
import { issueLCN } from '../services/api';
import { Html5Qrcode } from 'html5-qrcode';

export const IssueLCN: React.FC = () => {
    const navigate = useNavigate();
    const { wallet, fetchBalance, fetchTransactions } = useStore();
    const [amount, setAmount] = useState('');
    const [address, setAddress] = useState('');
    const [note, setNote] = useState('');
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [txHash, setTxHash] = useState<string | null>(null);
    const [showScanner, setShowScanner] = useState(false);
    const scannerRef = useRef<Html5Qrcode | null>(null);

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

        try {
            const scanner = new Html5Qrcode('qr-reader');
            scannerRef.current = scanner;

            await scanner.start(
                { facingMode: 'environment' },
                { fps: 10, qrbox: { width: 250, height: 250 } },
                (decodedText) => {
                    // Check if it's a valid Cardano address
                    if (decodedText.startsWith('addr_test1') || decodedText.startsWith('addr1')) {
                        setAddress(decodedText);
                        stopScanner();
                    } else {
                        setError('Invalid address format. Please scan a valid Cardano wallet QR code.');
                    }
                },
                () => { } // Ignore errors during scanning
            );
        } catch (err: any) {
            console.error('Failed to start scanner:', err);
            setError('Failed to access camera. Please ensure camera permissions are granted.');
            setShowScanner(false);
        }
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

    const handleIssue = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        const amountNum = parseFloat(amount);
        if (isNaN(amountNum) || amountNum <= 0) {
            setError('Please enter a valid amount');
            setLoading(false);
            return;
        }

        if (amountNum > wallet.balanceLCN) {
            setError(`Insufficient balance. You have ${wallet.balanceLCN.toLocaleString()} LCN.`);
            setLoading(false);
            return;
        }

        try {
            const response = await issueLCN(address, amountNum, note);
            setTxHash(response.data.tx_hash);
            setSuccess(true);
            await Promise.all([fetchBalance(), fetchTransactions()]);
        } catch (err: any) {
            setError(err.message || 'Failed to issue LCN');
        } finally {
            setLoading(false);
        }
    };

    if (success) {
        return (
            <div className="max-w-md mx-auto mt-12 text-center">
                <div className="mx-auto h-16 w-16 rounded-full bg-green-100 flex items-center justify-center mb-6">
                    <CheckCircle className="h-8 w-8 text-green-600" />
                </div>
                <h2 className="text-2xl font-bold mb-2 text-gray-900">LCN Issued Successfully!</h2>
                <p className="text-gray-600 mb-6">
                    You have sent {amount} LCN to the customer.
                </p>
                {txHash && (
                    <div className="p-4 rounded-lg mb-6 bg-gray-50 break-all">
                        <p className="text-xs uppercase tracking-wide mb-1 text-gray-500">Transaction Hash</p>
                        <p className="text-sm font-mono text-gray-700">{txHash}</p>
                    </div>
                )}
                <div className="space-y-3">
                    <Button onClick={() => {
                        setSuccess(false);
                        setAmount('');
                        setAddress('');
                        setNote('');
                        setTxHash(null);
                    }} className="w-full">
                        Issue More
                    </Button>
                    <Button variant="outline" onClick={() => navigate('/')} className="w-full">
                        Back to Dashboard
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto">
            <div className="mb-6 flex items-center">
                <Button variant="ghost" size="sm" onClick={() => navigate('/')} className="mr-4">
                    <ArrowLeft className="h-4 w-4 mr-1" /> Back
                </Button>
                <h1 className="text-2xl font-bold text-gray-900">Issue Rewards</h1>
            </div>

            <Card className="p-6">
                <div className="mb-6 p-4 rounded-lg bg-amber-50 flex items-start">
                    <Coins className="h-5 w-5 mt-0.5 mr-3 flex-shrink-0 text-amber-600" />
                    <div>
                        <h3 className="text-sm font-medium text-gray-900">Current Balance</h3>
                        <p className="text-2xl font-bold text-amber-600">{wallet.balanceLCN.toLocaleString()} LCN</p>
                        <p className="text-xs mt-1 text-gray-500">
                            Available to issue to customers
                        </p>
                    </div>
                </div>

                {error && (
                    <div className="mb-6 p-4 rounded-lg bg-red-50 flex items-center text-red-700">
                        <AlertCircle className="h-5 w-5 mr-2 flex-shrink-0" />
                        {error}
                    </div>
                )}

                {/* QR Scanner Modal */}
                {showScanner && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-75">
                        <div className="bg-white rounded-xl p-6 max-w-sm w-full mx-4">
                            <div className="flex justify-between items-center mb-4">
                                <h3 className="font-semibold text-lg text-gray-900">Scan Customer QR Code</h3>
                                <button
                                    onClick={stopScanner}
                                    className="p-1 hover:bg-gray-100 rounded-full"
                                >
                                    <X className="h-6 w-6 text-gray-500" />
                                </button>
                            </div>
                            <div id="qr-reader" className="w-full rounded-lg overflow-hidden"></div>
                            <p className="text-sm text-gray-500 mt-4 text-center">
                                Point camera at customer's wallet QR code
                            </p>
                        </div>
                    </div>
                )}

                <form onSubmit={handleIssue} className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">
                            Customer Wallet Address
                        </label>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                className="flex-1 px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:border-amber-500"
                                placeholder="addr_test1..."
                                value={address}
                                onChange={(e) => setAddress(e.target.value)}
                                required
                            />
                            <button
                                type="button"
                                onClick={startScanner}
                                className="px-4 py-2.5 bg-amber-100 text-amber-700 rounded-lg hover:bg-amber-200 transition-colors flex items-center gap-2"
                                title="Scan QR Code"
                            >
                                <Camera className="h-5 w-5" />
                                <span className="hidden sm:inline">Scan</span>
                            </button>
                        </div>
                        <p className="text-xs text-gray-500 mt-1">
                            Enter address or scan customer's QR code
                        </p>
                    </div>

                    <Input
                        label="Amount (LCN)"
                        type="number"
                        placeholder="0.00"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        required
                        min="0.01"
                        step="0.01"
                    />

                    <Input
                        label="Note (Optional)"
                        placeholder="e.g., Coffee purchase reward"
                        value={note}
                        onChange={(e) => setNote(e.target.value)}
                    />

                    <Button
                        type="submit"
                        size="lg"
                        className="w-full"
                        isLoading={loading}
                        disabled={!amount || !address}
                    >
                        Issue {amount ? `${amount} LCN` : 'Rewards'}
                    </Button>
                </form>
            </Card>
        </div>
    );
};