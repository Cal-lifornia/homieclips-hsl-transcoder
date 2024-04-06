package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Cal-lifornia/homieclips-hsl-transcoder/hls_converter"
	"github.com/fatih/color"
	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
)

type minioNotificationBody struct {
	EventName string      `json:"event_name"`
	Key       string      `json:"key"`
	Recorder  interface{} `json:"-"`
}

func (worker *Worker) transcodeUpload(msg amqp091.Delivery) (*Job, error) {
	color.Green("grabbed message")
	success := false

	var msgBody minioNotificationBody

	err := json.Unmarshal(msg.Body, &msgBody)
	if err != nil {
		return nil, err
	}
	// Get the object name
	objectName := strings.TrimPrefix(msgBody.Key, "homieclips/uploaded/")

	// Get the presigned minioUrl for transcoding
	minioUrl, err := worker.getUrl("uploaded/" + objectName)
	if err != nil {
		return nil, err
	}

	// Pass minioUrl to transcoder
	hls_converter.CreateHlsStreams(minioUrl.String(), objectName)

	// Send success to db
	_, err = worker.models.SendUploadLog(objectName, "successfully created hls streams")
	if err != nil {
		color.Red("failed to send log to database")
	}

	// Edit the master playlist
	err = hls_converter.EditMasterHls(objectName)
	if err != nil {
		return nil, err
	}

	// Get the hls files
	results, err := hls_converter.GetHlsResults(objectName)
	if err != nil {
		return nil, err
	}

	// Declare waitgroup for uploading the files
	var wg sync.WaitGroup

	// Upload the files
	for _, result := range results {
		var trouble error = nil

		wg.Add(1)
		go func(input string, name string, trouble error) {
			defer wg.Done()
			uploadInfo, err := uploadHslFragment(worker.minioClient, input, name)
			if err != nil {
				color.Red("Error: ", err)
				trouble = errors.New(err.Error())
				return
			}
			fmt.Println("Success: ", uploadInfo.Key)
		}(result, objectName, trouble)
		if trouble != nil {
			return nil, err
		}
	}
	wg.Wait()

	// Send success log to db
	_, err = worker.models.SendUploadLog(objectName, "successfully uploaded "+objectName+"hls chunks")
	if err != nil {
		color.Red("failed to send log to database")
	}

	// Copy the original object
	uploadInfo, err := copyOriginalUpload(worker.minioClient, objectName)
	if err != nil {
		return nil, err
	}

	// Send success log to db
	_, err = worker.models.SendUploadLog(objectName, "uploaded original file "+uploadInfo.Key)
	if err != nil {
		fmt.Println("failed to send log to database")
	}

	// Delete leftover fragments
	err = hls_converter.DeleteHlsFragments(objectName)
	if err != nil {
		return nil, err
	}

	// Delete original object
	err = deleteOriginalUpload(worker.minioClient, objectName)
	if err != nil {
		return nil, err
	}

	success = true

	return &Job{
		ObjectName: objectName,
		Success:    success,
	}, nil
}

func (worker *Worker) getUrl(objectName string) (*url.URL, error) {
	reqParams := make(url.Values)
	reqParams.Set("Content-Type", "video/mp4")
	reqParams.Set("Connection", "keep-alive")

	preSignedURL, err := worker.minioClient.PresignedGetObject(
		context.Background(),
		os.Getenv("BUCKET_NAME"),
		objectName,
		time.Hour,
		reqParams,
	)
	if err != nil {
		return nil, err
	}
	return preSignedURL, nil
}

func deleteOriginalUpload(client *minio.Client, objectName string) error {
	err := client.RemoveObject(context.Background(), os.Getenv("BUCKET_NAME"), "uploaded/"+objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func uploadHslFragment(client *minio.Client, fileName string, folderName string) (minio.UploadInfo, error) {
	objectName := fileName
	filePath := "./" + fileName
	var contentType string
	if strings.HasSuffix(fileName, ".ts") {
		contentType = "video/mp2t"
	} else if strings.HasSuffix(fileName, ".m3u8") {
		contentType = "application/x-mpegURL"
	} else {
		return minio.UploadInfo{}, errors.New("file is not .ts or .m3u8")
	}

	uploadInfo, err := client.FPutObject(
		context.Background(),
		os.Getenv("BUCKET_NAME"),
		"stream/"+folderName+"/"+objectName,
		filePath,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return minio.UploadInfo{}, errors.New(fileName + " failed to upload to minio: " + err.Error())
	}

	return uploadInfo, nil
}

func copyOriginalUpload(client *minio.Client, objectName string) (minio.UploadInfo, error) {
	srcOpts := minio.CopySrcOptions{
		Bucket: os.Getenv("BUCKET_NAME"),
		Object: "uploaded/" + objectName,
	}

	destOps := minio.CopyDestOptions{
		Bucket: os.Getenv("BUCKET_NAME"),
		Object: "originals/" + objectName,
	}

	uploadInfo, err := client.CopyObject(context.Background(), destOps, srcOpts)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return uploadInfo, nil
}
