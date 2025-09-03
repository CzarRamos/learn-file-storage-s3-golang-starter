package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
)

type aspectRatio struct {
	width  int
	height int
}

func getVideoAspectRatio(filepath string) (string, error) {

	var buffer *bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	cmd.Stdout = buffer

	err := cmd.Run()
	if err != nil {
		log.Fatal("Unable to run ffprobe")
		return "", err
	}

	var detectedAspectRatio aspectRatio
	err = json.Unmarshal(buffer.Bytes(), &detectedAspectRatio)
	if err != nil {
		log.Fatal("Unable to unmarshal aspect Ratio")
		return "", err
	}

	// check aspect ratio
	if detectedAspectRatio.width == 16*detectedAspectRatio.height/9 {
		return "16:9", nil
	} else if detectedAspectRatio.height == 16*detectedAspectRatio.width/9 {
		return "9:16", nil
	}
	return "other", nil
}
