import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../components/Toast';
import { LogIn, Mail, Lock, Sparkles, ArrowRight } from 'lucide-react';

export default function Login({ navigate }) {
  const { login } = useAuth();
  const toast = useToast();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (!email.trim() || !password) {
      setError('Please fill in all fields.');
      return;
    }

    setLoading(true);
    try {
      await login(email.trim(), password);
      toast.success('Welcome back!');
      navigate('/dashboard');
    } catch (err) {
      setError(err.message || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  const fillDemo = () => {
    setEmail('organizer@example.com');
    setPassword('secret123');
  };

  return (
    <div className="app-container" style={{ maxWidth: '460px', marginTop: '40px' }}>
      <div className="glass-panel" style={{ padding: '36px 32px' }}>
        <div style={{ textAlign: 'center', marginBottom: '28px' }}>
          <div
            style={{
              width: '52px',
              height: '52px',
              borderRadius: '14px',
              background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%)',
              border: '1px solid rgba(59, 130, 246, 0.3)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginBottom: '16px',
            }}
          >
            <LogIn size={26} color="#3b82f6" />
          </div>
          <h2 style={{ fontSize: '1.6rem', marginBottom: '6px' }}>Sign in to PulsePoll</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
            Create, manage, and broadcast your live audience polls.
          </p>
        </div>

        {error && (
          <div
            style={{
              padding: '12px 16px',
              borderRadius: 'var(--radius-md)',
              background: 'rgba(244, 63, 94, 0.12)',
              border: '1px solid rgba(244, 63, 94, 0.3)',
              color: '#fb7185',
              fontSize: '0.875rem',
              marginBottom: '20px',
            }}
          >
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="login-email">Email Address</label>
            <div style={{ position: 'relative' }}>
              <Mail
                size={18}
                color="var(--text-muted)"
                style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)' }}
              />
              <input
                id="login-email"
                type="email"
                className="form-input"
                style={{ paddingLeft: '42px' }}
                placeholder="you@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="login-password">Password</label>
            <div style={{ position: 'relative' }}>
              <Lock
                size={18}
                color="var(--text-muted)"
                style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)' }}
              />
              <input
                id="login-password"
                type="password"
                className="form-input"
                style={{ paddingLeft: '42px' }}
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-lg"
            style={{ width: '100%', marginTop: '8px' }}
            disabled={loading}
          >
            {loading ? 'Authenticating...' : 'Sign In'}
            <ArrowRight size={18} />
          </button>
        </form>

        <div style={{ marginTop: '20px', textAlign: 'center' }}>
          <button
            type="button"
            onClick={fillDemo}
            className="btn btn-sm btn-secondary"
            style={{ width: '100%', borderStyle: 'dashed' }}
          >
            <Sparkles size={14} color="#8b5cf6" />
            <span>Fill Demo Credentials</span>
          </button>
        </div>

        <div
          style={{
            marginTop: '28px',
            paddingTop: '20px',
            borderTop: '1px solid var(--border-subtle)',
            textAlign: 'center',
            fontSize: '0.9rem',
            color: 'var(--text-secondary)',
          }}
        >
          Don't have an account?{' '}
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              navigate('/signup');
            }}
            style={{ color: 'var(--accent-blue)', fontWeight: 600 }}
          >
            Sign up now
          </a>
        </div>
      </div>
    </div>
  );
}
