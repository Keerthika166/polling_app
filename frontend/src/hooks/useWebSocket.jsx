import { useEffect, useState, useRef, useCallback } from 'react';

export function useWebSocket(pollId, initialData = null) {
  const [results, setResults] = useState(initialData?.results || []);
  const [totalVotes, setTotalVotes] = useState(initialData?.totalVotes || 0);
  const [isClosed, setIsClosed] = useState(initialData?.status === 'closed');
  const [isConnected, setIsConnected] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(null);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  // Sync with initialData changes
  useEffect(() => {
    if (initialData) {
      if (initialData.results) setResults(initialData.results);
      if (typeof initialData.totalVotes === 'number') setTotalVotes(initialData.totalVotes);
      if (initialData.status) setIsClosed(initialData.status === 'closed');
    }
  }, [initialData]);

  const connect = useCallback(() => {
    if (!pollId) return;

    const apiBase = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    const wsProto = apiBase.startsWith('https') ? 'wss' : 'ws';
    const cleanHost = apiBase.replace(/^https?:\/\//, '').replace(/\/api\/?$/, '');
    const wsUrl = `${wsProto}://${cleanHost}/api/polls/${pollId}/live`;

    console.log(`[WebSocket] Connecting to ${wsUrl}`);
    const socket = new WebSocket(wsUrl);
    wsRef.current = socket;

    socket.onopen = () => {
      console.log(`[WebSocket] Connected to poll ${pollId}`);
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log('[WebSocket] Message received:', data);

        if (data.event === 'VOTE_UPDATED') {
          if (data.results) setResults(data.results);
          if (typeof data.totalVotes === 'number') setTotalVotes(data.totalVotes);
          setLastUpdated(Date.now());
        } else if (data.event === 'POLL_CLOSED') {
          setIsClosed(true);
        } else if (data.event === 'INITIAL_STATE') {
          if (data.results) setResults(data.results);
          if (typeof data.totalVotes === 'number') setTotalVotes(data.totalVotes);
          if (data.status === 'closed') setIsClosed(true);
          setLastUpdated(Date.now());
        }
      } catch (err) {
        console.error('[WebSocket] Error parsing message:', err);
      }
    };

    socket.onclose = (event) => {
      console.warn(`[WebSocket] Disconnected code=${event.code}, reason=${event.reason}`);
      setIsConnected(false);
      // Try reconnecting in 2.5s if not cleanly unmounted
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 2500);
    };

    socket.onerror = (err) => {
      console.error('[WebSocket] Error encountered:', err);
      socket.close();
    };
  }, [pollId]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    results,
    totalVotes,
    isClosed,
    isConnected,
    lastUpdated,
  };
}
