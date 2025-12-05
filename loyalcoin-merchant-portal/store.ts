import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Transaction, TransactionStatus, TransactionType, User, WalletState, BankAccount } from './types';
import * as api from './services/api';

// Exchange rate: 1 ADA = 10 LCN, 1 ETB = 10 LCN
const ADA_TO_LCN_RATE = 10;

interface AppState {
  user: User | null;
  token: string | null;
  wallet: WalletState;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Auth actions
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string, businessName: string) => Promise<void>;
  logout: () => void;

  // Wallet actions
  fetchBalance: () => Promise<void>;
  fetchTransactions: () => Promise<void>;

  // Local state updates
  addBankAccount: (account: BankAccount) => void;
  clearError: () => void;
}

const INITIAL_WALLET: WalletState = {
  balanceLCN: 0,
  transactions: [],
  bankAccounts: []
};

export const useStore = create<AppState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      wallet: INITIAL_WALLET,

      login: async (email: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await api.login(email, password);
          const { token, user } = response.data;

          api.setToken(token);

          set({
            isAuthenticated: true,
            token,
            user: {
              id: user.id,
              email: user.email,
              businessName: user.business_name || 'Merchant',
              walletAddress: user.wallet_address
            },
            isLoading: false,
            error: null
          });

          // Fetch wallet data after login
          await get().fetchBalance();
          await get().fetchTransactions();
        } catch (err) {
          const error = err as api.ApiError;
          set({
            isLoading: false,
            error: error.message || 'Login failed'
          });
          throw error;
        }
      },

      signup: async (email: string, password: string, businessName: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await api.signup(email, password, businessName);
          const { token, user } = response.data;

          api.setToken(token);

          set({
            isAuthenticated: true,
            token,
            user: {
              id: user.id,
              email: user.email,
              businessName: user.business_name || businessName,
              walletAddress: user.wallet_address
            },
            isLoading: false,
            error: null
          });

          // Fetch wallet data after signup
          await get().fetchBalance();
          await get().fetchTransactions();
        } catch (err) {
          const error = err as api.ApiError;
          set({
            isLoading: false,
            error: error.message || 'Signup failed'
          });
          throw error;
        }
      },

      logout: () => {
        api.clearToken();
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          wallet: INITIAL_WALLET,
          error: null
        });
      },

      fetchBalance: async () => {
        try {
          const response = await api.getBalance();
          // Use LCN balance directly from backend (lcn is in LCN units, lcn_atomic is in milli-LCN)
          const lcnBalance = response.data.lcn || (response.data.lcn_atomic / 1000);
          set((state) => ({
            wallet: {
              ...state.wallet,
              balanceLCN: lcnBalance
            }
          }));
        } catch (err) {
          console.error('Failed to fetch balance:', err);
        }
      },

      fetchTransactions: async () => {
        try {
          const response = await api.getTransactions(50, 0);
          const transactions: Transaction[] = (response.data.transactions || []).map((tx) => ({
            id: tx.id,
            date: tx.submitted_at,
            type: tx.type === 'ISSUANCE' ? TransactionType.ISSUE :
              tx.type === 'REDEMPTION' ? TransactionType.RECEIVE :
                tx.type === 'SETTLEMENT' ? TransactionType.SETTLEMENT :
                  TransactionType.ALLOCATION,
            // Convert from atomic units and then to LCN if needed
            amount: tx.amount_lcn / 1000,
            address: tx.to_address || tx.from_address,
            status: tx.status === 'CONFIRMED' ? TransactionStatus.CONFIRMED :
              tx.status === 'PENDING' ? TransactionStatus.PENDING :
                tx.status === 'FAILED' ? TransactionStatus.FAILED :
                  TransactionStatus.COMPLETED,
            txHash: tx.tx_hash
          }));

          set((state) => ({
            wallet: {
              ...state.wallet,
              transactions
            }
          }));
        } catch (err) {
          console.error('Failed to fetch transactions:', err);
        }
      },

      addBankAccount: (account: BankAccount) => set((state) => ({
        wallet: {
          ...state.wallet,
          bankAccounts: [...state.wallet.bankAccounts, account]
        }
      })),

      clearError: () => set({ error: null }),
    }),
    {
      name: 'loyalcoin-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        wallet: {
          ...state.wallet,
          transactions: []
        }
      }),
    }
  )
);

// Apply token on load
const state = useStore.getState();
if (state.token) {
  api.setToken(state.token);
}