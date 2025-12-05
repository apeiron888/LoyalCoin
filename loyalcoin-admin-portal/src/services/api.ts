// API Service for LoyalCoin Admin Portal
const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const getToken = (): string | null => localStorage.getItem('admin_token');
export const setToken = (token: string): void => localStorage.setItem('admin_token', token);
export const clearToken = (): void => localStorage.removeItem('admin_token');

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

// Types - matching the backend models exactly
export interface AuthResponse {
    status: string;
    data: {
        token: string;
        expires_at: string;
        user: { id: string; email: string; role: string; wallet_address: string };
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
}

export interface Settlement {
    id: string;
    merchant_id: string;
    amount_lcn: number;
    amount_etb: number;
    status: string;
    bank_account: { account_number: string; bank_name: string; account_holder: string };
    requested_at: string;
}

export interface ReserveStatus {
    lcn_balance: number;
    lcn_balance_atomic: number;
    governance_wallet_address: string;
    health: string;
}

// Auth
export async function login(email: string, password: string): Promise<AuthResponse> {
    return apiRequest<AuthResponse>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
}

// Admin Endpoints
export async function getReserveStatus(): Promise<{ status: string; data: ReserveStatus }> {
    return apiRequest('/api/v1/admin/reserve/status');
}

export async function getPendingAllocations(limit = 50, offset = 0): Promise<{
    status: string;
    data: { allocations: Allocation[]; total: number };
}> {
    return apiRequest(`/api/v1/admin/allocation/pending?limit=${limit}&offset=${offset}`);
}

export async function getPendingSettlements(limit = 50, offset = 0): Promise<{
    status: string;
    data: { settlements: Settlement[]; total: number };
}> {
    return apiRequest(`/api/v1/admin/settlement/pending?limit=${limit}&offset=${offset}`);
}

export async function approveAllocation(purchaseId: string, action: 'APPROVE' | 'REJECT', notes?: string): Promise<{
    status: string;
    data: { purchase_id: string; status: string };
}> {
    console.log('Calling approveAllocation with:', { purchaseId, action, notes });
    return apiRequest('/api/v1/admin/allocation/approve', {
        method: 'POST',
        body: JSON.stringify({ purchase_id: purchaseId, action, notes: notes || '' }),
    });
}

export async function approveSettlement(settlementId: string, action: 'APPROVE' | 'REJECT', paymentReference?: string, notes?: string): Promise<{
    status: string;
    data: { settlement_id: string; status: string };
}> {
    console.log('Calling approveSettlement with:', { settlementId, action, paymentReference, notes });
    return apiRequest('/api/v1/admin/settlement/approve', {
        method: 'POST',
        body: JSON.stringify({
            settlement_id: settlementId,
            action,
            payment_reference: paymentReference || `BANK-${Date.now()}`,
            notes: notes || ''
        }),
    });
}
