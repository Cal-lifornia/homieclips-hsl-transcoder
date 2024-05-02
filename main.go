package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

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
)

// const queueName = "UploadedClipsQueue"
var environment string

var (
	sqsClient *sqs.Client
	queueURL  string
)

func main() {
	testFFmpeg()

	err := godotenv.Load()
	if err != nil {
		zap.L().Fatal(err.Error(),
			zap.String("tag", "loading envs"),
			zap.String("service", "main"),
		)
	}

	zap.L().Info("successfully loaded .env",
		zap.String("tag", "loading envs"),
		zap.String("service", "main"),
	)

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

	zap.L().Info("polling sqs",
		zap.String("tag", "sqs"),
		zap.String("service", "main"),
	)
	go pollSqs(context.Background(), msgs, queueURL)

	for msg := range msgs {
		zap.L().Info("starting worker",
			zap.String("tag", "worker"),
			zap.String("service", "main"),
		)
		producer.StartWorker(msg)
		deleteMessage(context.Background(), msg, queueURL)
	}
}

func init() {
	environment = os.Getenv("ENVIRONMENT")
	var loggerConfig zap.Config
	if environment == "docker" {
		loggerConfig = zap.NewProductionConfig()
	} else {
		loggerConfig = zap.NewDevelopmentConfig()
	}
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerConfig.Build()
	zap.ReplaceGlobals(logger)
}

func testFFmpeg() {
	cmd := exec.Command("ffmpeg", "-version")
	_, err := cmd.Output()
	if err != nil {
		zap.L().Fatal("ffmpeg should be available: "+err.Error(),
			zap.String("tag", "ffmpeg-test"),
			zap.String("service", "main"),
		)
	}
}
