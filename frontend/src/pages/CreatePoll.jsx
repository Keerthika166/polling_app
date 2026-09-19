import React, { useState } from 'react';
import { api } from '../services/api';
import { useToast } from '../components/Toast';
import { Plus, Trash2, ArrowRight, HelpCircle, Clock, Sparkles } from 'lucide-react';

export default function CreatePoll({ navigate }) {
  const toast = useToast();

  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState(['', '']);
  const [durationHours, setDurationHours] = useState(0); // 0 = never
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleOptionChange = (index, value) => {
    const updated = [...options];
    updated[index] = value;
    setOptions(updated);
  };

  const addOption = () => {
    if (options.length >= 10) {
      toast.info('Maximum 10 options allowed');
      return;
    }
    setOptions([...options, '']);
  };

  const removeOption = (index) => {
    if (options.length <= 2) {
      toast.info('A poll must have at least 2 options');
      return;
    }
    setOptions(options.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const trimmedQuestion = question.trim();
    if (!trimmedQuestion) {
      setError('Please enter a poll question.');
      return;
    }

    const trimmedOptions = options.map((opt) => opt.trim()).filter(Boolean);
    if (trimmedOptions.length < 2) {
      setError('Please provide at least 2 non-empty options.');
      return;
    }

    // Duplicate check
    const uniqueOptions = new Set(trimmedOptions.map((o) => o.toLowerCase()));
    if (uniqueOptions.size !== trimmedOptions.length) {
      setError('Options must be unique. Please remove duplicates.');
      return;
    }

    setLoading(true);
    try {
      const data = await api.polls.create({
        question: trimmedQuestion,
        options: trimmedOptions,
        durationHours: Number(durationHours),
      });

      toast.success('Poll created successfully!');
      if (data.poll && (data.poll.id || data.poll._id)) {
        navigate(`/poll/${data.poll.id || data.poll._id}/results`);
      } else {
        navigate('/dashboard');
      }
    } catch (err) {
      setError(err.message || 'Failed to create poll');
    } finally {
      setLoading(false);
    }
  };

  const sampleTemplates = [
    {
      q: 'What is your favorite programming language?',
      opts: ['Python', 'JavaScript / TypeScript', 'Go', 'Java', 'Rust'],
    },
    {
      q: 'How should our team schedule the next sprint planning?',
      opts: ['Monday Morning', 'Tuesday Afternoon', 'Async via Slack', 'Friday Recap'],
    },
    {
      q: 'Which cloud provider do you prefer for microservices?',
      opts: ['AWS', 'Google Cloud', 'Microsoft Azure', 'DigitalOcean / VPS'],
    },
  ];

  const applyTemplate = (template) => {
    setQuestion(template.q);
    setOptions(template.opts);
  };

  return (
    <div className="app-container" style={{ maxWidth: '780px' }}>
      <div className="glass-panel" style={{ padding: '36px' }}>
        <div style={{ marginBottom: '28px' }}>
          <h1 style={{ fontSize: '1.8rem', marginBottom: '8px', color: 'var(--text-primary)' }}>
            Create a New Live Poll
          </h1>
          <p style={{ color: 'var(--text-secondary)' }}>
            Formulate your question, add choices, and share the live link with your audience.
          </p>
        </div>

        {/* Quick Template Presets */}
        <div style={{ marginBottom: '24px' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
            <Sparkles size={14} color="#0284c7" />
            <span>QUICK TEMPLATES</span>
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
            {sampleTemplates.map((t, idx) => (
              <button
                key={idx}
                type="button"
                onClick={() => applyTemplate(t)}
                className="btn btn-sm btn-secondary"
                style={{ fontSize: '0.8rem', borderStyle: 'dashed' }}
              >
                {t.q.substring(0, 36)}...
              </button>
            ))}
          </div>
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
              marginBottom: '24px',
            }}
          >
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Question Input */}
          <div className="form-group">
            <label className="form-label" htmlFor="poll-question">
              Poll Question <span style={{ color: 'var(--accent-rose)' }}>*</span>
            </label>
            <input
              id="poll-question"
              type="text"
              className="form-input"
              placeholder="e.g., What is your favorite programming language?"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              required
              maxLength={200}
            />
          </div>

          {/* Options List */}
          <div className="form-group">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
              <label className="form-label" style={{ margin: 0 }}>
                Voting Options ({options.length}/10) <span style={{ color: 'var(--accent-rose)' }}>*</span>
              </label>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Min 2 options</span>
            </div>

            {options.map((opt, idx) => (
              <div key={idx} style={{ display: 'flex', gap: '10px', marginBottom: '10px' }}>
                <span
                  style={{
                    width: '36px',
                    height: '42px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: '#f1f5f9',
                    border: '1px solid #e2e8f0',
                    borderRadius: 'var(--radius-md)',
                    fontWeight: 700,
                    fontSize: '0.85rem',
                    color: 'var(--text-secondary)',
                  }}
                >
                  {String.fromCharCode(65 + idx)}
                </span>
                <input
                  type="text"
                  className="form-input"
                  placeholder={`Option ${idx + 1}`}
                  value={opt}
                  onChange={(e) => handleOptionChange(idx, e.target.value)}
                  required
                />
                {options.length > 2 && (
                  <button
                    type="button"
                    onClick={() => removeOption(idx)}
                    className="btn btn-secondary"
                    title="Remove option"
                    style={{ padding: '0 12px', color: 'var(--accent-rose)' }}
                  >
                    <Trash2 size={16} />
                  </button>
                )}
              </div>
            ))}

            {options.length < 10 && (
              <button
                type="button"
                onClick={addOption}
                className="btn btn-secondary btn-sm"
                style={{ alignSelf: 'flex-start', marginTop: '6px', borderStyle: 'dashed' }}
              >
                <Plus size={15} />
                <span>Add Another Option</span>
              </button>
            )}
          </div>

          {/* Poll Expiration */}
          <div className="form-group" style={{ marginTop: '24px' }}>
            <label className="form-label" style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Clock size={16} />
              <span>Poll Duration</span>
            </label>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '10px' }}>
              {[
                { label: 'Never Expires', val: 0 },
                { label: '1 Hour', val: 1 },
                { label: '24 Hours', val: 24 },
                { label: '7 Days', val: 168 },
              ].map((item) => (
                <button
                  key={item.val}
                  type="button"
                  onClick={() => setDurationHours(item.val)}
                  className={`btn btn-sm ${durationHours === item.val ? 'btn-primary' : 'btn-secondary'}`}
                  style={{ padding: '10px 6px', fontSize: '0.85rem' }}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {/* Action Row */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginTop: '36px',
              paddingTop: '20px',
              borderTop: '1px solid var(--border-subtle)',
            }}
          >
            <button
              type="button"
              onClick={() => navigate('/dashboard')}
              className="btn btn-secondary"
            >
              Cancel
            </button>

            <button
              type="submit"
              className="btn btn-primary btn-lg"
              disabled={loading}
            >
              {loading ? 'Publishing Poll...' : 'Create & Launch Poll'}
              <ArrowRight size={18} />
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
