// API Service for LoyalCoin Merchant Portal
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Get stored token
const getToken = (): string | null => localStorage.getItem('lcn_token');

// Set stored token
export const setToken = (token: string): void => localStorage.setItem('lcn_token', token);

// Clear stored token
export const clearToken = (): void => localStorage.removeItem('lcn_token');

// Generic API request helper
async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const token = getToken();
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
    };

    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers,
    });

    const data = await response.json();

    if (!response.ok) {
        throw new ApiError(data.code || 'UNKNOWN_ERROR', data.message || 'Request failed', response.status);
    }

    return data;
}

// Custom error class for API errors
export class ApiError extends Error {
    code: string;
    status: number;

    constructor(code: string, message: string, status: number) {
        super(message);
        this.code = code;
        this.status = status;
        this.name = 'ApiError';
    }
}

// Response types
export interface AuthResponse {
    status: string;
    data: {
        token: string;
        expires_at: string;
        user: {
            id: string;
            email: string;
            role: string;
            wallet_address: string;
            business_name?: string;
        };
    };
}

export interface BalanceResponse {
    status: string;
    data: {
        address: string;
        ada: number;
        lovelace: number;
        lcn: number;
        lcn_atomic: number;
        other_assets: Record<string, number>;
    };
}

export interface Transaction {
    id: string;
    tx_hash: string;
    from_address: string;
    to_address: string;
    amount_lcn: number;
    type: string;
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

export interface Settlement {
    id: string;
    merchant_id: string;
    amount_lcn: number;
    amount_etb: number;
    status: string;
    bank_account: {
        account_number: string;
        bank_name: string;
        account_holder: string;
    };
    requested_at: string;
    processed_at?: string;
}

export interface SettlementHistoryResponse {
    status: string;
    data: {
        settlements: Settlement[] | null;
        total: number;
        limit: number;
        offset: number;
    };
}

export interface Allocation {
    id: string;
    merchant_id: string;
    amount_lcn: number;
    amount_etb_paid: number;
    payment_method: string;
    payment_reference: string;
    status: string;
    purchased_at: string;
    approved_at?: string;
}

export interface AllocationHistoryResponse {
    status: string;
    data: {
        allocations: Allocation[];
        total: number;
        limit: number;
        offset: number;
    };
}

// API functions

// Auth
export async function signup(
    email: string,
    password: string,
    businessName: string
): Promise<AuthResponse> {
    return apiRequest<AuthResponse>('/api/v1/auth/signup', {
        method: 'POST',
        body: JSON.stringify({
            email,
            password,
            role: 'MERCHANT',
            business_name: businessName,
        }),
    });
}

export async function login(email: string, password: string): Promise<AuthResponse> {
    return apiRequest<AuthResponse>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
}

// Wallet
export async function getBalance(): Promise<BalanceResponse> {
    return apiRequest<BalanceResponse>('/api/v1/wallet/balance');
}

export async function getTransactions(limit = 20, offset = 0): Promise<TransactionsResponse> {
    return apiRequest<TransactionsResponse>(
        `/api/v1/wallet/transactions?limit=${limit}&offset=${offset}`
    );
}

// LCN Operations
export async function issueLCN(
    toAddress: string,
    amount: number,
    note?: string
): Promise<{ status: string; data: { tx_hash: string; message: string } }> {
    return apiRequest('/api/v1/lcn/issue', {
        method: 'POST',
        body: JSON.stringify({
            customer_address: toAddress,
            amount_lcn: amount,
            reference: note,
        }),
    });
}

// Settlement
export async function requestSettlement(
    amountLCN: number,
    bankAccount: { account_number: string; bank_name: string; account_holder: string }
): Promise<{ status: string; data: { settlement_id: string; message: string } }> {
    return apiRequest('/api/v1/merchant/settlement/request', {
        method: 'POST',
        body: JSON.stringify({
            amount_lcn: amountLCN,
            bank_account: bankAccount,
        }),
    });
}

export async function getSettlementHistory(
    limit = 20,
    offset = 0
): Promise<SettlementHistoryResponse> {
    return apiRequest<SettlementHistoryResponse>(
        `/api/v1/merchant/settlement/history?limit=${limit}&offset=${offset}`
    );
}

// Allocation (Buy LCN)
// Note: amountLCN should be whole LCN units
export async function purchaseAllocation(
    amountLCN: number,
    paymentMethod: string,
    paymentReference: string,
    paymentProofUrl?: string
): Promise<{ status: string; data: { purchase_id: string; amount_lcn: number; amount_etb: number; status: string; message: string } }> {
    return apiRequest('/api/v1/merchant/allocation/purchase', {
        method: 'POST',
        body: JSON.stringify({
            amount_lcn: amountLCN,
            payment_method: paymentMethod,
            payment_reference: paymentReference,
            payment_proof_url: paymentProofUrl,
        }),
    });
}

export async function getAllocationHistory(
    limit = 20,
    offset = 0
): Promise<AllocationHistoryResponse> {
    return apiRequest<AllocationHistoryResponse>(
        `/api/v1/merchant/allocation/history?limit=${limit}&offset=${offset}`
    );
}
