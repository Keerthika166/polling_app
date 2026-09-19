import React from 'react';
import { Trophy } from 'lucide-react';

export default function ResultBar({ index, text, votes, percentage, isLeading, totalVotes }) {
  const letters = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J'];
  const optionLetter = letters[index] || (index + 1);

  // Safe percentage
  const pct = Math.max(0, Math.min(100, Math.round((percentage || 0) * 10) / 10));

  return (
    <div className="result-row">
      <div className="result-bar-bg">
        <div
          className={`result-bar-fill ${isLeading && votes > 0 ? 'winner' : ''}`}
          style={{ width: `${pct}%` }}
        />
        <div className="result-content">
          <div className="result-text">
            <span
              style={{
                width: '24px',
                height: '24px',
                borderRadius: '6px',
                background: isLeading && votes > 0 ? 'rgba(16, 185, 129, 0.2)' : 'rgba(255, 255, 255, 0.08)',
                color: isLeading && votes > 0 ? '#34d399' : 'var(--text-secondary)',
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '0.75rem',
                fontWeight: 700,
              }}
            >
              {optionLetter}
            </span>
            <span>{text}</span>
            {isLeading && votes > 0 && totalVotes > 0 && (
              <span
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '4px',
                  fontSize: '0.75rem',
                  color: '#34d399',
                  background: 'rgba(16, 185, 129, 0.15)',
                  padding: '2px 8px',
                  borderRadius: '12px',
                  fontWeight: 600,
                }}
              >
                <Trophy size={12} /> Leading
              </span>
            )}
          </div>

          <div className="result-meta">
            <span className="result-percentage">{pct}%</span>
            <span className="result-votes">
              {votes} {votes === 1 ? 'vote' : 'votes'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
