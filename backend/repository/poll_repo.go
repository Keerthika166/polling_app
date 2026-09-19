package repository

import (
	"context"
	"errors"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PollRepository struct {
	collection *mongo.Collection
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	return &PollRepository{
		collection: db.Collection("polls"),
	}
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	poll.Status = "active"
	poll.TotalVotes = 0

	_, err := r.collection.InsertOne(ctx, poll)
	return err
}

func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"share_code": shareCode}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepository) FindByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"creator_id": creatorID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

// IncrementOptionVote updates MongoDB persistent records
func (r *PollRepository) IncrementOptionVote(ctx context.Context, pollID primitive.ObjectID, optionID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": pollID, "options.id": optionID},
		bson.M{
			"$inc": bson.M{
				"options.$.votes": 1,
				"total_votes":     1,
			},
		},
	)
	return err
}

func (r *PollRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
