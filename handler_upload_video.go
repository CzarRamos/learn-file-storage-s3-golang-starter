package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	maxBytes := http.MaxBytesReader(w, r.Body, 1<<30)
	r.Body = maxBytes

	videoIDString := r.PathValue("videoID")
	videoUUID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// grab video metadata
	metadata, err := cfg.db.GetVideo(videoUUID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video file", err)
		return
	}

	// checks if user is the actual owner of the video
	if metadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	videoFile, videoHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video file", err)
		return
	}
	mediaType := videoHeader.Header.Get("Content-Type")
	defer videoFile.Close()

	// check if valid format (aka video/mp4)
	detectedMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse media type", err)
		return
	}
	if detectedMediaType != VIDEO_MP4 {
		respondWithError(w, http.StatusBadRequest, "invalid file type submitted", fmt.Errorf("error: invalid media type"))
		return
	}

	// create temp file
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create temp file", err)
		return
	}
	// tempFile not needed later on
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, videoFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy video file", err)
		return
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to seek in file", err)
		return
	}

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to populate slice with random bytes", err)
		return
	}
	generatedName := hex.EncodeToString(randomBytes)
	// saves the thumbnail in our assets folder
	newFileName := fmt.Sprintf("%s.%s", generatedName, detectedMediaType)
	// newFilePath := filepath.Join(cfg.assetsRoot, newFileName)
	// fmt.Printf("New File path made: %s ", newFilePath)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &newFileName,
		Body:        tempFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to upload to AWS bucket", err)
		return
	}

	// update the URL and our database
	dataURl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, newFileName)
	metadata.VideoURL = &dataURl
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video metadata", err)
		return
	}

}
