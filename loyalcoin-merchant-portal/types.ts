export enum TransactionType {
  ISSUE = 'ISSUE',
  RECEIVE = 'RECEIVE',
  SETTLEMENT = 'SETTLEMENT',
  ALLOCATION = 'ALLOCATION'
}

export enum TransactionStatus {
  CONFIRMED = 'CONFIRMED',
  PENDING = 'PENDING',
  FAILED = 'FAILED',
  COMPLETED = 'COMPLETED'
}

export interface Transaction {
  id: string;
  date: string; // ISO string
  type: TransactionType;
  amount: number; // Amount in LCN
  address: string;
  status: TransactionStatus;
  description?: string;
  txHash?: string;
}

export interface User {
  id: string;
  email: string;
  businessName: string;
  phone?: string;
  walletAddress: string;
}

export interface BankAccount {
  id: string;
  bankName: string;
  accountNumber: string;
  accountHolder: string;
  isPrimary: boolean;
}

export interface WalletState {
  balanceLCN: number; // Only LCN balance, no ADA display
  transactions: Transaction[];
  bankAccounts: BankAccount[];
}