import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import * as api from './services/api';

interface User {
    id: string;
    email: string;
    role: string;
}

interface AppState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    isLoading: boolean;
    error: string | null;
    isDarkMode: boolean;

    // Actions
    login: (email: string, password: string) => Promise<void>;
    logout: () => void;
    clearError: () => void;
    toggleTheme: () => void;
}

export const useStore = create<AppState>()(
    persist(
        (set, get) => ({
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
            isDarkMode: false,

            login: async (email: string, password: string) => {
                set({ isLoading: true, error: null });
                try {
                    const response = await api.login(email, password);
                    const { token, user } = response.data;

                    // Verify this is an admin
                    if (user.role !== 'ADMIN') {
                        throw new api.ApiError('403_FORBIDDEN', 'Access denied. Admin account required.', 403);
                    }

                    api.setToken(token);
                    set({
                        isAuthenticated: true,
                        token,
                        user: { id: user.id, email: user.email, role: user.role },
                        isLoading: false,
                        error: null,
                    });
                } catch (err) {
                    const error = err as api.ApiError;
                    set({ isLoading: false, error: error.message || 'Login failed' });
                    throw error;
                }
            },

            logout: () => {
                api.clearToken();
                set({ isAuthenticated: false, user: null, token: null, error: null });
            },

            clearError: () => set({ error: null }),

            toggleTheme: () => {
                const newMode = !get().isDarkMode;
                set({ isDarkMode: newMode });
                document.body.classList.toggle('dark', newMode);
                document.body.classList.toggle('light', !newMode);
            },
        }),
        {
            name: 'loyalcoin-admin-storage',
            partialize: (state) => ({
                token: state.token,
                user: state.user,
                isAuthenticated: state.isAuthenticated,
                isDarkMode: state.isDarkMode,
            }),
        }
    )
);

// Apply theme and token on load
const state = useStore.getState();
if (state.token) {
    api.setToken(state.token);
}
if (state.isDarkMode) {
    document.body.classList.add('dark');
} else {
    document.body.classList.add('light');
}
