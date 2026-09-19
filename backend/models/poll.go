package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PollOption struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type PollOptionResult struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Votes      int64   `json:"votes"`
	Percentage float64 `json:"percentage"`
}

type Poll struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question   string             `bson:"question" json:"question"`
	Options    []PollOption       `bson:"options" json:"options"`
	CreatorID  primitive.ObjectID `bson:"creator_id" json:"creatorId"`
	ShareCode  string             `bson:"share_code" json:"shareCode"`
	Status     string             `bson:"status" json:"status"` // "active", "closed"
	ExpiresAt  *time.Time         `bson:"expires_at,omitempty" json:"expiresAt,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
	TotalVotes int64              `bson:"total_votes" json:"totalVotes"`
}

type CreatePollRequest struct {
	Question      string   `json:"question" binding:"required,min=3,max=300"`
	Options       []string `json:"options" binding:"required,min=2,max=10"`
	DurationHours int      `json:"durationHours"` // 0 = never
}

type VoteRequest struct {
	OptionID          string `json:"optionId" binding:"required"`
	VoterFingerprint string `json:"voterFingerprint"`
}

type PollRealtimeUpdate struct {
	Event      string             `json:"event"` // "VOTE_UPDATED", "POLL_CLOSED"
	PollID     string             `json:"pollId"`
	TotalVotes int64              `json:"totalVotes"`
	Results    []PollOptionResult `json:"results"`
	Status     string             `json:"status,omitempty"`
}
