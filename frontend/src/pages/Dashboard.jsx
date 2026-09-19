import React, { useState, useEffect } from 'react';
import { api } from '../services/api';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../components/Toast';
import PollCard from '../components/PollCard';
import { PlusCircle, BarChart3, Search, RefreshCw, AlertCircle, HelpCircle } from 'lucide-react';

export default function Dashboard({ navigate }) {
  const { user } = useAuth();
  const toast = useToast();

  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all'); // all, active, closed
  const [search, setSearch] = useState('');

  const fetchPolls = async () => {
    setLoading(true);
    try {
      const data = await api.polls.getUserPolls();
      setPolls(data.polls || []);
    } catch (err) {
      toast.error(err.message || 'Failed to fetch polls');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPolls();
  }, []);

  const handleStatusChange = async (pollId, newStatus) => {
    try {
      if (newStatus === 'closed') {
        await api.polls.close(pollId);
        toast.success('Poll closed successfully');
        setPolls((prev) =>
          prev.map((p) => (p.id === pollId || p._id === pollId ? { ...p, status: 'closed' } : p))
        );
      }
    } catch (err) {
      toast.error(err.message || 'Failed to update poll');
    }
  };

  const handleDelete = async (pollId) => {
    if (!window.confirm('Are you sure you want to delete this poll? This action cannot be undone.')) {
      return;
    }
    try {
      await api.polls.delete(pollId);
      toast.success('Poll deleted');
      setPolls((prev) => prev.filter((p) => p.id !== pollId && p._id !== pollId));
    } catch (err) {
      toast.error(err.message || 'Failed to delete poll');
    }
  };

  // Filter & search
  const filteredPolls = polls.filter((poll) => {
    const isExpired = poll.expiresAt && new Date(poll.expiresAt) < new Date();
    const isClosed = poll.status === 'closed' || isExpired;

    if (filter === 'active' && isClosed) return false;
    if (filter === 'closed' && !isClosed) return false;

    if (search.trim()) {
      return poll.question.toLowerCase().includes(search.toLowerCase());
    }
    return true;
  });

  const totalVotesAcrossAll = polls.reduce((sum, p) => sum + (p.totalVotes || 0), 0);
  const activePollsCount = polls.filter(
    (p) => p.status === 'active' && (!p.expiresAt || new Date(p.expiresAt) > new Date())
  ).length;

  return (
    <div className="app-container">
      {/* Header Bar */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '16px',
          marginBottom: '32px',
        }}
      >
        <div>
          <h1 style={{ fontSize: '2rem', marginBottom: '6px', color: 'var(--text-primary)' }}>
            Welcome back, {user?.name?.split(' ')[0] || 'Creator'} 👋
          </h1>
          <p style={{ color: 'var(--text-secondary)' }}>
            Monitor your live polls, view real-time audience feedback, and distribute share links.
          </p>
        </div>

        <button
          onClick={() => navigate('/create-poll')}
          className="btn btn-primary btn-lg"
        >
          <PlusCircle size={18} />
          <span>Create New Poll</span>
        </button>
      </div>

      {/* Metrics Row */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
          gap: '16px',
          marginBottom: '32px',
        }}
      >
        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '8px' }}>
            Total Polls Created
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--accent-blue)' }}>
            {polls.length}
          </div>
        </div>

        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '8px' }}>
            Active Live Polls
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--accent-emerald)' }}>
            {activePollsCount}
          </div>
        </div>

        <div className="glass-panel" style={{ padding: '20px' }}>
          <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '8px' }}>
            Total Audience Votes
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--accent-cyan)' }}>
            {totalVotesAcrossAll}
          </div>
        </div>
      </div>

      {/* Filters & Search Toolbar */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '16px',
          marginBottom: '24px',
        }}
      >
        <div style={{ display: 'flex', gap: '8px' }}>
          {['all', 'active', 'closed'].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`btn btn-sm ${filter === f ? 'btn-primary' : 'btn-secondary'}`}
              style={{ textTransform: 'capitalize' }}
            >
              {f} Polls
            </button>
          ))}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div style={{ position: 'relative', minWidth: '240px' }}>
            <Search
              size={16}
              color="var(--text-muted)"
              style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)' }}
            />
            <input
              type="text"
              className="form-input"
              style={{ padding: '8px 12px 8px 36px', fontSize: '0.9rem' }}
              placeholder="Search polls..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          <button
            onClick={fetchPolls}
            className="btn btn-sm btn-secondary"
            title="Refresh poll list"
          >
            <RefreshCw size={15} />
          </button>
        </div>
      </div>

      {/* Polls List */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '60px 0', color: 'var(--text-secondary)' }}>
          <RefreshCw size={32} className="pulse-dot" style={{ margin: '0 auto 16px' }} />
          <p>Loading your polls...</p>
        </div>
      ) : filteredPolls.length === 0 ? (
        <div
          className="glass-panel"
          style={{
            textAlign: 'center',
            padding: '60px 24px',
            borderStyle: 'dashed',
          }}
        >
          <BarChart3 size={48} color="var(--text-muted)" style={{ margin: '0 auto 16px', opacity: 0.5 }} />
          <h3 style={{ marginBottom: '8px', fontSize: '1.25rem' }}>No polls found</h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '24px', maxWidth: '400px', margin: '0 auto 24px' }}>
            {search
              ? 'No polls match your search query.'
              : filter !== 'all'
              ? `You do not have any ${filter} polls right now.`
              : 'You haven’t created any live polls yet. Create your first poll to share with your audience!'}
          </p>
          <button onClick={() => navigate('/create-poll')} className="btn btn-primary">
            <PlusCircle size={16} />
            <span>Create Poll Now</span>
          </button>
        </div>
      ) : (
        <div className="grid-cols-2">
          {filteredPolls.map((poll) => (
            <PollCard
              key={poll.id || poll._id}
              poll={poll}
              onStatusChange={handleStatusChange}
              onDelete={handleDelete}
              navigate={navigate}
            />
          ))}
        </div>
      )}
    </div>
  );
}
