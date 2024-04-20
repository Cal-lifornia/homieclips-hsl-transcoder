package storage

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type Storage struct {
	*s3.Client
}

func New(awsConfig aws.Config) *Storage {
	s3Client := s3.NewFromConfig(awsConfig)
	return &Storage{s3Client}
}

func (storage *Storage) GetObject(objectName string) (*v4.PresignedHTTPRequest, error) {
	presignClient := s3.NewPresignClient(storage.Client)
	request, err := presignClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
			Key:    aws.String(objectName),
		},
		s3.WithPresignExpires(time.Minute*15),
	)
	if err != nil {
		return nil, err
	}

	return request, err
}

func (storage *Storage) UploadHslFragment(fileName string, folderName string) error {
	objectName := fileName
	filePath := "./" + fileName
	var contentType string
	if strings.HasSuffix(fileName, ".ts") {
		contentType = "video/mp2t"
	} else if strings.HasSuffix(fileName, ".m3u8") {
		contentType = "application/x-mpegURL"
	} else {
		return errors.New("file is not .ts or .m3u8")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = storage.Client.PutObject(
		context.Background(),
		&s3.PutObjectInput{
			Bucket:      aws.String(os.Getenv("BUCKET_NAME")),
			Key:         aws.String("stream/" + folderName + "/" + objectName),
			Body:        file,
			ContentType: aws.String(contentType),
		},
	)
	if err != nil {
		return errors.New(fileName + " failed to upload to s3: " + err.Error())
	}

	return nil
}

func (storage *Storage) MoveOriginalUpload(objectName string) error {
	_, err := storage.Client.CopyObject(context.Background(),
		&s3.CopyObjectInput{
			Bucket:     aws.String(os.Getenv("BUCKET_NAME")),
			CopySource: aws.String(os.Getenv("BUCKET_NAME") + "/uploaded/" + objectName),
			Key:        aws.String("archive/" + objectName),
		})

	if err != nil {
		return errors.New("failed copying original upload: " + err.Error())
	}

	err = storage.deleteOriginalUpload(objectName)
	if err != nil {
		return err
	}
	zap.L().Info("successfully moved "+objectName,
		zap.String("tag", "moving-object"),
		zap.String("service", "s3"))

	return nil
}

func (storage *Storage) deleteOriginalUpload(objectName string) error {
	_, err := storage.Client.DeleteObject(context.Background(),
		&s3.DeleteObjectInput{
			Key:    aws.String("uploaded/" + objectName),
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		})
	if err != nil {
		return errors.New("failed to delete original upload: " + err.Error())
	}

	return nil
}
