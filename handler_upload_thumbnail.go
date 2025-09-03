package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// Implemented the upload here

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	mediaType := header.Header.Get("Content-Type")
	defer file.Close()

	// Check if valid media type
	detectedMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse media type", err)
		return
	}
	if detectedMediaType != IMAGE_JPEG && detectedMediaType != IMAGE_PNG {
		respondWithError(w, http.StatusBadRequest, "invalid file type submitted", err)
		return
	}

	// get video UUID
	videoUUID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse Video ID", err)
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

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to populate slice with random bytes", err)
		return
	}
	generatedName := base64.RawURLEncoding.EncodeToString(randomBytes)

	// saves the thumbnail in our assets folder
	extension := strings.Split(mediaType, "/")
	newFileName := fmt.Sprintf("%s.%s", generatedName, extension[len(extension)-1])
	newFilePath := filepath.Join(cfg.assetsRoot, newFileName)
	fmt.Printf("New File path made: %s ", newFilePath)

	newFile, err := os.Create(newFilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get create new file", err)
		return
	}

	io.Copy(newFile, file)
	dataURl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, newFileName)

	// update the URL and our database
	metadata.ThumbnailURL = &dataURl
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metadata)
}
