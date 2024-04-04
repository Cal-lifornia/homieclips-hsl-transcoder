package worker

import (
	"log"

	"github.com/Cal-lifornia/homieclips-hsl-transcoder/db"
	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	data        chan Job
	quit        chan chan error
	minioClient *minio.Client
	models      *db.Models
}

type Job struct {
	ObjectName string
	Success    bool
}

func (worker *Worker) Close() error {
	ch := make(chan error)
	worker.quit <- ch
	return <-ch
}

func CreateWorker(minioClient *minio.Client, models *db.Models) *Worker {
	return &Worker{
		data:        make(chan Job),
		quit:        make(chan chan error),
		minioClient: minioClient,
		models:      models,
	}

}

func (worker *Worker) StartWorker(msgs <-chan amqp091.Delivery) {
	var forever chan struct{}

	go func() {
		for msg := range msgs {
			currentJob, err := worker.transcodeUpload(msg)
			if err != nil {
				_, err = worker.models.SendUploadError(currentJob.ObjectName, err)
				if err != nil {
					log.Printf("failed to upload error to db: %s\n", err)
					msg.Reject(false)
				}
				return
			}
			if currentJob.Success {
				err := worker.models.CompleteUpload(currentJob.ObjectName)
				if err != nil {
					log.Printf("failed to send complete to db: %s\n", err)
				}
				msg.Ack(false)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
