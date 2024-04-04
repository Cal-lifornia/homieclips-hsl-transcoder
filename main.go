package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Cal-lifornia/homieclips-hsl-transcoder/db"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/worker"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConnString := fmt.Sprintf("mongodb://%s:%s@%s", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASS"), os.Getenv("DB_ADDRESS"))

	dbCtx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)

	dbClient, err := mongo.Connect(dbCtx, options.Client().ApplyURI(dbConnString))
	if err != nil {
		cancelFunc()
		log.Fatalf("ran into error connecting to mongo instance %s\n", err)
	}

	minioClient, err := setupMinio(dbCtx)
	if err != nil {
		cancelFunc()
		log.Fatalf("ran into error connecting to minio: %s\n", err)
	}

	dbCtx.Done()

	models := db.New(dbClient, os.Getenv("DB_NAME"))

	conn, err := amqp.Dial("amqps://serveradmin:secret@rabbitmq.home.local:5671")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	rabbitCh := setupRabbitMQ(conn)
	defer rabbitCh.Close()

	msgs, err := rabbitCh.Consume(
		"uploaded_files",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to register consumer")

	producer := worker.CreateWorker(minioClient, models)

	producer.StartWorker(msgs)
}
