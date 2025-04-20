package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

func (cfg apiConfig) handlerUploadPhotos(w http.ResponseWriter, r *http.Request) {

	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	userID := r.Context().Value(UserIDKey).(uuid.UUID)

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	tripID := r.URL.Query().Get("tripID")

	stopID := r.URL.Query().Get("stopID")

	if len(tripID) > 0 && len(stopID) > 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid query params use either stop or trip id but not both", err, false)
		return
	}

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse memory", err, false)
		return
	}

	file, fileHeader, err := r.FormFile("trip_photo")

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get file", err, false)
		return
	}

	mediaTypeHeader := fileHeader.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(mediaTypeHeader)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid content type.", err, false)
		return
	}

	if mediaTypeHeader != "image/jpeg" && mediaTypeHeader != "image/png" {
		respondWithError(w, http.StatusForbidden, "Only .png and .jpeg files allowed.", err, false)
		return
	}

	extensions, err := mime.ExtensionsByType(mediaType)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	if len(extensions) < 1 {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	randomNumber, err := utils.Random32Generator()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	fileName := fmt.Sprintf("/%v%v", base64.RawURLEncoding.EncodeToString([]byte(randomNumber)), extensions[0])

	imageFilePath := filepath.Join(cfg.assetsRoot, fileName)

	savedFile, err := os.Create(imageFilePath)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	_, err = io.Copy(savedFile, file)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}
	photoURL := fmt.Sprintf("%v/%v", cfg.baseApiUrl, imageFilePath)

}
