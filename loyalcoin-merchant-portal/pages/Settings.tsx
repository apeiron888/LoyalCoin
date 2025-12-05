import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../store';
import { Card, Button, Input, Badge } from '../components/UIComponents';
import { ArrowLeft, User, Building, Wallet, Plus, Trash2, CheckCircle } from 'lucide-react';

export const Settings: React.FC = () => {
    const navigate = useNavigate();
    const { user, wallet, addBankAccount, logout } = useStore();
    const [showAddBank, setShowAddBank] = useState(false);
    const [bankName, setBankName] = useState('');
    const [accountNumber, setAccountNumber] = useState('');
    const [accountHolder, setAccountHolder] = useState('');

    const handleAddBank = (e: React.FormEvent) => {
        e.preventDefault();
        addBankAccount({
            id: Date.now().toString(),
            bankName,
            accountNumber,
            accountHolder,
            isPrimary: wallet.bankAccounts.length === 0
        });
        setBankName('');
        setAccountNumber('');
        setAccountHolder('');
        setShowAddBank(false);
    };

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="max-w-2xl mx-auto">
            <div className="mb-6 flex items-center">
                <Button variant="ghost" size="sm" onClick={() => navigate('/')} className="mr-4">
                    <ArrowLeft className="h-4 w-4 mr-1" /> Back
                </Button>
                <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
            </div>

            <div className="space-y-6">
                {/* Profile Section */}
                <Card className="p-6">
                    <div className="flex items-center gap-3 mb-4">
                        <User className="h-5 w-5 text-amber-600" />
                        <h2 className="text-lg font-bold text-gray-900">Profile</h2>
                    </div>

                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-500 mb-1">Business Name</label>
                            <p className="text-gray-900 font-medium">{user?.businessName}</p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-500 mb-1">Email</label>
                            <p className="text-gray-900">{user?.email}</p>
                        </div>
                    </div>
                </Card>

                {/* Wallet Section */}
                <Card className="p-6">
                    <div className="flex items-center gap-3 mb-4">
                        <Wallet className="h-5 w-5 text-amber-600" />
                        <h2 className="text-lg font-bold text-gray-900">Wallet</h2>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-500 mb-1">Wallet Address</label>
                        <div className="p-3 bg-gray-50 rounded-lg font-mono text-sm text-gray-700 break-all">
                            {user?.walletAddress || 'Not available'}
                        </div>
                    </div>
                </Card>

                {/* Bank Accounts Section */}
                <Card className="p-6">
                    <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-3">
                            <Building className="h-5 w-5 text-amber-600" />
                            <h2 className="text-lg font-bold text-gray-900">Bank Accounts</h2>
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setShowAddBank(!showAddBank)}
                        >
                            <Plus className="h-4 w-4 mr-1" />
                            Add Account
                        </Button>
                    </div>

                    {showAddBank && (
                        <form onSubmit={handleAddBank} className="mb-4 p-4 bg-gray-50 rounded-lg space-y-4">
                            <Input
                                label="Bank Name"
                                placeholder="e.g., Commercial Bank of Ethiopia"
                                value={bankName}
                                onChange={(e) => setBankName(e.target.value)}
                                required
                            />
                            <Input
                                label="Account Number"
                                placeholder="Your account number"
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
                            <div className="flex gap-2">
                                <Button type="submit" size="sm">Save Account</Button>
                                <Button type="button" variant="ghost" size="sm" onClick={() => setShowAddBank(false)}>Cancel</Button>
                            </div>
                        </form>
                    )}

                    {wallet.bankAccounts.length === 0 ? (
                        <div className="text-center py-6 text-gray-500">
                            <Building className="mx-auto h-8 w-8 text-gray-300 mb-2" />
                            <p>No bank accounts added yet</p>
                            <p className="text-sm text-gray-400">Add a bank account for settlements</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {wallet.bankAccounts.map((account) => (
                                <div key={account.id} className="flex items-center justify-between p-4 border border-gray-100 rounded-lg">
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <p className="font-medium text-gray-900">{account.bankName}</p>
                                            {account.isPrimary && <Badge variant="success">Primary</Badge>}
                                        </div>
                                        <p className="text-sm text-gray-500">{account.accountNumber} • {account.accountHolder}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </Card>

                {/* Sign Out */}
                <Card className="p-6">
                    <Button
                        variant="danger"
                        onClick={handleLogout}
                        className="w-full"
                    >
                        Sign Out
                    </Button>
                </Card>
            </div>
        </div>
    );
};