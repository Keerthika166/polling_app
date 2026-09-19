package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"live-polling-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PollRepository struct {
	collection *mongo.Collection
	memPolls   map[string]*models.Poll
	memMu      sync.RWMutex
}

func NewPollRepository(db *mongo.Database) *PollRepository {
	var col *mongo.Collection
	if db != nil {
		col = db.Collection("polls")
	}
	return &PollRepository{
		collection: col,
		memPolls:   make(map[string]*models.Poll),
	}
}

func (r *PollRepository) Create(ctx context.Context, poll *models.Poll) error {
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	poll.Status = "active"
	poll.TotalVotes = 0

	if r.collection != nil {
		_, err := r.collection.InsertOne(ctx, poll)
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	r.memPolls[poll.ID.Hex()] = poll
	return nil
}

func (r *PollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	if r.collection != nil {
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

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	if p, exists := r.memPolls[id.Hex()]; exists {
		// Return copy
		pCopy := *p
		return &pCopy, nil
	}
	return nil, nil
}

func (r *PollRepository) FindByShareCode(ctx context.Context, shareCode string) (*models.Poll, error) {
	if r.collection != nil {
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

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	for _, p := range r.memPolls {
		if p.ShareCode == shareCode {
			pCopy := *p
			return &pCopy, nil
		}
	}
	return nil, nil
}

func (r *PollRepository) FindByCreatorID(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	if r.collection != nil {
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

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	var polls []models.Poll
	for _, p := range r.memPolls {
		if p.CreatorID == creatorID {
			polls = append(polls, *p)
		}
	}
	if polls == nil {
		polls = []models.Poll{}
	}
	return polls, nil
}

func (r *PollRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	if r.collection != nil {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{"status": status}},
		)
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	if p, exists := r.memPolls[id.Hex()]; exists {
		p.Status = status
	}
	return nil
}

// IncrementOptionVote updates persistent records
func (r *PollRepository) IncrementOptionVote(ctx context.Context, pollID primitive.ObjectID, optionID string) error {
	if r.collection != nil {
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

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	if p, exists := r.memPolls[pollID.Hex()]; exists {
		p.TotalVotes++
		for i := range p.Options {
			if p.Options[i].ID == optionID {
				p.Options[i].Votes++
				break
			}
		}
	}
	return nil
}

func (r *PollRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if r.collection != nil {
		_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	delete(r.memPolls, id.Hex())
	return nil
}
