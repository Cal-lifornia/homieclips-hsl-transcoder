package worker

import (
	"errors"
	"fmt"
	"github.com/Cal-lifornia/homieclips-hsl-transcoder/hls_converter"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"sync"
)

func (worker *Worker) transcodeUpload(objectName string) (*Job, error) {
	zap.L().Info("beginning work on "+objectName,
		zap.String("tag", "job"),
		zap.String("service", "worker"),
	)
	success := false

	// Get the presigned objectUrl for transcoding
	objectUrl, err := worker.GetObject("uploaded/" + objectName)
	if err != nil {
		return nil, err
	}

	// Pass objectUrl to transcoder
	hls_converter.CreateHlsStreams(objectUrl.URL, objectName)

	// Send success to db
	_, err = worker.models.SendUploadLog(objectName, "successfully created hls streams")
	if err != nil {
		zap.L().Error(err.Error(),
			zap.String("tag", "sending log to db"),
			zap.String("service", "worker"),
		)
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
			err = worker.UploadHslFragment(input, name)
			if err != nil {
				color.Red("Error: ", err)
				trouble = errors.New(err.Error())
				return
			}
			zap.L().Info("uploaded fragment "+name,
				zap.String("tag", "uploading-fragment"),
				zap.String("service", "s3"))
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
	err = worker.MoveOriginalUpload(objectName)
	if err != nil {
		return nil, err
	}

	// Send success log to db
	_, err = worker.models.SendUploadLog(objectName, "uploaded original file "+objectName)
	if err != nil {
		fmt.Println("failed to send log to database")
	}

	// Delete leftover fragments
	err = hls_converter.DeleteHlsFragments(objectName)
	if err != nil {
		return nil, err
	}

	success = true

	return &Job{
		Success: success,
	}, nil
}
