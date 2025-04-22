package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

type MediaResponse struct {
	PhotoID  uuid.UUID `json:"photoID"`
	PhotoURL string    `json:"photoURL"`

	CreatedAt time.Time `json:"createdAt"`
}

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

	defer file.Close()

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

	fileName := fmt.Sprintf("%v%v", base64.RawURLEncoding.EncodeToString([]byte(randomNumber)), extensions[0])

	imageFilePath := filepath.Join(cfg.assetsRoot, fileName)
	imageFilePath = filepath.Clean(imageFilePath)

	if !strings.HasPrefix(imageFilePath, "assets/") {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", errors.New("failed to safely parse the filepath"), false)
		return
	}

	savedFile, err := os.Create(imageFilePath)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	defer savedFile.Close()

	_, err = io.Copy(savedFile, file)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	photoURL := fmt.Sprintf("%v/%v", cfg.baseApiUrl, imageFilePath)

	if len(tripID) > 0 {
		tripUUID, err := uuid.Parse(tripID)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid trip id", err, false)
			return
		}

		trip, err := cfg.db.GetTrip(r.Context(), database.GetTripParams{
			UserID: user.ID,
			ID:     tripUUID,
		})

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Failed to get this trip, it may have been deleted", err, false)
			return
		}

		media, err := cfg.db.CreateTripMedia(r.Context(), database.CreateTripMediaParams{
			TripID: uuid.NullUUID{
				UUID:  trip.ID,
				Valid: true,
			},
			PhotoUrl: sql.NullString{
				String: photoURL,
				Valid:  true,
			},
		})

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Could not create media", err, false)
			return
		}

		respondWithJSON(w, http.StatusCreated, ApiResponse{
			Status: "success",
			Data: struct {
				PhotoID  uuid.UUID `json:"photoID"`
				PhotoURL string    `json:"photoURL"`
			}{
				PhotoID:  media.ID,
				PhotoURL: media.PhotoUrl.String,
			},
		})

		return
	}

	if len(stopID) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid Stop ID passed", err, false)
		return
	}

	stopUUID, err := uuid.Parse(stopID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid stop id", err, false)
		return
	}

	stop, err := cfg.db.GetStop(r.Context(), database.GetStopParams{
		UserID: user.ID,
		ID:     stopUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get this stop, it may have been deleted", err, false)
		return
	}

	media, err := cfg.db.CreateTripMedia(r.Context(), database.CreateTripMediaParams{
		TripStopID: uuid.NullUUID{
			UUID:  stop.ID,
			Valid: true,
		},
		PhotoUrl: sql.NullString{
			String: photoURL,
			Valid:  true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not create media", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Status: "success",
		Data: struct {
			PhotoID  uuid.UUID `json:"photoID"`
			PhotoURL string    `json:"photoURL"`
		}{
			PhotoID:  media.ID,
			PhotoURL: media.PhotoUrl.String,
		},
	})

}

func (cfg apiConfig) handlerDeletePhoto(w http.ResponseWriter, r *http.Request) {
	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	userID := r.Context().Value(UserIDKey).(uuid.UUID)

	_, err = cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	photoID := chi.URLParam(r, "mediaID")

	if len(photoID) < 1 {
		respondWithError(w, http.StatusBadRequest, "Invalid  params.", err, false)
		return
	}

	photoUUID, err := uuid.Parse(photoID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	media, err := cfg.db.GetTripMediaById(r.Context(), photoUUID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to retrieve photo to delete", err, false)
		return
	}

	err = cfg.db.DeleteTripMedia(r.Context(), photoUUID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to delete this trip's photo", err, false)
		return
	}

	imageFileName := strings.Split(media.PhotoUrl.String, "assets/")[1]
	imageFilePath := path.Join(cfg.assetsRoot, imageFileName)

	if err = utils.DeleteMedia(imageFilePath); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong.", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})
}

func (cfg apiConfig) handlerGetMedium(w http.ResponseWriter, r *http.Request) {
	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	userID := r.Context().Value(UserIDKey).(uuid.UUID)

	_, err = cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	photoID := chi.URLParam(r, "mediaID")

	if len(photoID) < 1 {
		respondWithError(w, http.StatusBadRequest, "Invalid  params.", err, false)
		return
	}

	photoUUID, err := uuid.Parse(photoID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}

	media, err := cfg.db.GetTripMediaById(r.Context(), photoUUID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to retrieve photo to delete", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data: MediaResponse{
			PhotoID:  media.ID,
			PhotoURL: media.PhotoUrl.String,
		},
	})
}

func transformMedia(media database.TripMedium) MediaResponse {
	return MediaResponse{
		PhotoID:   media.ID,
		PhotoURL:  media.PhotoUrl.String,
		CreatedAt: media.CreatedAt,
	}
}

func (cfg apiConfig) handlerGetMedia(w http.ResponseWriter, r *http.Request) {
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

	if len(tripID) > 0 {
		tripUUID, err := uuid.Parse(tripID)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid trip id", err, false)
			return
		}

		trip, err := cfg.db.GetTrip(r.Context(), database.GetTripParams{
			UserID: user.ID,
			ID:     tripUUID,
		})

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Failed to get this trip, it may have been deleted", err, false)
			return
		}

		media, err := cfg.db.GetTripMediaByTripOrStopID(r.Context(), database.GetTripMediaByTripOrStopIDParams{
			TripID: uuid.NullUUID{
				UUID:  trip.ID,
				Valid: true,
			},
		})

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Failed to get this trips medias", err, false)
			return
		}

		mediaResponse := make([]MediaResponse, 0, len(media))

		for _, medium := range media {
			mediaResponse = append(mediaResponse, transformMedia(medium))
		}

		respondWithJSON(w, http.StatusOK, ApiResponse{
			Status: "success",
			Data:   mediaResponse,
		})

		return
	}

	if len(stopID) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid Stop ID passed", err, false)
		return
	}

	stopUUID, err := uuid.Parse(stopID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid stop id", err, false)
		return
	}

	stop, err := cfg.db.GetStop(r.Context(), database.GetStopParams{
		UserID: user.ID,
		ID:     stopUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get this stop, it may have been deleted", err, false)
		return
	}

	media, err := cfg.db.GetTripMediaByTripOrStopID(r.Context(), database.GetTripMediaByTripOrStopIDParams{
		TripStopID: uuid.NullUUID{
			UUID:  stop.ID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get this trips medias", err, false)
		return
	}

	mediaResponse := make([]MediaResponse, 0, len(media))

	for _, medium := range media {
		mediaResponse = append(mediaResponse, transformMedia(medium))
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   mediaResponse,
	})
}
