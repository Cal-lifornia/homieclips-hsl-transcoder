package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

func pollSqs(ctx context.Context, msgs chan<- types.Message, queueUrl string) {
	for {
		output, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueUrl),
			MaxNumberOfMessages: int32(2),
			WaitTimeSeconds:     int32(15),
		})

		if err != nil {
			zap.L().Fatal(err.Error(),
				zap.String("tag", "receiving-message"),
				zap.String("service", "sqs"),
			)
		}

		for _, message := range output.Messages {
			msgs <- message
		}
	}
}

func deleteMessage(ctx context.Context, msg types.Message, queueUrl string) {
	_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		zap.L().Error(err.Error(),
			zap.String("tag", "deleting-message"),
			zap.String("service", "sqs"),
		)
	}
}
