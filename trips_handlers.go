package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/database"
)

type TripResponse struct {
	ID                uuid.UUID       `json:"id"`
	StartDate         time.Time       `json:"startDate"`
	StartLat          interface{}     `json:"startLat"`
	StartLng          interface{}     `json:"startLng"`
	EndLat            interface{}     `json:"endLat"`
	EndLng            interface{}     `json:"endLng"`
	EndDate           sql.NullTime    `json:"endDate"`
	DistanceTravelled sql.NullFloat64 `json:"distanceTravelled"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	UserID            uuid.UUID       `json:"userId"`
}

func convertToTripResponse(dbTrip *database.GetTripRow, dbTrips *database.GetTripsRow) TripResponse {

	if dbTrip != nil {
		return TripResponse{
			ID:                dbTrip.ID,
			StartDate:         dbTrip.StartDate,
			StartLat:          dbTrip.StartLat,
			StartLng:          dbTrip.StartLng,
			EndLat:            dbTrip.EndLat,
			EndLng:            dbTrip.StartLng,
			EndDate:           dbTrip.EndDate,
			DistanceTravelled: dbTrip.DistanceTravelled,
			CreatedAt:         dbTrip.CreatedAt,
			UpdatedAt:         dbTrip.UpdatedAt,
			UserID:            dbTrip.UserID,
		}
	}

	return TripResponse{
		ID:                dbTrips.ID,
		StartDate:         dbTrips.StartDate,
		StartLat:          dbTrips.StartLat,
		StartLng:          dbTrips.StartLng,
		EndLat:            dbTrips.EndLat,
		EndLng:            dbTrips.StartLng,
		EndDate:           dbTrips.EndDate,
		DistanceTravelled: dbTrips.DistanceTravelled,
		CreatedAt:         dbTrips.CreatedAt,
		UpdatedAt:         dbTrips.UpdatedAt,
		UserID:            dbTrips.UserID,
	}
}

func (cfg apiConfig) handlerGetTrips(w http.ResponseWriter, r *http.Request) {
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

	trips, err := cfg.db.GetTrips(r.Context(), user.ID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to return users trips", err, false)
		return
	}

	jsonTrips := make([]TripResponse, 0, len(trips))

	for _, trip := range trips {

		jsonTrips = append(jsonTrips, convertToTripResponse(nil, &trip))

	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   jsonTrips,
	})
}

func (cfg apiConfig) handlerGetTrip(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusNotFound, "Failed to return users trips", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   convertToTripResponse(&trip, nil),
	})
}

type Location struct {
	Name string  `json:"name" validate:"required"`
	Lat  float64 `json:"lat" validate:"required"`
	Lng  float64 `json:"lng" validate:"required"`
}

type CreateTrip struct {
	StartDate     time.Time  `json:"startDate" validate:"required"`
	StartLocation Location   `json:"startLocation" validate:"required"`
	EndDate       *time.Time `json:"endDate,omitempty"`
}

func formatPoint(loc Location) string {
	return fmt.Sprintf("SRID=4326;POINT(%f %f)", loc.Lng, loc.Lat)
}

func (cfg apiConfig) handlerCreateTrip(w http.ResponseWriter, r *http.Request) {
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

	params := &CreateTrip{}

	if err = json.NewDecoder(r.Body).Decode(params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not read trip details", err, false)
		return
	}

	if err := validator.New().Struct(params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	var endDate sql.NullTime
	if params.EndDate != nil {
		endDate = sql.NullTime{Time: *params.EndDate, Valid: true}
	}

	tripID, err := cfg.db.CreateTrip(r.Context(), database.CreateTripParams{
		StartDate:     params.StartDate,
		EndDate:       endDate,
		StartLocation: formatPoint(params.StartLocation),
		UserID:        user.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to return users trip", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Status: "success",
		Data: struct {
			TripID uuid.UUID `json:"tripID"`
		}{
			TripID: tripID,
		},
	})
}

func (cfg apiConfig) handlerGetTrip(w http.ResponseWriter, r *http.Request) {
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

	tripID, err = cfg.db.UpdateTrip(r.Context(), database.UpdateTripParams{})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to return users trips", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data: struct {
			TripID uuid.UUID `json:"tripID"`
		}{
			TripID: tripID,
		},
	})
}
