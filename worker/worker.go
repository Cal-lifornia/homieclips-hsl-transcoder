package worker

import (
	"encoding/json"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/db"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/storage"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"log"
	"strings"
)

type Worker struct {
	*storage.Storage
	models *db.Models
}

type Job struct {
	Success bool
}

func CreateWorker(storageClient *storage.Storage, models *db.Models) *Worker {
	return &Worker{
		Storage: storageClient,
		models:  models,
	}

}

type sqsBaseNotification struct {
	Record []sqsNotificationBody `json:"Records"`
}

type sqsNotificationBody struct {
	EventName string `json:"eventName"`
	S3        struct {
		Object struct {
			Key string `json:"key"`
		} `json:"object"`
	} `json:"s3"`
}

func (worker *Worker) StartWorker(msg types.Message) {
	if strings.Contains(*msg.Body, "s3:TestEvent") {
		zap.L().Info("event is not the one required",
			zap.String("tag", "wrong event type"),
			zap.String("service", "worker"),
		)
		return
	}
	var msgBody sqsBaseNotification

	err := json.Unmarshal([]byte(*msg.Body), &msgBody)
	if err != nil {
		zap.L().Error(err.Error(),
			zap.String("tag", "unmarshaling message body"),
			zap.String("service", "worker"),
		)
		return
	}

	//Get the object name
	objectName := strings.TrimPrefix(msgBody.Record[0].S3.Object.Key, "uploaded/")

	currentJob, err := worker.transcodeUpload(objectName)
	if err != nil {
		_, err = worker.models.SendUploadError(objectName, err)
		if err != nil {
			log.Printf("failed to upload error to db: %s\n", err)
			if err != nil {
				color.Red("error: %s\n", err)
			}
		}
		return
	}
	if currentJob.Success {
		err := worker.models.CompleteUpload(objectName)
		if err != nil {
			log.Printf("failed to send complete to db: %s\n", err)
		}
	}

	zap.L().Info("completed work on "+objectName,
		zap.String("tag", "job"),
		zap.String("service", "worker"),
	)

}
