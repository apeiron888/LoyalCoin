import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button, Badge } from '../components/UIComponents';
import { ArrowLeft, TrendingUp, TrendingDown, CreditCard, Coins, RefreshCw } from 'lucide-react';
import { TransactionType } from '../types';

export const Transactions: React.FC = () => {
    const navigate = useNavigate();
    const { wallet, fetchTransactions } = useStore();
    const [isRefreshing, setIsRefreshing] = React.useState(false);

    const handleRefresh = async () => {
        setIsRefreshing(true);
        await fetchTransactions();
        setIsRefreshing(false);
    };

    const getTypeIcon = (type: TransactionType) => {
        switch (type) {
            case TransactionType.ISSUE:
                return <TrendingUp size={18} />;
            case TransactionType.RECEIVE:
                return <TrendingDown size={18} />;
            case TransactionType.SETTLEMENT:
                return <CreditCard size={18} />;
            default:
                return <Coins size={18} />;
        }
    };

    const getTypeLabel = (type: TransactionType) => {
        switch (type) {
            case TransactionType.ISSUE:
                return 'Issued Reward';
            case TransactionType.RECEIVE:
                return 'Received LCN';
            case TransactionType.SETTLEMENT:
                return 'Settlement';
            case TransactionType.ALLOCATION:
                return 'LCN Purchase';
            default:
                return type;
        }
    };

    const getTypeColor = (type: TransactionType) => {
        switch (type) {
            case TransactionType.ISSUE:
                return 'bg-orange-100 text-orange-600';
            case TransactionType.RECEIVE:
                return 'bg-green-100 text-green-600';
            case TransactionType.SETTLEMENT:
                return 'bg-blue-100 text-blue-600';
            case TransactionType.ALLOCATION:
                return 'bg-purple-100 text-purple-600';
            default:
                return 'bg-gray-100 text-gray-600';
        }
    };

    return (
        <div className="max-w-4xl mx-auto">
            <div className="mb-6 flex items-center justify-between">
                <div className="flex items-center">
                    <Button variant="ghost" size="sm" onClick={() => navigate('/')} className="mr-4">
                        <ArrowLeft className="h-4 w-4 mr-1" /> Back
                    </Button>
                    <h1 className="text-2xl font-bold text-gray-900">Transactions</h1>
                </div>
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleRefresh}
                    disabled={isRefreshing}
                >
                    <RefreshCw className={`h-4 w-4 mr-1 ${isRefreshing ? 'animate-spin' : ''}`} />
                    Refresh
                </Button>
            </div>

            <Card className="p-6">
                {wallet.transactions.length === 0 ? (
                    <div className="text-center py-12">
                        <Coins className="mx-auto h-12 w-12 text-gray-300 mb-3" />
                        <p className="text-gray-500 text-lg">No transactions yet</p>
                        <p className="text-sm text-gray-400 mt-1">Your transaction history will appear here</p>
                        <Button
                            variant="primary"
                            className="mt-4"
                            onClick={() => navigate('/issue')}
                        >
                            Issue Your First Reward
                        </Button>
                    </div>
                ) : (
                    <div className="space-y-4">
                        {wallet.transactions.map((tx) => (
                            <div
                                key={tx.id}
                                className="flex items-center justify-between p-4 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
                            >
                                <div className="flex items-center gap-4">
                                    <div className={`flex h-12 w-12 items-center justify-center rounded-full ${getTypeColor(tx.type)}`}>
                                        {getTypeIcon(tx.type)}
                                    </div>
                                    <div>
                                        <p className="font-medium text-gray-900">{getTypeLabel(tx.type)}</p>
                                        <p className="text-sm text-gray-500">
                                            {new Date(tx.date).toLocaleString()}
                                        </p>
                                        {tx.address && (
                                            <p className="text-xs text-gray-400 font-mono">
                                                {tx.address.substring(0, 20)}...{tx.address.substring(tx.address.length - 8)}
                                            </p>
                                        )}
                                    </div>
                                </div>
                                <div className="text-right">
                                    <p className={`text-lg font-bold ${tx.type === TransactionType.ISSUE || tx.type === TransactionType.SETTLEMENT
                                            ? 'text-red-600'
                                            : 'text-green-600'
                                        }`}>
                                        {tx.type === TransactionType.ISSUE || tx.type === TransactionType.SETTLEMENT ? '-' : '+'}
                                        {tx.amount.toLocaleString()} LCN
                                    </p>
                                    <Badge variant={
                                        tx.status === 'CONFIRMED' || tx.status === 'COMPLETED' ? 'success' :
                                            tx.status === 'PENDING' ? 'warning' : 'danger'
                                    }>
                                        {tx.status}
                                    </Badge>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </Card>
        </div>
    );
};
