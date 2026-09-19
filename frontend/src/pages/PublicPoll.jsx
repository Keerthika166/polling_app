import React, { useState, useEffect } from 'react';
import { api } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';
import { useToast } from '../components/Toast';
import ResultBar from '../components/ResultBar';
import confetti from 'canvas-confetti';
import { CheckCircle2, AlertCircle, Share2, BarChart2, Check, Lock, Radio } from 'lucide-react';

export default function PublicPoll({ shareCode, navigate }) {
  const toast = useToast();

  const [poll, setPoll] = useState(null);
  const [selectedOptionId, setSelectedOptionId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [voting, setVoting] = useState(false);
  const [hasVoted, setHasVoted] = useState(false);
  const [userVotedOptionId, setUserVotedOptionId] = useState(null);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  // Check local cache if user has voted on this shareCode
  useEffect(() => {
    const votedMap = JSON.parse(localStorage.getItem('user_voted_polls') || '{}');
    if (votedMap[shareCode]) {
      setHasVoted(true);
      setUserVotedOptionId(votedMap[shareCode]);
    }
  }, [shareCode]);

  useEffect(() => {
    async function loadPoll() {
      setLoading(true);
      try {
        const data = await api.polls.getByShareCode(shareCode);
        setPoll(data.poll);
      } catch (err) {
        setError(err.message || 'Poll not found or inactive');
      } finally {
        setLoading(false);
      }
    }
    loadPoll();
  }, [shareCode]);

  // Real-time WebSocket updates (active once poll is loaded)
  const pollId = poll?.id || poll?._id;
  const { results: wsResults, totalVotes: wsTotalVotes, isConnected, isClosed: wsIsClosed } = useWebSocket(
    pollId,
    poll
  );

  const currentResults = wsResults && wsResults.length > 0 ? wsResults : poll?.options || [];
  const currentTotalVotes = typeof wsTotalVotes === 'number' ? wsTotalVotes : poll?.totalVotes || 0;

  const isExpired = poll?.expiresAt && new Date(poll.expiresAt) < new Date();
  const isClosed = poll?.status === 'closed' || wsIsClosed || isExpired;

  // Identify leading option
  const maxVotes = Math.max(0, ...currentResults.map((r) => r.votes || 0));

  const handleVote = async () => {
    if (!selectedOptionId) {
      toast.info('Please choose an option to vote');
      return;
    }

    setVoting(true);
    setError('');

    try {
      await api.polls.vote(pollId, selectedOptionId);

      // Trigger celebratory confetti with fresh light ocean/emerald colors (no violet!)
      confetti({
        particleCount: 80,
        spread: 70,
        origin: { y: 0.6 },
        colors: ['#2563eb', '#0284c7', '#059669', '#10b981', '#f59e0b'],
      });

      toast.success('Your vote has been recorded!');
      setHasVoted(true);
      setUserVotedOptionId(selectedOptionId);

      // Store in localStorage
      const votedMap = JSON.parse(localStorage.getItem('user_voted_polls') || '{}');
      votedMap[shareCode] = selectedOptionId;
      localStorage.setItem('user_voted_polls', JSON.stringify(votedMap));
    } catch (err) {
      if (err.message && err.message.toLowerCase().includes('already voted')) {
        toast.info('You have already voted in this poll');
        setHasVoted(true);
      } else {
        setError(err.message || 'Failed to submit vote. Please try again.');
      }
    } finally {
      setVoting(false);
    }
  };

  const copyShareLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      toast.success('Link copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  if (loading) {
    return (
      <div className="app-container" style={{ maxWidth: '580px', textAlign: 'center', marginTop: '60px' }}>
        <div className="pulse-dot" style={{ margin: '0 auto 16px', width: '16px', height: '16px' }} />
        <p style={{ color: 'var(--text-secondary)' }}>Loading poll...</p>
      </div>
    );
  }

  if (error && !poll) {
    return (
      <div className="app-container" style={{ maxWidth: '540px', marginTop: '60px' }}>
        <div className="glass-panel" style={{ padding: '36px', textAlign: 'center' }}>
          <AlertCircle size={44} color="#e11d48" style={{ margin: '0 auto 16px' }} />
          <h2 style={{ marginBottom: '10px', color: 'var(--text-primary)' }}>Poll Unavailable</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '24px' }}>
            {error || 'This poll does not exist or the link may have expired.'}
          </p>
          <button onClick={() => navigate('/dashboard')} className="btn btn-secondary">
            Go to Home
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="app-container" style={{ maxWidth: '640px' }}>
      <div className="glass-panel" style={{ padding: '36px 30px' }}>
        {/* Header Badges */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span className="live-badge">
              <span className="pulse-dot" />
              {isConnected ? 'LIVE SYNC' : 'CONNECTING...'}
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

          <button
            onClick={copyShareLink}
            className="btn btn-sm btn-secondary"
            title="Share this poll"
          >
            {copied ? <Check size={14} color="#059669" /> : <Share2 size={14} />}
            <span>{copied ? 'Copied' : 'Share'}</span>
          </button>
        </div>

        {/* Question Title */}
        <h1 style={{ fontSize: '1.75rem', lineHeight: '1.35', marginBottom: '12px', color: 'var(--text-primary)' }}>
          {poll.question}
        </h1>

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '16px',
            fontSize: '0.85rem',
            color: 'var(--text-secondary)',
            marginBottom: '28px',
          }}
        >
          <span>
            Total votes: <strong style={{ color: 'var(--text-primary)' }}>{currentTotalVotes}</strong>
          </span>
          {poll.expiresAt && (
            <span>
              Expires:{' '}
              {new Date(poll.expiresAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          )}
        </div>

        {error && (
          <div
            style={{
              padding: '12px 16px',
              borderRadius: 'var(--radius-md)',
              background: '#fff1f2',
              border: '1px solid #fecdd3',
              color: '#e11d48',
              fontSize: '0.875rem',
              marginBottom: '20px',
            }}
          >
            {error}
          </div>
        )}

        {/* State 1: User Has NOT Voted & Poll is Active -> Show Voting Options */}
        {!hasVoted && !isClosed ? (
          <div>
            <div style={{ marginBottom: '24px' }}>
              {poll.options.map((opt, idx) => {
                const optId = opt.id || opt._id || `opt_${idx}`;
                const isSelected = selectedOptionId === optId;

                return (
                  <div
                    key={optId}
                    onClick={() => setSelectedOptionId(optId)}
                    className={`voting-option-card ${isSelected ? 'selected' : ''}`}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
                      <div className="radio-indicator">
                        <div className="radio-dot" />
                      </div>
                      <span style={{ fontWeight: 600, fontSize: '1rem', color: 'var(--text-primary)' }}>
                        {opt.text}
                      </span>
                    </div>

                    <span
                      style={{
                        fontSize: '0.8rem',
                        fontWeight: 700,
                        color: 'var(--text-secondary)',
                        padding: '2px 8px',
                        borderRadius: '4px',
                        background: '#f1f5f9',
                        border: '1px solid #e2e8f0',
                      }}
                    >
                      {String.fromCharCode(65 + idx)}
                    </span>
                  </div>
                );
              })}
            </div>

            <button
              onClick={handleVote}
              className="btn btn-primary btn-lg"
              style={{ width: '100%' }}
              disabled={voting || !selectedOptionId}
            >
              {voting ? 'Submitting Vote...' : 'Submit Vote'}
            </button>
          </div>
        ) : (
          /* State 2: User HAS Voted or Poll is Closed -> Show Live Real-Time Results */
          <div>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                padding: '12px 16px',
                borderRadius: 'var(--radius-md)',
                background: isClosed ? '#fff1f2' : '#ecfdf5',
                border: `1px solid ${isClosed ? '#fecdd3' : '#a7f3d0'}`,
                color: isClosed ? '#9f1239' : '#065f46',
                fontSize: '0.9rem',
                marginBottom: '24px',
              }}
            >
              {isClosed ? <Lock size={18} /> : <CheckCircle2 size={18} />}
              <span>
                {isClosed
                  ? 'This poll is now closed. Final results are displayed below.'
                  : 'Your vote has been submitted! Watching live results in real-time.'}
              </span>
            </div>

            <div style={{ marginBottom: '24px' }}>
              {currentResults.map((opt, idx) => {
                const optId = opt.id || opt._id || `opt_${idx}`;
                const votes = opt.votes || 0;
                const percentage =
                  currentTotalVotes > 0 ? (votes / currentTotalVotes) * 100 : 0;
                const isLeading = votes === maxVotes && maxVotes > 0;
                const isUserChoice = userVotedOptionId === optId;

                return (
                  <div key={optId} style={{ position: 'relative' }}>
                    {isUserChoice && (
                      <span
                        style={{
                          position: 'absolute',
                          right: '12px',
                          top: '-10px',
                          zIndex: 10,
                          fontSize: '0.68rem',
                          background: 'var(--accent-blue)',
                          color: '#fff',
                          padding: '2px 8px',
                          borderRadius: '10px',
                          fontWeight: 700,
                          boxShadow: '0 2px 6px rgba(37, 99, 235, 0.25)',
                        }}
                      >
                        Your Choice
                      </span>
                    )}
                    <ResultBar
                      index={idx}
                      text={opt.text}
                      votes={votes}
                      percentage={percentage}
                      isLeading={isLeading}
                      totalVotes={currentTotalVotes}
                    />
                  </div>
                );
              })}
            </div>

            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                paddingTop: '16px',
                borderTop: '1px solid var(--border-subtle)',
              }}
            >
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                Updates dynamically without page refresh
              </span>
              <button
                onClick={() => navigate(`/poll/${pollId}/results`)}
                className="btn btn-sm btn-secondary"
              >
                <BarChart2 size={15} />
                <span>Full Results View</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
