import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button } from '../components/UIComponents';
import { ArrowLeft, Copy, Check, QrCode } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';

export const Receive: React.FC = () => {
    const navigate = useNavigate();
    const { user, wallet, fetchBalance } = useStore();
    const [copied, setCopied] = useState(false);

    const walletAddress = user?.walletAddress || '';

    const copyAddress = async () => {
        try {
            await navigator.clipboard.writeText(walletAddress);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    const handleRefresh = () => {
        fetchBalance();
    };

    return (
        <div className="max-w-2xl mx-auto">
            <div className="mb-6 flex items-center">
                <Button variant="ghost" size="sm" onClick={() => navigate('/')} className="mr-4">
                    <ArrowLeft className="h-4 w-4 mr-1" /> Back
                </Button>
                <h1 className="text-2xl font-bold text-gray-900">Receive Coins</h1>
            </div>

            <Card className="p-6">
                {/* Current Balance */}
                <div className="mb-6 p-4 rounded-lg bg-green-50 text-center">
                    <p className="text-sm text-gray-600 mb-1">Current Balance</p>
                    <p className="text-3xl font-bold text-green-600">
                        {wallet.balanceLCN.toLocaleString()} LCN
                    </p>
                    <button
                        onClick={handleRefresh}
                        className="mt-2 text-sm text-green-600 hover:text-green-700 underline"
                    >
                        Refresh Balance
                    </button>
                </div>

                {/* QR Code */}
                <div className="flex flex-col items-center mb-6">
                    <div className="p-4 bg-white rounded-xl shadow-lg border">
                        <QRCodeSVG
                            value={walletAddress}
                            size={220}
                            level="H"
                            includeMargin={true}
                            bgColor="#ffffff"
                            fgColor="#1f2937"
                        />
                    </div>
                    <p className="mt-4 text-sm text-gray-500 text-center">
                        Customers can scan this QR code to send you LCN
                    </p>
                </div>

                {/* Wallet Address */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Your Wallet Address
                    </label>
                    <div className="flex items-center gap-2">
                        <div className="flex-1 p-3 bg-gray-50 rounded-lg border border-gray-200 font-mono text-xs break-all">
                            {walletAddress}
                        </div>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={copyAddress}
                            className="flex-shrink-0"
                        >
                            {copied ? (
                                <>
                                    <Check className="h-4 w-4 mr-1 text-green-600" />
                                    Copied
                                </>
                            ) : (
                                <>
                                    <Copy className="h-4 w-4 mr-1" />
                                    Copy
                                </>
                            )}
                        </Button>
                    </div>
                </div>

                {/* Instructions */}
                <div className="mt-6 p-4 bg-amber-50 rounded-lg border border-amber-100">
                    <div className="flex items-start">
                        <QrCode className="h-5 w-5 text-amber-600 mt-0.5 mr-3 flex-shrink-0" />
                        <div>
                            <h3 className="font-medium text-gray-900 mb-1">How to receive coins</h3>
                            <ol className="text-sm text-gray-600 space-y-1">
                                <li>1. Show this QR code to the customer</li>
                                <li>2. Customer scans it with their LoyalCoin app</li>
                                <li>3. Customer enters the LCN amount to redeem</li>
                                <li>4. LCN will be transferred to your wallet</li>
                            </ol>
                        </div>
                    </div>
                </div>
            </Card>
        </div>
    );
};
