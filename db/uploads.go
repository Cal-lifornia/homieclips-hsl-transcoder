package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Uploads struct {
	ObjectName string    `json:"object_name,omitempty" bson:"object_name"`
	Ready      bool      `json:"ready" bson:"ready"`
	Logs       []string  `json:"logs,omitempty" bson:"logs,omitempty"`
	Errors     []string  `json:"errors,omitempty" bson:"errors,omitempty"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

func (models *Models) CompleteUpload(objectName string) error {
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

	_, err := uploadsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	_, err = clipsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (models *Models) SendUploadLog(objectName string, data string) (*mongo.UpdateResult, error) {
	updateData := log.Prefix() + " " + data

	filter := bson.D{
		{Key: "object_name", Value: objectName},
	}
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "logs", Value: updateData},
		},
		},
		{Key: "$set", Value: bson.D{
			{Key: "updated_at", Value: time.Now()},
		},
		},
	}

	result, err := uploadsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (models *Models) SendUploadError(objectName string, err error) (*mongo.UpdateResult, error) {
	updateData := log.Prefix() + " " + err.Error()

	filter := bson.D{
		{Key: "object_name", Value: objectName},
	}
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "errors", Value: updateData},
		},
		},
		{Key: "$set", Value: bson.D{
			{Key: "updated_at", Value: time.Now()},
		},
		},
	}

	result, err := uploadsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	return result, err
}
