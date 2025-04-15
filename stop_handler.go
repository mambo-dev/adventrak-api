package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

type StopResponse struct {
	ID           uuid.UUID   `json:"id"`
	LocationName string      `json:"locationName"`
	CreatedAt    time.Time   `json:"createdAt"`
	EndLat       interface{} `json:"endLat"`
	EndLng       interface{} `json:"endLng"`
}

func convertToStopRow(rows *database.GetStopsRow, row *database.GetStopRow) StopResponse {
	if row != nil {
		return StopResponse{
			ID:           row.ID,
			LocationName: row.LocationName,
			CreatedAt:    row.CreatedAt,
			EndLat:       row.EndLat,
			EndLng:       row.EndLng,
		}
	}

	return StopResponse{
		ID:           rows.ID,
		LocationName: rows.LocationName,
		CreatedAt:    rows.CreatedAt,
		EndLat:       rows.EndLat,
		EndLng:       rows.EndLng,
	}
}

func (cfg apiConfig) handlerGetStops(w http.ResponseWriter, r *http.Request) {

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

	if len(tripID) <= 0 || tripID == "" {
		respondWithError(w, http.StatusBadRequest, "Failed to get trip ID from query params", err, false)
		return
	}

	tripUUID, err := uuid.Parse(tripID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid trip ID sent through query params", err, false)
		return
	}

	stops, err := cfg.db.GetStops(r.Context(), database.GetStopsParams{
		UserID: user.ID,
		TripID: tripUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to return trip stops", err, false)
		return
	}

	stopsResponse := make([]StopResponse, 0, len(stops))
	for _, stop := range stops {
		stopsResponse = append(stopsResponse, convertToStopRow(&stop, nil))
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   stopsResponse,
	})

}

func (cfg apiConfig) handlerGetStop(w http.ResponseWriter, r *http.Request) {
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

	stopID := chi.URLParam(r, "stopID")

	stopUUID, err := uuid.Parse(stopID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	stop, err := cfg.db.GetStop(r.Context(), database.GetStopParams{
		UserID: user.ID,
		ID:     stopUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to return trip stop", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   convertToStopRow(nil, &stop),
	})
}

type StopParams struct {
	LocationTag utils.Location
	TripID      uuid.UUID
}

func (cfg apiConfig) handlerCreateStop(w http.ResponseWriter, r *http.Request) {
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

	tripID := chi.URLParam(r, "tripID")

	tripUUID, err := uuid.Parse(tripID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	trip, err := cfg.db.GetTrip(r.Context(), database.GetTripParams{
		UserID: user.ID,
		ID:     tripUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find trip possibly deleted", err, false)
		return
	}

	params := &StopParams{}
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not read trip details", err, false)
		return
	}

	stopID, err := cfg.db.CreateStop(r.Context(), database.CreateStopParams{
		LocationName: params.LocationTag.Name,
		LocationTag:  utils.FormatPoint(params.LocationTag),
		TripID:       trip.ID,
		UserID:       trip.UserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create  stop", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Status: "success",
		Data: struct {
			StopID uuid.UUID `json:"stopID"`
		}{
			StopID: stopID,
		},
	})
}

func (cfg apiConfig) handlerUpdateStop(w http.ResponseWriter, r *http.Request) {
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

	stopID := chi.URLParam(r, "stopID")

	stopUUID, err := uuid.Parse(stopID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	stop, err := cfg.db.GetStop(r.Context(), database.GetStopParams{
		UserID: user.ID,
		ID:     stopUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find stop to update possibly not deleted or you did not create this stop", err, false)
		return
	}

	params := &StopParams{}
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not read trip details", err, false)
		return
	}

	updatedStopID, err := cfg.db.UpdateStop(r.Context(), database.UpdateStopParams{
		LocationName: params.LocationTag.Name,
		LocationTag:  utils.FormatPoint(params.LocationTag),
		UserID:       user.ID,
		ID:           stop.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create  stop", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Status: "success",
		Data: struct {
			StopID uuid.UUID `json:"stopID"`
		}{
			StopID: updatedStopID,
		},
	})
}

func (cfg apiConfig) handlerDeleteStop(w http.ResponseWriter, r *http.Request) {

}
