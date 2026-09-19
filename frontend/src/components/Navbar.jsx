import React from 'react';
import { useAuth } from '../hooks/useAuth';
import { BarChart3, PlusCircle, LayoutDashboard, LogOut, LogIn, UserPlus } from 'lucide-react';

export default function Navbar({ currentPath, navigate }) {
  const { user, logout } = useAuth();

  return (
    <header className="navbar">
      <div className="navbar-container">
        <a
          href="#"
          onClick={(e) => {
            e.preventDefault();
            navigate(user ? '/dashboard' : '/login');
          }}
          className="brand-logo"
        >
          <BarChart3 size={24} color="#3b82f6" />
          <span>PulsePoll</span>
          <span className="live-badge" style={{ fontSize: '0.65rem', padding: '2px 8px' }}>
            <span className="pulse-dot"></span> LIVE
          </span>
        </a>

        <nav style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
          {user ? (
            <>
              <button
                onClick={() => navigate('/dashboard')}
                className={`btn btn-sm ${currentPath === '/dashboard' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <LayoutDashboard size={16} />
                <span>Dashboard</span>
              </button>
              <button
                onClick={() => navigate('/create-poll')}
                className={`btn btn-sm ${currentPath === '/create-poll' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <PlusCircle size={16} />
                <span>Create Poll</span>
              </button>

              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  paddingLeft: '12px',
                  borderLeft: '1px solid var(--border-subtle)',
                }}
              >
                <span
                  style={{
                    fontSize: '0.85rem',
                    color: 'var(--text-secondary)',
                    maxWidth: '140px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {user.name || user.email}
                </span>
                <button
                  onClick={() => {
                    logout();
                    navigate('/login');
                  }}
                  className="btn btn-sm btn-secondary"
                  title="Log out"
                  style={{ padding: '6px 8px' }}
                >
                  <LogOut size={15} />
                </button>
              </div>
            </>
          ) : (
            <>
              <button
                onClick={() => navigate('/login')}
                className={`btn btn-sm ${currentPath === '/login' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <LogIn size={16} />
                <span>Log In</span>
              </button>
              <button
                onClick={() => navigate('/signup')}
                className={`btn btn-sm ${currentPath === '/signup' ? 'btn-primary' : 'btn-secondary'}`}
              >
                <UserPlus size={16} />
                <span>Sign Up</span>
              </button>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
