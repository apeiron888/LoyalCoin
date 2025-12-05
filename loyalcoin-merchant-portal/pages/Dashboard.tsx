import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  TrendingUp,
  TrendingDown,
  ArrowRight,
  Coins,
  CreditCard,
  Plus,
  RefreshCw,
  QrCode
} from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import { useStore } from '../store';
import { Card, Button, Badge } from '../components/UIComponents';
import { TransactionType } from '../types';

// Exchange rate: 1 ETB = 10 LCN
const LCN_TO_ETB_RATE = 10;

interface StatsCardProps {
  title: string;
  value: string;
  subValue?: string;
  trend?: 'up' | 'down';
  trendValue?: string;
  icon: React.ReactNode;
  color: string;
}

const StatsCard: React.FC<StatsCardProps> = ({ title, value, subValue, trend, trendValue, icon, color }) => {
  return (
    <Card className="p-6">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <h3 className="mt-2 text-3xl font-bold text-gray-900">{value}</h3>
          {subValue && <p className="mt-1 text-sm text-gray-500">{subValue}</p>}
        </div>
        <div className={`rounded-xl p-3 ${color}`}>
          {icon}
        </div>
      </div>
      {trendValue && (
        <div className="mt-4 flex items-center text-sm">
          {trend === 'up' ? (
            <TrendingUp className="mr-1 h-4 w-4 text-green-500" />
          ) : (
            <TrendingDown className="mr-1 h-4 w-4 text-red-500" />
          )}
          <span className={trend === 'up' ? 'text-green-600' : 'text-red-600'}>
            {trendValue}
          </span>
          <span className="ml-1 text-gray-500">vs last month</span>
        </div>
      )}
    </Card>
  );
};

const getLast7Days = () => {
  const days = [];
  for (let i = 6; i >= 0; i--) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    days.push(d);
  }
  return days;
};

// Convert LCN to ETB (1 ETB = 10 LCN, so LCN / 10 = ETB)
const lcnToEtb = (lcn: number): number => lcn / LCN_TO_ETB_RATE;

export const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const { wallet, user, fetchBalance, fetchTransactions } = useStore();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const recentTx = wallet.transactions.slice(0, 5);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([fetchBalance(), fetchTransactions()]);
    } finally {
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    handleRefresh();
  }, []);

  const thisMonthIssued = wallet.transactions
    .filter(tx => {
      const txDate = new Date(tx.date);
      const now = new Date();
      return tx.type === TransactionType.ISSUE &&
        txDate.getMonth() === now.getMonth() &&
        txDate.getFullYear() === now.getFullYear();
    })
    .reduce((sum, tx) => sum + tx.amount, 0);

  const last7Days = getLast7Days();
  const chartData = last7Days.map(day => {
    const dayName = day.toLocaleDateString('en-US', { weekday: 'short' });
    const dateStr = day.toISOString().split('T')[0];

    const issuedAmount = wallet.transactions
      .filter(tx => {
        const txDate = new Date(tx.date).toISOString().split('T')[0];
        return tx.type === TransactionType.ISSUE && txDate === dateStr;
      })
      .reduce((sum, tx) => sum + tx.amount, 0);

    return { name: dayName, issued: issuedAmount };
  });

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-500">Welcome back, {user?.businessName}! Here's what's happening today.</p>
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

      {/* Quick Stats - Only LCN, with ETB equivalent */}
      <div className="grid gap-6 md:grid-cols-2">
        <StatsCard
          title="LCN Balance"
          value={`${wallet.balanceLCN.toLocaleString()} LCN`}
          subValue={`≈ ${lcnToEtb(wallet.balanceLCN).toLocaleString()} ETB`}
          icon={<Coins className="text-amber-600" size={24} />}
          color="bg-amber-50"
        />
        <StatsCard
          title="Issued This Month"
          value={`${thisMonthIssued.toLocaleString()} LCN`}
          subValue={`${wallet.transactions.filter(tx => tx.type === TransactionType.ISSUE).length} rewards issued`}
          icon={<TrendingUp className="text-emerald-600" size={24} />}
          color="bg-emerald-50"
        />
      </div>

      {/* Quick Actions */}
      <div className="grid gap-4 md:grid-cols-4">
        <Button
          variant="primary"
          size="lg"
          onClick={() => navigate('/issue')}
          className="w-full shadow-sm"
        >
          <Coins className="mr-2 h-5 w-5" />
          Issue Rewards
        </Button>
        <Button
          variant="secondary"
          size="lg"
          onClick={() => navigate('/receive')}
          className="w-full shadow-sm"
        >
          <QrCode className="mr-2 h-5 w-5" />
          Receive Coins
        </Button>
        <Button
          variant="secondary"
          size="lg"
          onClick={() => navigate('/buy')}
          className="w-full shadow-sm"
        >
          <Plus className="mr-2 h-5 w-5" />
          Buy LCN
        </Button>
        <Button
          variant="outline"
          size="lg"
          onClick={() => navigate('/cash-out')}
          className="w-full shadow-sm"
        >
          <CreditCard className="mr-2 h-5 w-5" />
          Cash Out
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Chart */}
        <Card className="p-6">
          <h3 className="text-lg font-bold mb-6 text-gray-900">Issuance Overview (LCN)</h3>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                <XAxis
                  dataKey="name"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#6b7280', fontSize: 12 }}
                  dy={10}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#6b7280', fontSize: 12 }}
                  dx={-10}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#ffffff',
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px',
                    color: '#000000',
                    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)'
                  }}
                  formatter={(value: number) => [`${value} LCN`, 'Issued']}
                />
                <Bar dataKey="issued" fill="#f59e0b" radius={[8, 8, 0, 0]} barSize={32} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Recent Transactions - All amounts in LCN */}
        <Card className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-bold text-gray-900">Recent Transactions</h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/transactions')}
            >
              View All
              <ArrowRight className="ml-1 h-4 w-4" />
            </Button>
          </div>

          {recentTx.length === 0 ? (
            <div className="p-8 text-center">
              <Coins className="mx-auto h-12 w-12 mb-3 text-gray-300" />
              <p className="text-gray-500">No transactions yet</p>
              <p className="text-sm mt-1 text-gray-400">Issue rewards to customers to get started</p>
            </div>
          ) : (
            <div className="space-y-4">
              {recentTx.map((tx) => (
                <div key={tx.id} className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors">
                  <div className="flex items-center gap-3">
                    <div className={`flex h-10 w-10 items-center justify-center rounded-full ${tx.type === TransactionType.ISSUE ? 'bg-orange-100 text-orange-600' :
                      tx.type === TransactionType.RECEIVE ? 'bg-green-100 text-green-600' :
                        'bg-gray-100 text-gray-600'
                      }`}>
                      {tx.type === TransactionType.ISSUE ? <TrendingUp size={18} /> :
                        tx.type === TransactionType.RECEIVE ? <TrendingDown size={18} /> :
                          <CreditCard size={18} />}
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">
                        {tx.type === TransactionType.ISSUE ? 'Issued Reward' :
                          tx.type === TransactionType.RECEIVE ? 'Received LCN' :
                            tx.type === TransactionType.SETTLEMENT ? 'Settlement' : 'Purchase'}
                      </p>
                      <p className="text-xs text-gray-500">
                        {new Date(tx.date).toLocaleDateString()} • {tx.address?.substring(0, 12)}...
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={`font-medium ${tx.type === TransactionType.ISSUE || tx.type === TransactionType.SETTLEMENT
                      ? 'text-red-600'
                      : 'text-green-600'
                      }`}>
                      {tx.type === TransactionType.ISSUE || tx.type === TransactionType.SETTLEMENT ? '-' : '+'}
                      {tx.amount.toLocaleString()} LCN
                    </p>
                    <Badge variant={tx.status === 'CONFIRMED' || tx.status === 'COMPLETED' ? 'success' : 'warning'}>
                      {tx.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};