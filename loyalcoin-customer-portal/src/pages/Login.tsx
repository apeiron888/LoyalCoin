import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff } from 'lucide-react';
import { login } from '../services/api';
import { useStore, saveUser } from '../store';

export const Login: React.FC = () => {
    const navigate = useNavigate();
    const { setUser } = useStore();

    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setLoading(true);

        try {
            const response = await login(email, password);
            const { token, user } = response.data;

            // Check if user is a customer
            if (user.role !== 'CUSTOMER') {
                setError('This app is for customers only. Please use the merchant portal.');
                setLoading(false);
                return;
            }

            setUser(user, token);
            saveUser(user);
            navigate('/');
        } catch (err: any) {
            setError(err.message || 'Login failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <div className="auth-header">
                <div className="auth-logo">🪙</div>
                <h1 className="auth-title">LoyalCoin</h1>
                <p className="auth-subtitle">Your loyalty points wallet</p>
            </div>

            <form className="auth-form" onSubmit={handleSubmit}>
                {error && (
                    <div className="alert alert-error">{error}</div>
                )}

                <div className="form-group">
                    <label className="form-label">Email</label>
                    <input
                        type="email"
                        className="form-input"
                        placeholder="Enter your email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                    />
                </div>

                <div className="form-group">
                    <label className="form-label">Password</label>
                    <div style={{ position: 'relative' }}>
                        <input
                            type={showPassword ? 'text' : 'password'}
                            className="form-input"
                            placeholder="Enter your password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                            style={{ paddingRight: '3rem' }}
                        />
                        <button
                            type="button"
                            onClick={() => setShowPassword(!showPassword)}
                            style={{
                                position: 'absolute',
                                right: '0.75rem',
                                top: '50%',
                                transform: 'translateY(-50%)',
                                background: 'none',
                                border: 'none',
                                color: 'var(--text-muted)',
                                cursor: 'pointer',
                            }}
                        >
                            {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                        </button>
                    </div>
                </div>

                <button
                    type="submit"
                    className="btn btn-primary btn-block mt-2"
                    disabled={loading}
                >
                    {loading ? <span className="spinner" /> : 'Sign In'}
                </button>
            </form>

            <div className="auth-footer">
                Don't have an account? <Link to="/signup">Sign up</Link>
            </div>
        </div>
    );
};
