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
)

type UserRepository struct {
	collection *mongo.Collection
	memUsers   map[string]*models.User
	memMu      sync.RWMutex
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	var col *mongo.Collection
	if db != nil {
		col = db.Collection("users")
	}
	return &UserRepository{
		collection: col,
		memUsers:   make(map[string]*models.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()

	if r.collection != nil {
		_, err := r.collection.InsertOne(ctx, user)
		return err
	}

	// Memory fallback
	r.memMu.Lock()
	defer r.memMu.Unlock()
	r.memUsers[user.Email] = user
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.collection != nil {
		var user models.User
		err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, nil
			}
			return nil, err
		}
		return &user, nil
	}

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	if u, exists := r.memUsers[email]; exists {
		return u, nil
	}
	return nil, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	if r.collection != nil {
		var user models.User
		err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, nil
			}
			return nil, err
		}
		return &user, nil
	}

	// Memory fallback
	r.memMu.RLock()
	defer r.memMu.RUnlock()
	for _, u := range r.memUsers {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}
