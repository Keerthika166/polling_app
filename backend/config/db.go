package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func ConnectMongoDB(cfg *Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("[MongoDB] Connecting to %s...", cfg.MongoURI)
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("[MongoDB] Ping failed: %v", err)
		return nil, err
	}

	log.Printf("[MongoDB] Successfully connected to database: %s", cfg.DBName)
	db := client.Database(cfg.DBName)

	// Ensure indexes in background
	go ensureIndexes(db)

	return &Database{
		Client: client,
		DB:     db,
	}, nil
}

func ensureIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Users unique email index
	userCol := db.Collection("users")
	_, _ = userCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// 2. Polls share_code index
	pollsCol := db.Collection("polls")
	_, _ = pollsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "share_code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = pollsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "creator_id", Value: 1}},
	})

	// 3. Votes compound index for duplicate voting prevention
	votesCol := db.Collection("votes")
	_, _ = votesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}
