package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Clip struct {
	ObjectName   string    `json:"object_name,omitempty" bson:"object_name"`
	FriendlyName string    `json:"friendly_name,omitempty" bson:"friendly_name"`
	GameName     string    `json:"game_name,omitempty" bson:"game_name"`
	Ready        bool      `json:"ready,omitempty" bson:"ready,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

func (models *Models) CreateClip(clip Clip) (*mongo.InsertOneResult, error) {
	result, err := clipsCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (models *Models) SetClipToReady(objectName string) (*mongo.UpdateResult, error) {
	filter := bson.D{
		{Key: "object_name", Value: objectName},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "ready", Value: "true"},
		},
		},
		{Key: "$set", Value: bson.D{
			{Key: "updated_at", Value: time.Now()},
		},
		},
	}

	result, err := clipsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	return result, err
}
