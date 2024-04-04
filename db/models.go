package db

import "go.mongodb.org/mongo-driver/mongo"

type Models struct {
	db *mongo.Client
}

var clipsCollection *mongo.Collection
var uploadsCollection *mongo.Collection

func New(mongoClient *mongo.Client, dbName string) *Models {
	newModels := &Models{
		db: mongoClient,
	}

	newModels.init(dbName)
	return newModels
}

func (models *Models) init(dbName string) {
	clipsCollection = models.db.Database(dbName).Collection("clips")
	uploadsCollection = models.db.Database(dbName).Collection("uploads")
}
