package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/auth"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/mailer"
	"github.com/mambo-dev/adventrak-backend/internal/utils"
)

func (cfg *apiConfig) handlerSignup(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,gte=8"`
		Username string `json:"username" validate:"required,min=5,max=20"`
	}

	err := rateLimit(w, r, "signup")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode parameters", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create password hash", err, false)
		return
	}

	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Username:     params.Username,
		PasswordHash: passwordHash,
		Email:        params.Email,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create user", err, false)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make refresh token", err, false)
		return
	}

	token, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 730),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, ApiResponse{
		Status: "success",
		Data: UserAuthResponse{
			Username:     user.Username,
			CreatedAt:    user.CreatedAt,
			AccessToken:  accessToken,
			ID:           user.ID,
			RefreshToken: token.Token,
		},
	})

}

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Password string `json:"password" validate:"required,gte=8"`
		Username string `json:"username" validate:"required,min=5,max=20"`
	}

	err := rateLimit(w, r, "login")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	decoder := json.NewDecoder(r.Body)

	params := &Params{}

	err = decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	user, err := cfg.db.GetUser(context.Background(), database.GetUserParams{
		Username: params.Username,
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.PasswordHash)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Minute*10)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make refresh token", err, false)
		return
	}

	token, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 730),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token", err, false)
		return
	}

	respondWithJSON(w, http.StatusAccepted, ApiResponse{
		Status: "success",
		Data: UserAuthResponse{
			ID:           user.ID,
			Username:     user.Username,
			AccessToken:  accessToken,
			RefreshToken: token.Token,
			CreatedAt:    user.CreatedAt,
		},
	},
	)

}

func (cfg apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		RefreshToken string `json:"refreshToken"`
	}

	decoder := json.NewDecoder(r.Body)

	params := &Params{}

	err := decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	if params.RefreshToken == "" || len(params.RefreshToken) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid Refresh Token", err, false)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(context.Background(), params.RefreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid Refresh Token", err, false)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Refresh Token was revoked at:%v", refreshToken.RevokedAt.Time), err, false)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh Token is expired login again", err, false)
		return
	}

	if !refreshToken.UserID.Valid {
		respondWithError(w, http.StatusInternalServerError, "Failed to register user to token during login", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID.UUID, cfg.jwtSecret, time.Minute*10)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	newRefreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make refresh token", err, false)
		return
	}

	token, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 730),
		UserID:    refreshToken.UserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err, false)
		return
	}

	type RefreshResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	respondWithJSON(w, http.StatusAccepted, ApiResponse{
		Status: "success",
		Data: RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: token.Token,
		},
	})

}

func (cfg apiConfig) handlerSendVerification(w http.ResponseWriter, r *http.Request) {

	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get bearer token", err, false)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Token. Please login again.", err, false)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	randomNumber, err := utils.Random32Generator()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong!", err, false)
		return
	}

	err = cfg.db.SetVerificationCode(r.Context(), database.SetVerificationCodeParams{
		VerificationCode: string(randomNumber),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		VerificationExpiresAt: sql.NullTime{
			Time:  time.Now().Add(time.Minute * 15),
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save verification Code", err, false)
		return
	}

	HTMLTemplate := mailer.MakeEmailTemplate("Verify Email address.",
		fmt.Sprintf(`Thank you %s for joining adven trak kindly click the button to verify you email.`, user.Username),
		fmt.Sprintf("%s/verify-email?code=%s&user_email=%s", cfg.frontEndURL, randomNumber, user.Email))

	err = mailer.SendEmail(mailer.EmailDetails{
		FromEmail:   mailer.SystemEmails["system"].Email,
		FromName:    mailer.SystemEmails["system"].Name,
		ToEmail:     user.Email,
		ToName:      user.Username,
		Subject:     "Verify your email address.",
		HtmlContent: HTMLTemplate,
	}, cfg.sendGridApiKey)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to send verification email", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})

}

func (cfg apiConfig) handlerVerifyEmail(w http.ResponseWriter, r *http.Request) {
	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get bearer token", err, false)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Token. Please login again.", err, false)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	verificationCode := r.URL.Query().Get("code")
	verificationEmail := r.URL.Query().Get("email")

	if verificationEmail == "" || verificationCode == "" {
		respondWithError(w, http.StatusBadRequest, "No query params sent", err, false)
		return
	}

	if verificationEmail != user.Email {
		respondWithError(w, http.StatusBadRequest, "Email must match logged in user.", err, false)
		return
	}

	account, err := cfg.db.GetUserAccount(r.Context(), uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user account", err, false)
		return
	}

	if account.VerificationCode != verificationCode {
		respondWithError(w, http.StatusForbidden, "Invalid verification code", err, false)
		return
	}

	err = cfg.db.VerifyAccount(r.Context(), uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not verify user's email", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})

}

func (cfg apiConfig) handlerResetRequest(w http.ResponseWriter, r *http.Request) {

	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	type Params struct {
		Email string `json:"email" validate:"required,email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := &Params{}

	err = decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode request body", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		Email: params.Email,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	randomNumber, err := utils.Random32Generator()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not generate reset Code", err, false)
		return
	}

	err = cfg.db.SetResetCode(r.Context(), database.SetResetCodeParams{
		ResetCode: sql.NullString{
			String: string(randomNumber),
			Valid:  true,
		},
		ResetCodeExpiresAt: sql.NullTime{
			Time:  time.Now().Add(time.Minute * 15),
			Valid: true,
		},
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save reset Code", err, false)
		return
	}

	HTMLTemplate := mailer.MakeEmailTemplate("Password Reset Request",
		fmt.Sprintln(`We have received a password request reset for you account if this was not you, 
		you can safely ignore this email. 
		If you sent one click the button below <strong>The link below expires after 15 minutes.</strong>`),
		fmt.Sprintf("%s/reset-password?reset_code=%s&user_email=%s", cfg.frontEndURL, string(randomNumber), user.Email))

	err = mailer.SendEmail(mailer.EmailDetails{
		FromEmail:   mailer.SystemEmails["system"].Email,
		FromName:    mailer.SystemEmails["system"].Name,
		ToEmail:     user.Email,
		ToName:      user.Username,
		Subject:     "Password reset request.",
		HtmlContent: HTMLTemplate,
	}, cfg.sendGridApiKey)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to send verification email", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})

}

func (cfg apiConfig) handlerResetPassword(w http.ResponseWriter, r *http.Request) {

	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	type Params struct {
		Password        string `json:"password" validate:"required,gte=8"`
		ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
	}

	params := &Params{}

	decoder := json.NewDecoder(r.Body)

	err = decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	resetCode := r.URL.Query().Get("code")
	resetEmail := r.URL.Query().Get("email")

	if resetEmail == "" || resetCode == "" {
		respondWithError(w, http.StatusBadRequest, "No query params sent", err, false)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		Email: resetEmail,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "User account may have been deleted", err, false)
		return
	}

	account, err := cfg.db.GetUserAccount(r.Context(), uuid.NullUUID{
		UUID: user.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user account", err, false)
		return
	}

	if account.ResetCode.String != resetCode {
		respondWithError(w, http.StatusForbidden, "Invalid reset code", nil, false)
		return
	}

	if time.Now().After(account.ResetCodeExpiresAt.Time) {
		respondWithError(w, http.StatusForbidden, "Reset link has expired", nil, false)
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create password hash", err, false)
		return
	}

	err = cfg.db.SetResetCode(r.Context(), database.SetResetCodeParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		ResetCode: sql.NullString{
			Valid: false,
		},
		ResetCodeExpiresAt: sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to  invalidate reset code", err, false)
		return
	}

	err = cfg.db.UpdatePassword(r.Context(), database.UpdatePasswordParams{
		PasswordHash: passwordHash,
		ID:           user.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to  update password", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})

}

func (cfg apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get bearer token", err, false)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Token. You are already logged out", err, false)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No valid token for this user found.", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, ApiResponse{
		Status: "success",
		Data:   nil,
	})

}
