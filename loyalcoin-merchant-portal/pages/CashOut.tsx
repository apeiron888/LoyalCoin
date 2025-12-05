import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button, Input, Badge } from '../components/UIComponents';
import { ArrowLeft, CreditCard, CheckCircle, AlertCircle, Building } from 'lucide-react';
import { requestSettlement, getSettlementHistory, Settlement } from '../services/api';

// Exchange rate for display: 10 LCN = 1 ETB
const LCN_TO_ETB_RATE = 10;

export const CashOut: React.FC = () => {
    const navigate = useNavigate();
    const { wallet, fetchBalance, fetchTransactions } = useStore();
    const [amountLCN, setAmountLCN] = useState('');
    const [bankName, setBankName] = useState('');
    const [accountNumber, setAccountNumber] = useState('');
    const [accountHolder, setAccountHolder] = useState('');
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [settlements, setSettlements] = useState<Settlement[]>([]);
    const [activeTab, setActiveTab] = useState<'request' | 'history'>('request');

    // Convert LCN to ETB for display (10 LCN = 1 ETB)
    const lcnAmount = parseFloat(amountLCN) || 0;
    const etbAmount = lcnAmount / LCN_TO_ETB_RATE;

    useEffect(() => {
        loadHistory();
    }, []);

    const loadHistory = async () => {
        try {
            const response = await getSettlementHistory(20, 0);
            setSettlements(response.data.settlements || []);
        } catch (err) {
            console.error('Failed to load settlements:', err);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        if (lcnAmount <= 0) {
            setError('Please enter a valid amount');
            setLoading(false);
            return;
        }

        if (lcnAmount > wallet.balanceLCN) {
            setError(`Insufficient balance. You have ${wallet.balanceLCN.toLocaleString()} LCN.`);
            setLoading(false);
            return;
        }

        try {
            await requestSettlement(lcnAmount, {
                bank_name: bankName,
                account_number: accountNumber,
                account_holder: accountHolder
            });
            setSuccess(true);
            await Promise.all([fetchBalance(), fetchTransactions()]);
            loadHistory();
        } catch (err: any) {
            setError(err.message || 'Failed to request settlement');
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
                <h2 className="text-2xl font-bold mb-2 text-gray-900">Settlement Requested!</h2>
                <p className="text-gray-600 mb-6">
                    Your request to convert <strong>{lcnAmount.toLocaleString()} LCN</strong> to <strong>{etbAmount.toLocaleString()} ETB</strong> has been submitted.
                    Funds will be transferred to your bank account within 1-2 business days.
                </p>
                <div className="space-y-3">
                    <Button onClick={() => {
                        setSuccess(false);
                        setAmountLCN('');
                    }} className="w-full">
                        Make Another Request
                    </Button>
                    <Button variant="outline" onClick={() => navigate('/')} className="w-full">
                        Back to Dashboard
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-4xl mx-auto">
            <div className="mb-6 flex items-center">
                <Button variant="ghost" size="sm" onClick={() => navigate('/')} className="mr-4">
                    <ArrowLeft className="h-4 w-4 mr-1" /> Back
                </Button>
                <h1 className="text-2xl font-bold text-gray-900">Cash Out</h1>
            </div>

            <div className="flex border-b border-gray-200 mb-6">
                <button
                    className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'request'
                        ? 'text-amber-600 border-amber-600'
                        : 'text-gray-500 border-transparent hover:text-gray-700'
                        }`}
                    onClick={() => setActiveTab('request')}
                >
                    Request Settlement
                </button>
                <button
                    className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'history'
                        ? 'text-amber-600 border-amber-600'
                        : 'text-gray-500 border-transparent hover:text-gray-700'
                        }`}
                    onClick={() => setActiveTab('history')}
                >
                    Settlement History
                </button>
            </div>

            {activeTab === 'request' ? (
                <div className="grid gap-6 lg:grid-cols-2">
                    <Card className="p-6">
                        <div className="mb-6 p-4 rounded-lg bg-amber-50 border border-amber-200">
                            <p className="text-sm text-gray-600">Available Balance</p>
                            <p className="text-2xl font-bold text-amber-600">{wallet.balanceLCN.toLocaleString()} LCN</p>
                            <p className="text-xs text-gray-500">≈ {(wallet.balanceLCN / LCN_TO_ETB_RATE).toLocaleString()} ETB</p>
                        </div>

                        {error && (
                            <div className="mb-4 p-4 rounded-lg bg-red-50 flex items-center text-red-700">
                                <AlertCircle className="h-5 w-5 mr-2 flex-shrink-0" />
                                {error}
                            </div>
                        )}

                        <form onSubmit={handleSubmit} className="space-y-4">
                            <Input
                                label="Amount to Cash Out (LCN)"
                                type="number"
                                placeholder="Enter LCN amount"
                                value={amountLCN}
                                onChange={(e) => setAmountLCN(e.target.value)}
                                required
                                min="10"
                                step="10"
                            />

                            {amountLCN && (
                                <div className="p-4 bg-green-50 rounded-lg border border-green-200">
                                    <div className="flex justify-between items-center mb-2">
                                        <span className="text-sm text-gray-600">You convert:</span>
                                        <span className="text-lg font-bold text-gray-900">{lcnAmount.toLocaleString()} LCN</span>
                                    </div>
                                    <div className="flex justify-between items-center">
                                        <span className="text-sm text-gray-600">You receive:</span>
                                        <span className="text-xl font-bold text-green-600">{etbAmount.toLocaleString()} ETB</span>
                                    </div>
                                    <p className="text-xs text-gray-500 mt-2 text-center">Exchange rate: 10 LCN = 1 ETB</p>
                                </div>
                            )}

                            <Input
                                label="Bank Name"
                                placeholder="e.g., Commercial Bank of Ethiopia"
                                value={bankName}
                                onChange={(e) => setBankName(e.target.value)}
                                required
                            />

                            <Input
                                label="Account Number"
                                placeholder="Your bank account number"
                                value={accountNumber}
                                onChange={(e) => setAccountNumber(e.target.value)}
                                required
                            />

                            <Input
                                label="Account Holder Name"
                                placeholder="Name on the account"
                                value={accountHolder}
                                onChange={(e) => setAccountHolder(e.target.value)}
                                required
                            />

                            <Button
                                type="submit"
                                size="lg"
                                className="w-full"
                                isLoading={loading}
                                disabled={!amountLCN || !bankName || !accountNumber || !accountHolder}
                            >
                                Request Settlement
                            </Button>
                        </form>
                    </Card>

                    <Card className="p-6">
                        <h2 className="text-lg font-bold mb-4 text-gray-900">How Settlement Works</h2>

                        <div className="space-y-4">
                            <div className="flex gap-4">
                                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center text-amber-600 font-bold">1</div>
                                <div>
                                    <h3 className="font-medium text-gray-900">Submit Request</h3>
                                    <p className="text-sm text-gray-500">Enter the LCN amount and your bank details</p>
                                </div>
                            </div>

                            <div className="flex gap-4">
                                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center text-amber-600 font-bold">2</div>
                                <div>
                                    <h3 className="font-medium text-gray-900">Admin Review</h3>
                                    <p className="text-sm text-gray-500">Our team verifies and processes your request</p>
                                </div>
                            </div>

                            <div className="flex gap-4">
                                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center text-amber-600 font-bold">3</div>
                                <div>
                                    <h3 className="font-medium text-gray-900">Receive Funds</h3>
                                    <p className="text-sm text-gray-500">ETB deposited to your bank (1-2 business days)</p>
                                </div>
                            </div>

                            <div className="p-4 bg-blue-50 rounded-lg border border-blue-200 mt-6">
                                <h3 className="font-medium text-blue-800 mb-2">Exchange Rate</h3>
                                <p className="text-blue-700 text-lg font-bold">10 LCN = 1 ETB</p>
                                <p className="text-sm text-blue-600">No fees charged for settlements</p>
                            </div>
                        </div>
                    </Card>
                </div>
            ) : (
                <Card className="p-6">
                    <h2 className="text-lg font-bold mb-4 text-gray-900">Settlement History</h2>

                    {settlements.length === 0 ? (
                        <div className="text-center py-8">
                            <CreditCard className="mx-auto h-12 w-12 text-gray-300 mb-3" />
                            <p className="text-gray-500">No settlements yet</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {settlements.map((settlement) => (
                                <div key={settlement.id} className="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                                    <div className="flex items-center gap-4">
                                        <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center">
                                            <Building className="h-5 w-5 text-gray-600" />
                                        </div>
                                        <div>
                                            <p className="font-medium text-gray-900">{settlement.amount_lcn.toLocaleString()} LCN → {settlement.amount_etb.toLocaleString()} ETB</p>
                                            <p className="text-sm text-gray-500">
                                                {new Date(settlement.requested_at).toLocaleDateString()} • {settlement.bank_account.bank_name}
                                            </p>
                                        </div>
                                    </div>
                                    <Badge variant={
                                        settlement.status === 'COMPLETED' ? 'success' :
                                            settlement.status === 'PENDING' ? 'warning' : 'danger'
                                    }>
                                        {settlement.status}
                                    </Badge>
                                </div>
                            ))}
                        </div>
                    )}
                </Card>
            )}
        </div>
    );
};
