import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button, Input, Badge } from '../components/UIComponents';
import { ArrowLeft, Coins, CreditCard, CheckCircle, AlertCircle, Building } from 'lucide-react';
import { purchaseAllocation, getAllocationHistory, Allocation } from '../services/api';

// Exchange rate for display: 1 tADA = 100 LCN
const ETB_TO_LCN_RATE = 100;

export const BuyLCN: React.FC = () => {
  const navigate = useNavigate();
  const { fetchBalance, fetchTransactions } = useStore();
  const [amountLCN, setAmountLCN] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('BANK_TRANSFER');
  const [paymentReference, setPaymentReference] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [purchaseHistory, setPurchaseHistory] = useState<Allocation[]>([]);
  const [activeTab, setActiveTab] = useState<'buy' | 'history'>('buy');

  // Convert LCN to ETB for display (1 ETB = 10 LCN)
  const lcnAmount = parseFloat(amountLCN) || 0;
  const etbAmount = lcnAmount / ETB_TO_LCN_RATE;

  useEffect(() => {
    loadHistory();
  }, []);

  const loadHistory = async () => {
    try {
      const response = await getAllocationHistory(20, 0);
      setPurchaseHistory(response.data.allocations || []);
    } catch (err) {
      console.error('Failed to load history:', err);
    }
  };

  const handlePurchase = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (lcnAmount <= 0) {
      setError('Please enter a valid amount');
      setLoading(false);
      return;
    }

    try {
      // Send whole LCN amount - backend expects whole LCN, not atomic units
      await purchaseAllocation(lcnAmount, paymentMethod, paymentReference);
      setSuccess(true);
      await Promise.all([fetchBalance(), fetchTransactions()]);
      loadHistory();
    } catch (err: any) {
      setError(err.message || 'Failed to submit purchase request');
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
        <h2 className="text-2xl font-bold mb-2 text-gray-900">Purchase Request Submitted!</h2>
        <p className="text-gray-600 mb-6">
          Your request to purchase <strong>{lcnAmount.toLocaleString()} LCN</strong> for <strong>{etbAmount.toLocaleString()} ETB</strong> has been submitted.
          An admin will verify your payment and send the LCN to your wallet.
        </p>
        <div className="space-y-3">
          <Button onClick={() => {
            setSuccess(false);
            setAmountLCN('');
            setPaymentReference('');
          }} className="w-full">
            Make Another Purchase
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
        <h1 className="text-2xl font-bold text-gray-900">Buy LCN</h1>
      </div>

      <div className="flex border-b border-gray-200 mb-6">
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'buy'
            ? 'text-amber-600 border-amber-600'
            : 'text-gray-500 border-transparent hover:text-gray-700'
            }`}
          onClick={() => setActiveTab('buy')}
        >
          Purchase LCN
        </button>
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'history'
            ? 'text-amber-600 border-amber-600'
            : 'text-gray-500 border-transparent hover:text-gray-700'
            }`}
          onClick={() => setActiveTab('history')}
        >
          Purchase History
        </button>
      </div>

      {activeTab === 'buy' ? (
        <div className="grid gap-6 lg:grid-cols-2">
          <Card className="p-6">
            <h2 className="text-lg font-bold mb-4 text-gray-900">Purchase Details</h2>

            {error && (
              <div className="mb-4 p-4 rounded-lg bg-red-50 flex items-center text-red-700">
                <AlertCircle className="h-5 w-5 mr-2 flex-shrink-0" />
                {error}
              </div>
            )}

            <form onSubmit={handlePurchase} className="space-y-4">
              <Input
                label="Amount (LCN)"
                type="number"
                placeholder="Enter LCN amount"
                value={amountLCN}
                onChange={(e) => setAmountLCN(e.target.value)}
                required
                min="10"
                step="10"
              />

              <div className="p-4 bg-amber-50 rounded-lg border border-amber-200">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm text-gray-600">You will receive:</span>
                  <span className="text-xl font-bold text-amber-600">{lcnAmount.toLocaleString()} LCN</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">You pay:</span>
                  <span className="text-lg font-bold text-gray-900">{etbAmount.toLocaleString()} ETB</span>
                </div>
                <p className="text-xs text-gray-500 mt-2 text-center">Exchange rate: 1 ETB = 10 LCN</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Payment Method</label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setPaymentMethod('BANK_TRANSFER')}
                    className={`p-3 border rounded-lg flex items-center justify-center gap-2 ${paymentMethod === 'BANK_TRANSFER'
                      ? 'border-amber-500 bg-amber-50 text-amber-700'
                      : 'border-gray-200 hover:border-gray-300'
                      }`}
                  >
                    <Building className="h-4 w-4" />
                    Bank Transfer
                  </button>
                  <button
                    type="button"
                    onClick={() => setPaymentMethod('MOBILE_MONEY')}
                    className={`p-3 border rounded-lg flex items-center justify-center gap-2 ${paymentMethod === 'MOBILE_MONEY'
                      ? 'border-amber-500 bg-amber-50 text-amber-700'
                      : 'border-gray-200 hover:border-gray-300'
                      }`}
                  >
                    <CreditCard className="h-4 w-4" />
                    Mobile Money
                  </button>
                </div>
              </div>

              <Input
                label="Payment Reference"
                placeholder="Transaction ID or reference number"
                value={paymentReference}
                onChange={(e) => setPaymentReference(e.target.value)}
                required
              />

              <Button
                type="submit"
                size="lg"
                className="w-full"
                isLoading={loading}
                disabled={!amountLCN || !paymentReference}
              >
                Submit Purchase Request
              </Button>
            </form>
          </Card>

          <Card className="p-6">
            <h2 className="text-lg font-bold mb-4 text-gray-900">Payment Instructions</h2>

            <div className="space-y-4">
              <div className="p-4 bg-gray-50 rounded-lg">
                <h3 className="font-medium text-gray-900 mb-2">Bank Transfer</h3>
                <div className="text-sm text-gray-600 space-y-1">
                  <p><span className="font-medium">Bank:</span> Commercial Bank of Ethiopia</p>
                  <p><span className="font-medium">Account:</span> 1000123456789</p>
                  <p><span className="font-medium">Name:</span> LoyalCoin Ltd</p>
                </div>
              </div>

              <div className="border-l-4 border-amber-400 bg-amber-50 p-4 rounded-r-lg">
                <h3 className="font-medium text-amber-800 mb-2">Important</h3>
                <ul className="text-sm text-amber-700 list-disc list-inside space-y-1">
                  <li>Include your email as payment reference</li>
                  <li>LCN will be credited after admin verification</li>
                  <li>Processing usually takes 1-2 business hours</li>
                </ul>
              </div>

              <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
                <h3 className="font-medium text-blue-800 mb-2">Exchange Rate</h3>
                <p className="text-blue-700 text-lg font-bold">1 ETB = 10 LCN</p>
              </div>
            </div>
          </Card>
        </div>
      ) : (
        <Card className="p-6">
          <h2 className="text-lg font-bold mb-4 text-gray-900">Purchase History</h2>

          {purchaseHistory.length === 0 ? (
            <div className="text-center py-8">
              <Coins className="mx-auto h-12 w-12 text-gray-300 mb-3" />
              <p className="text-gray-500">No purchase history yet</p>
            </div>
          ) : (
            <div className="space-y-3">
              {purchaseHistory.map((purchase) => (
                <div key={purchase.id} className="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                  <div>
                    <p className="font-medium text-gray-900">{purchase.amount_lcn.toLocaleString()} LCN</p>
                    <p className="text-sm text-gray-500">
                      {new Date(purchase.purchased_at).toLocaleDateString()} • Paid {purchase.amount_etb_paid.toLocaleString()} ETB
                    </p>
                  </div>
                  <Badge variant={
                    purchase.status === 'APPROVED' ? 'success' :
                      purchase.status === 'PENDING' ? 'warning' : 'danger'
                  }>
                    {purchase.status}
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