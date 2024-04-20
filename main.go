package main

import (
	"context"
	"fmt"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/db"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/storage"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/worker"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

// const queueName = "UploadedClipsQueue"

var (
	sqsClient *sqs.Client
	queueURL  string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	queueURL = os.Getenv("SQS_QUEUE_URL")

	dbConnString := fmt.Sprintf("mongodb://%s:%s@%s", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASS"), os.Getenv("DB_ADDRESS"))

	dbCtx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)

	dbClient, err := mongo.Connect(dbCtx, options.Client().ApplyURI(dbConnString))
	if err != nil {
		cancelFunc()
		zap.L().Fatal(err.Error(),
			zap.String("tag", "connecting-to-db"),
			zap.String("service", "main"),
		)
	}

	dbCtx.Done()

	models := db.New(dbClient, os.Getenv("DB_NAME"))

	awsCtx, canFunc := context.WithTimeout(context.Background(), 10*time.Second)

	awsConfig, err := config.LoadDefaultConfig(awsCtx, config.WithRegion("ap-southeast-2"))
	if err != nil {
		canFunc()
		zap.L().Fatal(err.Error(),
			zap.String("tag", "connecting-to-s3"),
			zap.String("service", "main"),
		)
	}

	sqsClient = sqs.NewFromConfig(awsConfig)
	msgs := make(chan types.Message, 2)

	storageClient := storage.New(awsConfig)

	producer := worker.CreateWorker(storageClient, models)

	go pollSqs(context.Background(), msgs, queueURL)

	for msg := range msgs {
		producer.StartWorker(msg)
		deleteMessage(context.Background(), msg)
	}
}

func init() {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerConfig.Build()
	zap.ReplaceGlobals(logger)

}
