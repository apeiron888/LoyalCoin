// API Service for LoyalCoin Customer Portal
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const getToken = (): string | null => localStorage.getItem('customer_token');
export const setToken = (token: string): void => localStorage.setItem('customer_token', token);
export const clearToken = (): void => localStorage.removeItem('customer_token');

export class ApiError extends Error {
    code: string;
    status: number;
    constructor(code: string, message: string, status: number) {
        super(message);
        this.code = code;
        this.status = status;
    }
}

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = getToken();
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
    };

    const response = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });
    const data = await response.json();

    if (!response.ok) {
        throw new ApiError(data.code || 'UNKNOWN_ERROR', data.message || 'Request failed', response.status);
    }
    return data;
}

// Types
export interface User {
    id: string;
    email: string;
    username?: string;
    role: string;
    wallet_address: string;
}

export interface AuthResponse {
    status: string;
    data: {
        token: string;
        expires_at: string;
        user: User;
    };
}

export interface SignupResponse {
    status: string;
    data: {
        user_id: string;
        wallet_address: string;
        role: string;
    };
}

export interface Balance {
    address: string;
    ada: number;
    lovelace: number;
    lcn: number;
    lcn_atomic: number;
}

export interface BalanceResponse {
    status: string;
    data: Balance;
}

export interface Transaction {
    tx_hash: string;
    type: string;
    direction: 'sent' | 'received';
    from_address: string;
    to_address: string;
    amount_lcn: number;
    status: string;
    submitted_at: string;
    confirmed_at?: string;
}

export interface TransactionsResponse {
    status: string;
    data: {
        transactions: Transaction[];
        count: number;
        limit: number;
        offset: number;
    };
}

export interface RedeemResponse {
    status: string;
    data: {
        tx_hash: string;
        amount_lcn: number;
    };
}

// Auth APIs
export async function signup(email: string, password: string, username: string): Promise<SignupResponse> {
    return apiRequest<SignupResponse>('/api/v1/auth/signup', {
        method: 'POST',
        body: JSON.stringify({
            email,
            password,
            role: 'CUSTOMER',
            username,
        }),
    });
}

export async function login(email: string, password: string): Promise<AuthResponse> {
    return apiRequest<AuthResponse>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
}

// Wallet APIs
export async function getBalance(): Promise<BalanceResponse> {
    return apiRequest<BalanceResponse>('/api/v1/wallet/balance');
}

export async function getTransactions(limit = 20, offset = 0): Promise<TransactionsResponse> {
    return apiRequest<TransactionsResponse>(`/api/v1/wallet/transactions?limit=${limit}&offset=${offset}`);
}

// LCN APIs
export async function redeemLCN(merchantAddress: string, amountLCN: number): Promise<RedeemResponse> {
    return apiRequest<RedeemResponse>('/api/v1/lcn/redeem', {
        method: 'POST',
        body: JSON.stringify({
            merchant_address: merchantAddress,
            amount_lcn: amountLCN,
        }),
    });
}
