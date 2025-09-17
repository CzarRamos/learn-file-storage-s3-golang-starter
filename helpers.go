package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type AspectRatio struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func getVideoAspectRatio(filepath string) (string, error) {

	var buffer bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	cmd.Stdout = &buffer

	err := cmd.Run()
	if err != nil {
		log.Fatal("Unable to run ffprobe")
		return "", err
	}

	var detectedAspectRatio AspectRatio
	err = json.Unmarshal(buffer.Bytes(), &detectedAspectRatio)
	if err != nil {
		log.Fatal("Unable to unmarshal aspect Ratio")
		return "", err
	}

	// check aspect ratio
	if detectedAspectRatio.Streams[0].Width == 16*detectedAspectRatio.Streams[0].Height/9 {
		return "landscape", nil
	} else if detectedAspectRatio.Streams[0].Height == 16*detectedAspectRatio.Streams[0].Width/9 {
		return "portrait", nil
	}
	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {

	outputPath := fmt.Sprintf("%s.%s", filePath, ".processing")

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
// 	presignClient := s3.NewPresignClient(s3Client)
// 	req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
// 		Bucket: &bucket,
// 		Key:    &key,
// 	}, s3.WithPresignExpires(expireTime))
// 	if err != nil {
// 		return "", err
// 	}

// 	return req.URL, nil
// }

// func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
// 	parts := strings.SplitN(*video.VideoURL, ",", 2)
// 	if len(parts) != 2 {
// 		return database.Video{}, fmt.Errorf("invalid video_url format")
// 	}

// 	// bucket, key := parts[0], parts[1]
// 	// presignURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute)
// 	// if err != nil {
// 	// 	log.Print("error: unable to convert db video to signed video")
// 	// 	return database.Video{}, err
// 	// }

// 	// video.VideoURL = &presignURL

// 	return video, nil
// }
