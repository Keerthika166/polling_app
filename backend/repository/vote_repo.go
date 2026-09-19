package repository

import (
	"context"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VoteRepository struct {
	collection *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	return &VoteRepository{
		collection: db.Collection("votes"),
	}
}

func (r *VoteRepository) RecordVote(ctx context.Context, vote *models.Vote) error {
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, vote)
	return err
}

func (r *VoteRepository) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterID string) (bool, error) {
	if voterID == "" {
		return false, nil
	}
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"poll_id":  pollID,
		"voter_id": voterID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *VoteRepository) DeleteByPollID(ctx context.Context, pollID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"poll_id": pollID})
	return err
}
