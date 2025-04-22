package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

type TripResponse struct {
	ID                uuid.UUID       `json:"id"`
	StartDate         time.Time       `json:"startDate"`
	StartLocationName string          `json:"startLocationName"`
	EndLocationName   sql.NullString  `json:"endLocationName"`
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
			StartLocationName: dbTrip.StartLocationName,
			EndLocationName:   dbTrip.EndLocationName,
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
		StartLocationName: dbTrips.StartLocationName,
		EndLocationName:   dbTrips.EndLocationName,
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

type TripDetails struct {
	StartDate     time.Time      `json:"startDate" validate:"required"`
	StartLocation utils.Location `json:"startLocation" validate:"required"`
	EndDate       *time.Time     `json:"endDate,omitempty"`
	TripTitle     string         `json:"tripTitle" validate:"required"`
}

type EndTrip struct {
	EndDate     *time.Time     `json:"endDate,omitempty"`
	EndLocation utils.Location `json:"endLocation" validate:"required"`
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

	params := &TripDetails{}

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
		StartDate:         params.StartDate,
		EndDate:           endDate,
		StartLocation:     utils.FormatPoint(params.StartLocation),
		UserID:            user.ID,
		StartLocationName: params.StartLocation.Name,
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

func (cfg apiConfig) handlerUpdateTripDetails(w http.ResponseWriter, r *http.Request) {
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

	paramID := chi.URLParam(r, "tripID")

	tripUUID, err := uuid.Parse(paramID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	params := &TripDetails{}

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

	tripID, err := cfg.db.UpdateTrip(r.Context(), database.UpdateTripParams{
		EndDate:           endDate,
		StartLocation:     utils.FormatPoint(params.StartLocation),
		UserID:            user.ID,
		ID:                tripUUID,
		StartLocationName: params.StartLocation.Name,
		UpdatedAt:         time.Now(),
	})

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

func (cfg apiConfig) handlerMarkTripComplete(w http.ResponseWriter, r *http.Request) {
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

	paramID := chi.URLParam(r, "tripID")

	tripUUID, err := uuid.Parse(paramID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	params := &EndTrip{}

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

	tripID, err := cfg.db.MarkTripEnd(r.Context(), database.MarkTripEndParams{
		EndDate:     endDate,
		EndLocation: utils.FormatPoint(params.EndLocation),
		EndLocationName: sql.NullString{
			String: params.EndLocation.Name,
			Valid:  true,
		},
		UserID:    user.ID,
		ID:        tripUUID,
		UpdatedAt: time.Now(),
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to compelete trip", err, false)
		return
	}

	distanceTravelled, err := cfg.db.GetTripDistance(r.Context(), tripUUID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get distance travelled", err, true)
		return
	}

	log.Println(math.Round(distanceTravelled))

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data: struct {
			TripID            uuid.UUID `json:"tripID"`
			DistanceTravelled float64   `json:"distanceTravelled"`
		}{
			TripID:            tripID,
			DistanceTravelled: math.Round(distanceTravelled),
		},
	})
}

func (cfg apiConfig) handlerDeleteTrip(w http.ResponseWriter, r *http.Request) {
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

	paramID := chi.URLParam(r, "tripID")

	tripUUID, err := uuid.Parse(paramID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid route path", err, false)
		return
	}

	media, err := cfg.db.GetTripMediaByTripOrStopID(r.Context(), database.GetTripMediaByTripOrStopIDParams{
		TripID: uuid.NullUUID{
			UUID:  tripUUID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get trip assets", err, false)
		return
	}

	for _, medium := range media {
		imageFileName := strings.Split(medium.PhotoUrl.String, "assets/")[1]
		imageFilePath := path.Join(cfg.assetsRoot, imageFileName)

		if err = utils.DeleteMedia(imageFilePath); err != nil {
			log.Printf("Failed to delete media at %v", imageFilePath)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong.", err, false)
			return
		}

	}

	err = cfg.db.DeleteTrip(r.Context(), database.DeleteTripParams{
		UserID: user.ID,
		ID:     tripUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to delete user's trip", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})
}
