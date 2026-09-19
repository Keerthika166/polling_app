import React, { useState } from 'react';
import { Share2, BarChart2, ExternalLink, PowerOff, Trash2, Check, Clock, Users } from 'lucide-react';
import { useToast } from './Toast';

export default function PollCard({ poll, onStatusChange, onDelete, navigate }) {
  const toast = useToast();
  const [copied, setCopied] = useState(false);

  const shareUrl = `${window.location.origin}/poll/${poll.shareCode}`;

  const copyShareLink = async (e) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      toast.success('Share link copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const isExpired = poll.expiresAt && new Date(poll.expiresAt) < new Date();
  const isClosed = poll.status === 'closed' || isExpired;

  const totalVotes = poll.totalVotes || 0;

  return (
    <div className="glass-panel" style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '12px' }}>
        <h3 style={{ fontSize: '1.2rem', lineHeight: '1.4', flex: 1, color: 'var(--text-primary)' }}>
          {poll.question}
        </h3>
        <span
          className={!isClosed ? 'live-badge' : ''}
          style={
            isClosed
              ? {
                  padding: '4px 10px',
                  borderRadius: '9999px',
                  fontSize: '0.75rem',
                  fontWeight: 700,
                  background: '#fff1f2',
                  color: '#e11d48',
                  border: '1px solid #fecdd3',
                  textTransform: 'uppercase',
                }
              : {}
          }
        >
          {!isClosed && <span className="pulse-dot" />}
          {isExpired ? 'Expired' : poll.status === 'closed' ? 'Closed' : 'Active'}
        </span>
      </div>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '16px',
          fontSize: '0.85rem',
          color: 'var(--text-secondary)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <Users size={16} color="var(--accent-blue)" />
          <strong style={{ color: 'var(--text-primary)' }}>{totalVotes}</strong> {totalVotes === 1 ? 'vote' : 'votes'}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <Clock size={16} />
          <span>{new Date(poll.createdAt).toLocaleDateString()}</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <span style={{ color: 'var(--text-muted)' }}>Code:</span>
          <code
            style={{
              background: '#f1f5f9',
              padding: '2px 8px',
              borderRadius: '4px',
              fontFamily: 'JetBrains Mono',
              color: 'var(--accent-blue)',
              fontWeight: 600,
              border: '1px solid #e2e8f0',
            }}
          >
            {poll.shareCode}
          </code>
        </div>
      </div>

      <div
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          gap: '8px',
          marginTop: 'auto',
          paddingTop: '16px',
          borderTop: '1px solid var(--border-subtle)',
        }}
      >
        <button
          onClick={() => navigate(`/poll/${poll.id || poll._id}/results`)}
          className="btn btn-sm btn-primary"
        >
          <BarChart2 size={15} />
          <span>Live Results</span>
        </button>

        <button
          onClick={() => navigate(`/poll/${poll.shareCode}`)}
          className="btn btn-sm btn-secondary"
        >
          <ExternalLink size={15} />
          <span>Vote Page</span>
        </button>

        <button
          onClick={copyShareLink}
          className="btn btn-sm btn-secondary"
          title="Copy direct share link"
        >
          {copied ? <Check size={15} color="#059669" /> : <Share2 size={15} />}
          <span>{copied ? 'Copied' : 'Share Link'}</span>
        </button>

        {!isClosed && (
          <button
            onClick={() => onStatusChange(poll.id || poll._id, 'closed')}
            className="btn btn-sm btn-secondary"
            title="End voting"
            style={{ marginLeft: 'auto' }}
          >
            <PowerOff size={15} />
            <span>Close Poll</span>
          </button>
        )}

        <button
          onClick={() => onDelete(poll.id || poll._id)}
          className="btn btn-sm btn-danger"
          title="Delete poll"
          style={isClosed ? { marginLeft: 'auto' } : {}}
        >
          <Trash2 size={15} />
        </button>
      </div>
    </div>
  );
}
