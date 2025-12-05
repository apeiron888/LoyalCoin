import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff, CheckCircle } from 'lucide-react';
import { signup, login } from '../services/api';
import { useStore, saveUser } from '../store';

export const Signup: React.FC = () => {
    const navigate = useNavigate();
    const { setUser } = useStore();

    const [username, setUsername] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setLoading(true);

        try {
            // Create account
            await signup(email, password, username);
            setSuccess(true);

            // Auto-login after signup
            setTimeout(async () => {
                try {
                    const loginResponse = await login(email, password);
                    const { token, user } = loginResponse.data;
                    setUser(user, token);
                    saveUser(user);
                    navigate('/');
                } catch {
                    // If auto-login fails, redirect to login page
                    navigate('/login');
                }
            }, 1500);
        } catch (err: any) {
            setError(err.message || 'Signup failed. Please try again.');
            setLoading(false);
        }
    };

    if (success) {
        return (
            <div className="auth-page" style={{ justifyContent: 'center' }}>
                <div className="text-center">
                    <div style={{
                        width: '80px',
                        height: '80px',
                        borderRadius: '50%',
                        background: 'rgba(16, 185, 129, 0.2)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        margin: '0 auto 1.5rem'
                    }}>
                        <CheckCircle size={40} color="#10B981" />
                    </div>
                    <h2 style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>Account Created!</h2>
                    <p className="text-muted">Your wallet is being set up...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="auth-page">
            <div className="auth-header">
                <div className="auth-logo">🪙</div>
                <h1 className="auth-title">Join LoyalCoin</h1>
                <p className="auth-subtitle">Start collecting loyalty points</p>
            </div>

            <form className="auth-form" onSubmit={handleSubmit}>
                {error && (
                    <div className="alert alert-error">{error}</div>
                )}

                <div className="form-group">
                    <label className="form-label">Username</label>
                    <input
                        type="text"
                        className="form-input"
                        placeholder="Choose a username"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        minLength={3}
                    />
                </div>

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
                            placeholder="Create a password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                            minLength={8}
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
                    <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.5rem' }}>
                        At least 8 characters with uppercase, lowercase, and a number
                    </p>
                </div>

                <button
                    type="submit"
                    className="btn btn-primary btn-block mt-2"
                    disabled={loading}
                >
                    {loading ? <span className="spinner" /> : 'Create Account'}
                </button>
            </form>

            <div className="auth-footer">
                Already have an account? <Link to="/login">Sign in</Link>
            </div>
        </div>
    );
};
