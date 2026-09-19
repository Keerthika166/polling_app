package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"live-polling-backend/models"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{client: client}
}

func (s *RedisService) voteKey(pollID string) string {
	return fmt.Sprintf("poll:%s:votes", pollID)
}

func (s *RedisService) channelName(pollID string) string {
	return fmt.Sprintf("poll_updates:%s", pollID)
}

// InitPollVotes initializes Redis Hash with 0 votes for each option
func (s *RedisService) InitPollVotes(ctx context.Context, pollID string, optionIDs []string) error {
	if s.client == nil {
		return nil
	}
	key := s.voteKey(pollID)
	pipe := s.client.Pipeline()
	for _, optID := range optionIDs {
		pipe.HSetNX(ctx, key, optID, 0)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[Redis] Failed to initialize poll votes hash: %v", err)
	}
	return err
}

// IncrementVote executes atomic HINCRBY in Redis
func (s *RedisService) IncrementVote(ctx context.Context, pollID string, optionID string) (int64, error) {
	if s.client == nil {
		return 0, nil
	}
	key := s.voteKey(pollID)
	val, err := s.client.HIncrBy(ctx, key, optionID, 1).Result()
	if err != nil {
		log.Printf("[Redis] HIncrBy error on %s %s: %v", key, optionID, err)
		return 0, err
	}
	return val, nil
}

// GetPollVotes retrieves all option counts from Redis Hash
func (s *RedisService) GetPollVotes(ctx context.Context, pollID string) (map[string]int64, error) {
	result := make(map[string]int64)
	if s.client == nil {
		return result, nil
	}

	key := s.voteKey(pollID)
	vals, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return result, err
	}

	for optID, countStr := range vals {
		c, _ := strconv.ParseInt(countStr, 10, 64)
		result[optID] = c
	}
	return result, nil
}

// PublishVoteUpdate broadcasts updated counts via Redis Pub/Sub
func (s *RedisService) PublishVoteUpdate(ctx context.Context, pollID string, update *models.PollRealtimeUpdate) error {
	if s.client == nil {
		return nil
	}
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}

	ch := s.channelName(pollID)
	err = s.client.Publish(ctx, ch, payload).Err()
	if err != nil {
		log.Printf("[Redis Pub/Sub] Publish failed on channel %s: %v", ch, err)
		return err
	}
	log.Printf("[Redis Pub/Sub] Broadcasted update on channel %s for poll %s", ch, pollID)
	return nil
}

// Subscribe returns a Redis Pub/Sub subscription for a given poll channel
func (s *RedisService) Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	if s.client == nil {
		return nil
	}
	return s.client.Subscribe(ctx, s.channelName(pollID))
}

// DeletePoll cleans up Redis keys associated with the poll
func (s *RedisService) DeletePoll(ctx context.Context, pollID string) error {
	if s.client == nil {
		return nil
	}
	return s.client.Del(ctx, s.voteKey(pollID)).Err()
}
