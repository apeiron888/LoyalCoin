import { create } from 'zustand';
import { getBalance, getTransactions, User, Balance, Transaction, setToken, clearToken } from './services/api';

interface AppState {
    // Auth
    user: User | null;
    isAuthenticated: boolean;

    // Wallet
    balance: Balance | null;
    transactions: Transaction[];

    // Loading states
    isLoading: boolean;
    error: string | null;

    // Actions
    setUser: (user: User | null, token?: string) => void;
    logout: () => void;
    fetchBalance: () => Promise<void>;
    fetchTransactions: () => Promise<void>;
    clearError: () => void;
}

export const useStore = create<AppState>((set) => ({
    // Initial state
    user: null,
    isAuthenticated: false,
    balance: null,
    transactions: [],
    isLoading: false,
    error: null,

    // Set user after login/signup
    setUser: (user, token) => {
        if (token) {
            setToken(token);
        }
        set({ user, isAuthenticated: !!user });
    },

    // Logout
    logout: () => {
        clearToken();
        set({
            user: null,
            isAuthenticated: false,
            balance: null,
            transactions: [],
        });
    },

    // Fetch wallet balance
    fetchBalance: async () => {
        try {
            set({ isLoading: true, error: null });
            const response = await getBalance();
            set({ balance: response.data, isLoading: false });
        } catch (err: any) {
            set({ error: err.message, isLoading: false });
        }
    },

    // Fetch transactions
    fetchTransactions: async () => {
        try {
            set({ isLoading: true, error: null });
            const response = await getTransactions(50, 0);
            set({ transactions: response.data.transactions || [], isLoading: false });
        } catch (err: any) {
            set({ error: err.message, isLoading: false });
        }
    },

    // Clear error
    clearError: () => set({ error: null }),
}));

// Initialize from localStorage on app load
export function initializeAuth(): boolean {
    const token = localStorage.getItem('customer_token');
    const userStr = localStorage.getItem('customer_user');

    if (token && userStr) {
        try {
            const user = JSON.parse(userStr);
            useStore.getState().setUser(user);
            return true;
        } catch {
            clearToken();
            localStorage.removeItem('customer_user');
        }
    }
    return false;
}

// Save user to localStorage
export function saveUser(user: User): void {
    localStorage.setItem('customer_user', JSON.stringify(user));
}
