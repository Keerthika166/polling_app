package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VoteRepository struct {
	collection *mongo.Collection
	memVotes   map[string]bool
	memMu      sync.RWMutex
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	var col *mongo.Collection
	if db != nil {
		col = db.Collection("votes")
	}
	return &VoteRepository{
		collection: col,
		memVotes:   make(map[string]bool),
	}
}

func (r *VoteRepository) RecordVote(ctx context.Context, vote *models.Vote) error {
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()

	if r.collection != nil {
		_, err := r.collection.InsertOne(ctx, vote)
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	key := fmt.Sprintf("%s:%s", vote.PollID.Hex(), vote.VoterID)
	r.memVotes[key] = true
	return nil
}

func (r *VoteRepository) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterID string) (bool, error) {
	if voterID == "" {
		return false, nil
	}

	if r.collection != nil {
		count, err := r.collection.CountDocuments(ctx, bson.M{
			"poll_id":  pollID,
			"voter_id": voterID,
		})
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	key := fmt.Sprintf("%s:%s", pollID.Hex(), voterID)
	return r.memVotes[key], nil
}

func (r *VoteRepository) DeleteByPollID(ctx context.Context, pollID primitive.ObjectID) error {
	if r.collection != nil {
		_, err := r.collection.DeleteMany(ctx, bson.M{"poll_id": pollID})
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	prefix := fmt.Sprintf("%s:", pollID.Hex())
	for k := range r.memVotes {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(r.memVotes, k)
		}
	}
	return nil
}
