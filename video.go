package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type FFprobeOut struct {
	Streams []struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	log.Printf("Starting to process video file: %s", filePath)

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	output := &bytes.Buffer{}
	cmd.Stdout = output
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running ffprobe command: %v", err)
		return "", err
	}

	log.Println("ffprobe command executed successfully, parsing output...")
	var result FFprobeOut
	err = json.Unmarshal(output.Bytes(), &result)
	if err != nil {
		log.Printf("Error unmarshalling ffprobe output: %v", err)
		return "", err
	}

	if len(result.Streams) == 0 {
		log.Println("No streams found in video file")
		return "", fmt.Errorf("no streams found in video file")
	}

	width := result.Streams[0].Width
	height := result.Streams[0].Height
	log.Printf("Video dimensions: width=%d, height=%d", width, height)

	if height == 0 {
		log.Println("Height is zero, invalid video dimensions")
		return "", fmt.Errorf("height is zero")
	}

	ratio := float64(width) / float64(height)
	log.Printf("Calculated aspect ratio: %.2f", ratio)

	// Determine if the ratio is close to 16:9 or 9:16 or none of it
	if ratio >= 1.7 && ratio <= 1.8 {
		log.Println("Aspect ratio is approximately 16:9")
		return "16:9", nil
	}
	if ratio >= 0.55 && ratio <= 0.6 {
		log.Println("Aspect ratio is approximately 9:16")
		return "9:16", nil
	}

	log.Println("Aspect ratio does not match 16:9 or 9:16, returning 'other'")
	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outFilePath, nil
}
