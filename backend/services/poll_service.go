package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"live-polling-backend/models"
	"live-polling-backend/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PollService struct {
	pollRepo     *repository.PollRepository
	voteRepo     *repository.VoteRepository
	redisService *RedisService
}

func NewPollService(
	pollRepo *repository.PollRepository,
	voteRepo *repository.VoteRepository,
	redisService *RedisService,
) *PollService {
	return &PollService{
		pollRepo:     pollRepo,
		voteRepo:     voteRepo,
		redisService: redisService,
	}
}

// Generate unique 6-character alphanumeric share code
func generateShareCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // omit ambiguous chars
	b := make([]byte, 6)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func (s *PollService) CreatePoll(ctx context.Context, creatorID string, req *models.CreatePollRequest) (*models.Poll, error) {
	cID, err := primitive.ObjectIDFromHex(creatorID)
	if err != nil {
		return nil, errors.New("invalid creator ID")
	}

	// Validate question
	q := strings.TrimSpace(req.Question)
	if len(q) < 3 {
		return nil, errors.New("question must be at least 3 characters")
	}

	// Validate options
	if len(req.Options) < 2 || len(req.Options) > 10 {
		return nil, errors.New("a poll must have between 2 and 10 options")
	}

	seen := make(map[string]bool)
	var pollOptions []models.PollOption
	var optionIDs []string

	for idx, optText := range req.Options {
		trimmed := strings.TrimSpace(optText)
		if trimmed == "" {
			return nil, errors.New("option text cannot be empty")
		}
		lower := strings.ToLower(trimmed)
		if seen[lower] {
			return nil, fmt.Errorf("duplicate option: %s", trimmed)
		}
		seen[lower] = true

		optID := fmt.Sprintf("opt_%d", idx+1)
		pollOptions = append(pollOptions, models.PollOption{
			ID:    optID,
			Text:  trimmed,
			Votes: 0,
		})
		optionIDs = append(optionIDs, optID)
	}

	// Calculate expiration if duration specified
	var expiresAt *time.Time
	if req.DurationHours > 0 {
		exp := time.Now().Add(time.Duration(req.DurationHours) * time.Hour)
		expiresAt = &exp
	}

	poll := &models.Poll{
		Question:   q,
		Options:    pollOptions,
		CreatorID:  cID,
		ShareCode:  generateShareCode(),
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		Status:     "active",
		TotalVotes: 0,
	}

	if err := s.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	// Initialize Redis Hash for fast O(1) in-memory vote counters
	_ = s.redisService.InitPollVotes(ctx, poll.ID.Hex(), optionIDs)

	return poll, nil
}

func (s *PollService) GetUserPolls(ctx context.Context, creatorID string) ([]models.Poll, error) {
	cID, err := primitive.ObjectIDFromHex(creatorID)
	if err != nil {
		return nil, errors.New("invalid creator ID")
	}

	polls, err := s.pollRepo.FindByCreatorID(ctx, cID)
	if err != nil {
		return nil, err
	}

	// Enrich with real-time vote totals from Redis
	for i := range polls {
		pID := polls[i].ID.Hex()
		redisVotes, err := s.redisService.GetPollVotes(ctx, pID)
		if err == nil && len(redisVotes) > 0 {
			var total int64
			for j := range polls[i].Options {
				if v, ok := redisVotes[polls[i].Options[j].ID]; ok {
					polls[i].Options[j].Votes = v
				}
				total += polls[i].Options[j].Votes
			}
			polls[i].TotalVotes = total
		}
	}

	return polls, nil
}

func (s *PollService) GetPollByID(ctx context.Context, idStr string) (*models.Poll, error) {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errors.New("invalid poll ID")
	}

	poll, err := s.pollRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, errors.New("poll not found")
	}

	s.enrichWithRedisCounts(ctx, poll)
	return poll, nil
}

func (s *PollService) GetPollByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	poll, err := s.pollRepo.FindByShareCode(ctx, strings.ToUpper(strings.TrimSpace(shareCode)))
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, errors.New("poll not found")
	}

	s.enrichWithRedisCounts(ctx, poll)
	return poll, nil
}

func (s *PollService) enrichWithRedisCounts(ctx context.Context, poll *models.Poll) {
	redisVotes, err := s.redisService.GetPollVotes(ctx, poll.ID.Hex())
	if err == nil && len(redisVotes) > 0 {
		var total int64
		for i := range poll.Options {
			if count, ok := redisVotes[poll.Options[i].ID]; ok {
				poll.Options[i].Votes = count
			}
			total += poll.Options[i].Votes
		}
		poll.TotalVotes = total
	}
}

func (s *PollService) Vote(
	ctx context.Context,
	pollIDStr string,
	req *models.VoteRequest,
	ipAddress string,
	userAgent string,
) (*models.PollRealtimeUpdate, error) {
	pollID, err := primitive.ObjectIDFromHex(pollIDStr)
	if err != nil {
		return nil, errors.New("invalid poll ID")
	}

	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return nil, errors.New("poll not found")
	}

	// 1. Validation: Poll status
	if poll.Status == "closed" {
		return nil, errors.New("this poll is closed and no longer accepting votes")
	}

	// 2. Validation: Expiration
	if poll.ExpiresAt != nil && time.Now().After(*poll.ExpiresAt) {
		return nil, errors.New("this poll has expired")
	}

	// 3. Validation: Option belongs to poll
	var validOption bool
	for _, opt := range poll.Options {
		if opt.ID == req.OptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, errors.New("invalid option selected")
	}

	// 4. Duplicate vote verification
	voterIdentifier := req.VoterFingerprint
	if voterIdentifier == "" {
		voterIdentifier = ipAddress
	}

	hasVoted, err := s.voteRepo.HasVoted(ctx, pollID, voterIdentifier)
	if err != nil {
		return nil, err
	}
	if hasVoted {
		return nil, errors.New("you have already voted in this poll")
	}

	// 5. Store audit record in MongoDB votes collection
	vote := &models.Vote{
		PollID:    pollID,
		OptionID:  req.OptionID,
		VoterID:   voterIdentifier,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	if err := s.voteRepo.RecordVote(ctx, vote); err != nil {
		return nil, errors.New("failed to record vote")
	}

	// 6. Update MongoDB persistent poll counts
	_ = s.pollRepo.IncrementOptionVote(ctx, pollID, req.OptionID)

	// 7. Atomic Redis increment (Fast O(1) in-memory)
	_, _ = s.redisService.IncrementVote(ctx, pollIDStr, req.OptionID)

	// 8. Fetch updated counts from Redis
	redisVotes, _ := s.redisService.GetPollVotes(ctx, pollIDStr)

	var totalVotes int64
	for _, opt := range poll.Options {
		c := redisVotes[opt.ID]
		totalVotes += c
	}

	var results []models.PollOptionResult
	for _, opt := range poll.Options {
		c := redisVotes[opt.ID]
		var pct float64
		if totalVotes > 0 {
			pct = (float64(c) / float64(totalVotes)) * 100.0
		}
		results = append(results, models.PollOptionResult{
			ID:         opt.ID,
			Text:       opt.Text,
			Votes:      c,
			Percentage: pct,
		})
	}

	updatePayload := &models.PollRealtimeUpdate{
		Event:      "VOTE_UPDATED",
		PollID:     pollIDStr,
		TotalVotes: totalVotes,
		Results:    results,
	}

	// 9. Publish update event via Redis Pub/Sub to trigger Go WebSocket broadcast!
	_ = s.redisService.PublishVoteUpdate(ctx, pollIDStr, updatePayload)

	return updatePayload, nil
}

func (s *PollService) ClosePoll(ctx context.Context, pollIDStr string, userID string) error {
	pollID, err := primitive.ObjectIDFromHex(pollIDStr)
	if err != nil {
		return errors.New("invalid poll ID")
	}

	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return errors.New("poll not found")
	}

	if poll.CreatorID.Hex() != userID {
		return errors.New("unauthorized: only the poll creator can close this poll")
	}

	if err := s.pollRepo.UpdateStatus(ctx, pollID, "closed"); err != nil {
		return err
	}

	// Broadcast POLL_CLOSED event via Redis Pub/Sub
	_ = s.redisService.PublishVoteUpdate(ctx, pollIDStr, &models.PollRealtimeUpdate{
		Event:  "POLL_CLOSED",
		PollID: pollIDStr,
		Status: "closed",
	})

	return nil
}

func (s *PollService) DeletePoll(ctx context.Context, pollIDStr string, userID string) error {
	pollID, err := primitive.ObjectIDFromHex(pollIDStr)
	if err != nil {
		return errors.New("invalid poll ID")
	}

	poll, err := s.pollRepo.FindByID(ctx, pollID)
	if err != nil || poll == nil {
		return errors.New("poll not found")
	}

	if poll.CreatorID.Hex() != userID {
		return errors.New("unauthorized: only the poll creator can delete this poll")
	}

	_ = s.pollRepo.Delete(ctx, pollID)
	_ = s.voteRepo.DeleteByPollID(ctx, pollID)
	_ = s.redisService.DeletePoll(ctx, pollIDStr)

	return nil
}

func (s *PollService) BuildInitialSnapshot(ctx context.Context, poll *models.Poll) *models.PollRealtimeUpdate {
	s.enrichWithRedisCounts(ctx, poll)

	var results []models.PollOptionResult
	for _, opt := range poll.Options {
		var pct float64
		if poll.TotalVotes > 0 {
			pct = (float64(opt.Votes) / float64(poll.TotalVotes)) * 100.0
		}
		results = append(results, models.PollOptionResult{
			ID:         opt.ID,
			Text:       opt.Text,
			Votes:      opt.Votes,
			Percentage: pct,
		})
	}

	return &models.PollRealtimeUpdate{
		Event:      "INITIAL_STATE",
		PollID:     poll.ID.Hex(),
		TotalVotes: poll.TotalVotes,
		Results:    results,
		Status:     poll.Status,
	}
}
