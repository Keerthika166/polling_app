import React, { useState, useEffect } from 'react';
import { AuthProvider, useAuth } from './hooks/useAuth';
import { ToastProvider } from './components/Toast';
import Navbar from './components/Navbar';
import Login from './pages/Login';
import Signup from './pages/Signup';
import Dashboard from './pages/Dashboard';
import CreatePoll from './pages/CreatePoll';
import PublicPoll from './pages/PublicPoll';
import PollResults from './pages/PollResults';

function MainApp() {
  const { user, loading } = useAuth();
  const [currentPath, setCurrentPath] = useState(window.location.pathname);

  // Sync with browser back/forward buttons
  useEffect(() => {
    const handlePopState = () => {
      setCurrentPath(window.location.pathname);
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  const navigate = (path) => {
    window.history.pushState({}, '', path);
    setCurrentPath(path);
    window.scrollTo(0, 0);
  };

  if (loading) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexDirection: 'column',
          gap: '16px',
        }}
      >
        <div className="pulse-dot" style={{ width: '20px', height: '20px' }} />
        <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Initializing PulsePoll...
        </span>
      </div>
    );
  }

  // Routing logic
  const renderRoute = () => {
    // 1. Live Poll Results: /poll/:id/results
    const resultsMatch = currentPath.match(/^\/poll\/([a-zA-Z0-9_-]+)\/results\/?$/);
    if (resultsMatch) {
      return <PollResults pollId={resultsMatch[1]} navigate={navigate} />;
    }

    // 2. Public Voting Page: /poll/:shareCode
    const pollMatch = currentPath.match(/^\/poll\/([a-zA-Z0-9_-]+)\/?$/);
    if (pollMatch) {
      return <PublicPoll shareCode={pollMatch[1]} navigate={navigate} />;
    }

    // 3. Create Poll: /create-poll
    if (currentPath === '/create-poll') {
      if (!user) {
        return <Login navigate={navigate} />;
      }
      return <CreatePoll navigate={navigate} />;
    }

    // 4. Dashboard: /dashboard
    if (currentPath === '/dashboard') {
      if (!user) {
        return <Login navigate={navigate} />;
      }
      return <Dashboard navigate={navigate} />;
    }

    // 5. Signup: /signup
    if (currentPath === '/signup') {
      if (user) {
        return <Dashboard navigate={navigate} />;
      }
      return <Signup navigate={navigate} />;
    }

    // 6. Login: /login
    if (currentPath === '/login') {
      if (user) {
        return <Dashboard navigate={navigate} />;
      }
      return <Login navigate={navigate} />;
    }

    // Default root /
    if (user) {
      return <Dashboard navigate={navigate} />;
    }
    return <Login navigate={navigate} />;
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Navbar currentPath={currentPath} navigate={navigate} />
      <main style={{ flex: 1 }}>{renderRoute()}</main>
      <footer
        style={{
          borderTop: '1px solid var(--border-subtle)',
          padding: '24px 20px',
          textAlign: 'center',
          color: 'var(--text-muted)',
          fontSize: '0.85rem',
        }}
      >
        <span>PulsePoll &copy; {new Date().getFullYear()} &bull; GUVI Developer Internship Project &bull; Powered by React, Go, MongoDB &amp; Redis</span>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <AuthProvider>
        <MainApp />
      </AuthProvider>
    </ToastProvider>
  );
}
