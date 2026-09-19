const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Generate or retrieve unique voter anonymous client ID for duplicate vote prevention
export function getVoterFingerprint() {
  let fid = localStorage.getItem('voter_fingerprint');
  if (!fid) {
    fid = 'voter_' + Math.random().toString(36).substring(2, 11) + Date.now().toString(36);
    localStorage.setItem('voter_fingerprint', fid);
  }
  return fid;
}

async function request(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const token = localStorage.getItem('token');

  const headers = {
    'Content-Type': 'application/json',
    'X-Voter-Fingerprint': getVoterFingerprint(),
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const error = new Error(data.error || data.message || 'An error occurred');
    error.status = response.status;
    error.data = data;
    throw error;
  }

  return data;
}

export const api = {
  auth: {
    signup: (name, email, password) =>
      request('/auth/signup', {
        method: 'POST',
        body: JSON.stringify({ name, email, password }),
      }),
    login: (email, password) =>
      request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    getMe: () => request('/auth/me'),
  },

  polls: {
    create: (data) =>
      request('/polls', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getUserPolls: () => request('/polls'),
    getById: (id) => request(`/polls/${id}`),
    getByShareCode: (shareCode) => request(`/polls/share/${shareCode}`),
    close: (id) =>
      request(`/polls/${id}/close`, {
        method: 'PATCH',
      }),
    delete: (id) =>
      request(`/polls/${id}`, {
        method: 'DELETE',
      }),
    vote: (id, optionId) =>
      request(`/polls/${id}/vote`, {
        method: 'POST',
        body: JSON.stringify({
          optionId,
          voterFingerprint: getVoterFingerprint(),
        }),
      }),
    getResults: (id) => request(`/polls/${id}/results`),
  },
};
