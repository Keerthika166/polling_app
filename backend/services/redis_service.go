package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"live-polling-backend/models"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client    *redis.Client
	memCounts map[string]map[string]int64
	memSubs   map[string][]chan []byte
	memMu     sync.RWMutex
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client:    client,
		memCounts: make(map[string]map[string]int64),
		memSubs:   make(map[string][]chan []byte),
	}
}

func (s *RedisService) HasLiveClient() bool {
	return s.client != nil
}

// Ping tests live connection to Redis server
func (s *RedisService) Ping(ctx context.Context) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("redis client is not initialized")
	}
	return s.client.Ping(ctx).Result()
}

func (s *RedisService) voteKey(pollID string) string {
	return fmt.Sprintf("poll:%s:votes", pollID)
}

func (s *RedisService) channelName(pollID string) string {
	return fmt.Sprintf("poll_updates:%s", pollID)
}

// InitPollVotes initializes Redis Hash with 0 votes for each option
func (s *RedisService) InitPollVotes(ctx context.Context, pollID string, optionIDs []string) error {
	if s.client != nil {
		key := s.voteKey(pollID)
		pipe := s.client.Pipeline()
		for _, optID := range optionIDs {
			pipe.HSetNX(ctx, key, optID, 0)
		}
		_, err := pipe.Exec(ctx)
		if err == nil {
			return nil
		}
		log.Printf("[Redis] HSetNX failed, using in-memory store: %v", err)
	}

	// In-memory fallback
	s.memMu.Lock()
	defer s.memMu.Unlock()
	if _, exists := s.memCounts[pollID]; !exists {
		s.memCounts[pollID] = make(map[string]int64)
	}
	for _, optID := range optionIDs {
		if _, ok := s.memCounts[pollID][optID]; !ok {
			s.memCounts[pollID][optID] = 0
		}
	}
	return nil
}

// IncrementVote executes atomic HINCRBY in Redis
func (s *RedisService) IncrementVote(ctx context.Context, pollID string, optionID string) (int64, error) {
	if s.client != nil {
		key := s.voteKey(pollID)
		val, err := s.client.HIncrBy(ctx, key, optionID, 1).Result()
		if err == nil {
			return val, nil
		}
		log.Printf("[Redis] HIncrBy error on %s %s: %v. Using in-memory store.", key, optionID, err)
	}

	// In-memory fallback
	s.memMu.Lock()
	defer s.memMu.Unlock()
	if _, exists := s.memCounts[pollID]; !exists {
		s.memCounts[pollID] = make(map[string]int64)
	}
	s.memCounts[pollID][optionID]++
	return s.memCounts[pollID][optionID], nil
}

// GetPollVotes retrieves all option counts from Redis Hash
func (s *RedisService) GetPollVotes(ctx context.Context, pollID string) (map[string]int64, error) {
	if s.client != nil {
		key := s.voteKey(pollID)
		vals, err := s.client.HGetAll(ctx, key).Result()
		if err == nil && len(vals) > 0 {
			result := make(map[string]int64)
			for optID, countStr := range vals {
				c, _ := strconv.ParseInt(countStr, 10, 64)
				result[optID] = c
			}
			return result, nil
		}
	}

	// In-memory fallback
	s.memMu.RLock()
	defer s.memMu.RUnlock()
	result := make(map[string]int64)
	if counts, exists := s.memCounts[pollID]; exists {
		for k, v := range counts {
			result[k] = v
		}
	}
	return result, nil
}

// PublishVoteUpdate broadcasts updated counts via Redis Pub/Sub
func (s *RedisService) PublishVoteUpdate(ctx context.Context, pollID string, update *models.PollRealtimeUpdate) error {
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}

	if s.client != nil {
		ch := s.channelName(pollID)
		err := s.client.Publish(ctx, ch, payload).Err()
		if err == nil {
			log.Printf("[Redis Pub/Sub] Broadcasted update on channel %s for poll %s", ch, pollID)
			return nil
		}
		log.Printf("[Redis Pub/Sub] Publish failed on channel %s: %v. Using in-memory broadcaster.", ch, err)
	}

	// In-memory broadcaster fallback
	s.memMu.RLock()
	defer s.memMu.RUnlock()
	if channels, ok := s.memSubs[pollID]; ok {
		for _, ch := range channels {
			select {
			case ch <- payload:
			default:
			}
		}
	}
	return nil
}

// Subscribe returns a Redis Pub/Sub subscription for a given poll channel
func (s *RedisService) Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	if s.client == nil {
		return nil
	}
	return s.client.Subscribe(ctx, s.channelName(pollID))
}

// SubscribeFallback registers an in-memory channel subscription when Redis is offline
func (s *RedisService) SubscribeFallback(pollID string, ch chan []byte) func() {
	s.memMu.Lock()
	s.memSubs[pollID] = append(s.memSubs[pollID], ch)
	s.memMu.Unlock()

	return func() {
		s.memMu.Lock()
		defer s.memMu.Unlock()
		subs := s.memSubs[pollID]
		for i, c := range subs {
			if c == ch {
				s.memSubs[pollID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// DeletePoll cleans up Redis keys associated with the poll
func (s *RedisService) DeletePoll(ctx context.Context, pollID string) error {
	if s.client != nil {
		_ = s.client.Del(ctx, s.voteKey(pollID)).Err()
	}

	s.memMu.Lock()
	delete(s.memCounts, pollID)
	delete(s.memSubs, pollID)
	s.memMu.Unlock()
	return nil
}
