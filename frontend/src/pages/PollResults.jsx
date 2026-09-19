import React, { useState, useEffect } from 'react';
import { api } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../components/Toast';
import ResultBar from '../components/ResultBar';
import { Share2, Check, ArrowLeft, PowerOff, Users, Clock, Radio, ExternalLink } from 'lucide-react';

export default function PollResults({ pollId, navigate }) {
  const { user } = useAuth();
  const toast = useToast();

  const [poll, setPoll] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    async function loadResults() {
      setLoading(true);
      try {
        const data = await api.polls.getResults(pollId);
        setPoll(data.poll);
      } catch (err) {
        setError(err.message || 'Failed to fetch poll results');
      } finally {
        setLoading(false);
      }
    }
    loadResults();
  }, [pollId]);

  // Connect to live WebSocket stream
  const { results: wsResults, totalVotes: wsTotalVotes, isConnected, isClosed: wsIsClosed } = useWebSocket(
    pollId,
    poll
  );

  const currentResults = wsResults && wsResults.length > 0 ? wsResults : poll?.options || [];
  const currentTotalVotes = typeof wsTotalVotes === 'number' ? wsTotalVotes : poll?.totalVotes || 0;

  const isExpired = poll?.expiresAt && new Date(poll.expiresAt) < new Date();
  const isClosed = poll?.status === 'closed' || wsIsClosed || isExpired;

  const maxVotes = Math.max(0, ...currentResults.map((r) => r.votes || 0));

  const shareUrl = poll ? `${window.location.origin}/poll/${poll.shareCode}` : '';

  const copyShareLink = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      toast.success('Share link copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const handleClosePoll = async () => {
    try {
      await api.polls.close(pollId);
      toast.success('Poll closed');
      setPoll((prev) => ({ ...prev, status: 'closed' }));
    } catch (err) {
      toast.error(err.message || 'Failed to close poll');
    }
  };

  if (loading) {
    return (
      <div className="app-container" style={{ maxWidth: '640px', textAlign: 'center', marginTop: '60px' }}>
        <div className="pulse-dot" style={{ margin: '0 auto 16px', width: '16px', height: '16px' }} />
        <p style={{ color: 'var(--text-secondary)' }}>Loading live results...</p>
      </div>
    );
  }

  if (error || !poll) {
    return (
      <div className="app-container" style={{ maxWidth: '540px', marginTop: '60px' }}>
        <div className="glass-panel" style={{ padding: '36px', textAlign: 'center' }}>
          <h2 style={{ marginBottom: '10px', color: 'var(--text-primary)' }}>Results Unavailable</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '24px' }}>
            {error || 'Unable to find this poll.'}
          </p>
          <button onClick={() => navigate('/dashboard')} className="btn btn-secondary">
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  const isCreator = user && (user.id === poll.creatorId || user._id === poll.creatorId);

  return (
    <div className="app-container" style={{ maxWidth: '800px' }}>
      {/* Top Navigation Row */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '24px',
        }}
      >
        <button onClick={() => navigate('/dashboard')} className="btn btn-sm btn-secondary">
          <ArrowLeft size={16} />
          <span>Dashboard</span>
        </button>

        <div style={{ display: 'flex', gap: '8px' }}>
          <button onClick={() => navigate(`/poll/${poll.shareCode}`)} className="btn btn-sm btn-secondary">
            <ExternalLink size={15} />
            <span>Open Vote Page</span>
          </button>

          <button onClick={copyShareLink} className="btn btn-sm btn-primary">
            {copied ? <Check size={15} /> : <Share2 size={15} />}
            <span>{copied ? 'Copied' : 'Share Link'}</span>
          </button>
        </div>
      </div>

      {/* Main Results Card */}
      <div className="glass-panel" style={{ padding: '40px 36px' }}>
        {/* Live Indicator Banner */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '24px',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <span className="live-badge">
              <span className="pulse-dot" />
              {isConnected ? 'LIVE WEBSOCKET FEED' : 'CONNECTING...'}
            </span>
            {isClosed && (
              <span
                style={{
                  padding: '4px 10px',
                  borderRadius: '9999px',
                  fontSize: '0.75rem',
                  fontWeight: 700,
                  background: '#fff1f2',
                  color: '#e11d48',
                  border: '1px solid #fecdd3',
                  textTransform: 'uppercase',
                }}
              >
                Poll Closed
              </span>
            )}
          </div>

          {isCreator && !isClosed && (
            <button
              onClick={handleClosePoll}
              className="btn btn-sm btn-secondary"
              style={{ color: '#e11d48' }}
            >
              <PowerOff size={14} />
              <span>Close Poll</span>
            </button>
          )}
        </div>

        {/* Question Title */}
        <h1 style={{ fontSize: '2rem', lineHeight: '1.3', marginBottom: '20px', color: 'var(--text-primary)' }}>
          {poll.question}
        </h1>

        {/* Metrics Pill Bar */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '20px',
            padding: '16px 20px',
            background: '#f8fafc',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--border-subtle)',
            marginBottom: '32px',
            fontSize: '0.9rem',
            color: 'var(--text-secondary)',
            flexWrap: 'wrap',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Users size={18} color="var(--accent-blue)" />
            <span>
              Total Votes:{' '}
              <strong style={{ color: 'var(--text-primary)', fontSize: '1.1rem' }}>
                {currentTotalVotes}
              </strong>
            </span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Clock size={16} />
            <span>Created {new Date(poll.createdAt).toLocaleDateString()}</span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginLeft: 'auto' }}>
            <span>Audience Code:</span>
            <code
              style={{
                background: '#eff6ff',
                padding: '3px 8px',
                borderRadius: '4px',
                fontFamily: 'JetBrains Mono',
                fontWeight: 600,
                color: 'var(--accent-blue)',
                border: '1px solid #bfdbfe',
              }}
            >
              {poll.shareCode}
            </code>
          </div>
        </div>

        {/* Live Result Bars */}
        <div style={{ marginBottom: '32px' }}>
          {currentResults.map((opt, idx) => {
            const optId = opt.id || opt._id || `opt_${idx}`;
            const votes = opt.votes || 0;
            const percentage =
              currentTotalVotes > 0 ? (votes / currentTotalVotes) * 100 : 0;
            const isLeading = votes === maxVotes && maxVotes > 0;

            return (
              <ResultBar
                key={optId}
                index={idx}
                text={opt.text}
                votes={votes}
                percentage={percentage}
                isLeading={isLeading}
                totalVotes={currentTotalVotes}
              />
            );
          })}
        </div>

        {/* Footer info */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            paddingTop: '20px',
            borderTop: '1px solid var(--border-subtle)',
            fontSize: '0.85rem',
            color: 'var(--text-muted)',
            flexWrap: 'wrap',
            gap: '10px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <Radio size={14} color="#059669" />
            <span>Streaming live vote changes via Redis Pub/Sub &amp; WebSockets</span>
          </div>

          <span>Share URL: {shareUrl}</span>
        </div>
      </div>
    </div>
  );
}
