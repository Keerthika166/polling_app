package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    primitive.ObjectID `bson:"poll_id" json:"pollId"`
	OptionID  string             `bson:"option_id" json:"optionId"`
	VoterID   string             `bson:"voter_id" json:"voterId"`
	IPAddress string             `bson:"ip_address" json:"ipAddress"`
	UserAgent string             `bson:"user_agent" json:"userAgent"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}
