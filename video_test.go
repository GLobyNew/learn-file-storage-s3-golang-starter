package main

import (
	"os/exec"
	"testing"
)

func TestGetVideoAspectRatio(t *testing.T) {
	// Test with a valid video file
	videoFile := "testSamples/test169.mp4" // Replace with a valid video file path
	expectedAspectRatio := "16:9"          // Adjust based on the actual video file

	// Mock the ffprobe command
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", videoFile)
	_, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error running ffprobe command: %v", err)
	}

	// Call the function to test
	actualAspectRatio, err := getVideoAspectRatio(videoFile)
	if err != nil {
		t.Fatalf("Error getting aspect ratio: %v", err)
	}

	if actualAspectRatio != expectedAspectRatio {
		t.Errorf("Expected aspect ratio %s, got %s", expectedAspectRatio, actualAspectRatio)
	}
}
