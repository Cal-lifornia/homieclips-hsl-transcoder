package hls_converter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type resolutionCommandResponse struct {
	Programs []interface{}     `json:"programs"`
	Streams  []videoResolution `json:"streams"`
}

type videoResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

var RES1080P = videoResolution{
	Width:  1920,
	Height: 1080,
}

var RES1440P = videoResolution{
	Width:  2560,
	Height: 1440,
}

var RES2160P = videoResolution{
	Width:  3840,
	Height: 2160,
}

func CreateHlsStreams(fileUrl string, objectName string) {
	command := fmt.Sprintf(ffmpegCommand, fileUrl, objectName, os.Getenv("APP_URL"), objectName, objectName, objectName)
	parts := strings.Split(command, "\n")

	cmd := exec.Command(parts[0], parts[1:]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())
}

func GetVideoResolution(videoPath string) (videoResolution, error) {
	command := fmt.Sprintf(getResolutionCommand, videoPath)

	parts := strings.Fields(command)

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return videoResolution{}, errors.New("ran into error running command: " + err.Error())
	}
	var results resolutionCommandResponse

	err = json.Unmarshal(output, &results)
	if err != nil {
		return videoResolution{}, errors.New("failed to unmarshal results: " + err.Error())
	}

	return videoResolution{
		Height: results.Streams[0].Height,
		Width:  results.Streams[0].Width,
	}, nil
}

func EditMasterHls(objectName string) error {
	file, err := os.ReadFile(objectName + "_master.m3u8")
	if err != nil {
		return err
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		if strings.Contains(line, objectName) {
			lines[i] = os.Getenv("APP_URL") + "/api/storage/" + objectName + "/" + line
		}
	}

	result := strings.Join(lines, "\n")
	err = os.WriteFile(objectName+".m3u8", []byte(result), 0644)
	if err != nil {
		return err
	}

	err = os.Remove(objectName + "_master.m3u8")
	if err != nil {
		return err
	}

	return nil
}

func GetHlsResults(objectName string) ([]string, error) {
	var results = []string{}
	files, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fileName := file.Name()

		if strings.HasPrefix(fileName, objectName) {
			if file.IsDir() {
				subFiles, err := os.ReadDir("./" + fileName)
				if err != nil {
					return nil, err
				}
				for _, subFile := range subFiles {
					results = append(results, subFile.Name())
				}
			} else {
				results = append(results, fileName)
			}
		}
	}
	return results, nil
}

func DeleteHlsFragments(objectName string) error {
	files, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileName := file.Name()

		if strings.HasPrefix(fileName, objectName) {
			if file.IsDir() {
				subFiles, err := os.ReadDir("./" + fileName)
				if err != nil {
					return err
				}
				for _, subFile := range subFiles {
					os.Remove(subFile.Name())
				}
			} else {
				os.Remove(objectName)
			}
		}
	}
	return nil
}
