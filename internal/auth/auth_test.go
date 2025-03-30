package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"Valid password": {
			input: "AdventrakPassword",
			want:  true,
		},
		"Another password": {
			input: "SecureP@ss123",
			want:  false,
		},
		"Wrong password test": {
			input: "WrongPassword",
			want:  false,
		},
	}

	correctPassword := "AdventrakPassword"
	hashed, err := HashPassword(correctPassword)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	if len(hashed) == 0 {
		t.Error("Expected hashed password but got empty")
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			err = CheckPasswordHash(tc.input, hashed)

			got := err == nil
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Error(diff)
			}
		})
	}
}

var secretKey = "testmakejwt"

func TestMakeJWT(t *testing.T) {
	type Input struct {
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}

	validUUID := uuid.New()

	tests := map[string]struct {
		input Input
		want  bool
	}{
		"Valid Input": {
			input: Input{
				userID:      validUUID,
				tokenSecret: secretKey,
				expiresIn:   time.Hour,
			},
			want: true,
		},
		"Invalid Secret": {
			input: Input{
				userID:      validUUID,
				tokenSecret: "invalid secret",
				expiresIn:   time.Hour,
			},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			jwtSecret, err := MakeJWT(tc.input.userID, tc.input.tokenSecret, tc.input.expiresIn)

			if err != nil {
				t.Fatalf("MakeJWT failed: %v", err)
			}

			_, err = ValidateJWT(jwtSecret, secretKey)

			got := err == nil

			diff := cmp.Diff(tc.want, got)

			if diff != "" {
				t.Error(diff)
			}
		})
	}

}

func TestValidateJWT(t *testing.T) {

	validTokenString, err := MakeJWT(uuid.New(), secretKey, time.Hour)

	if err != nil {
		t.Fatal("failed to make jwt ")
	}

	invalidTokenSecret, err := MakeJWT(uuid.New(), "invalid token string", time.Hour)

	if err != nil {
		t.Fatal("failed to make jwt ")
	}

	tests := map[string]struct {
		input string
		want  bool
	}{

		"Valid JWT": {
			input: validTokenString,
			want:  true,
		},
		"Invalid jwt secret": {
			input: invalidTokenSecret,
			want:  false,
		},
		"Invalid Issuer": {
			input: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODUxNDA5ODQsImlhdCI6MTQ4NTEzNzM4NCwiaXNzIjoiYWNtZS5jb20iLCJzdWIiOiIyOWFjMGMxOC0wYjRhLTQyY2YtODJmYy0wM2Q1NzAzMThhMWQiLCJhcHBsaWNhdGlvbklkIjoiNzkxMDM3MzQtOTdhYi00ZDFhLWFmMzctZTAwNmQwNWQyOTUyIiwicm9sZXMiOltdfQ.Mp0Pcwsz5VECK11Kf2ZZNF_SMKu5CgBeLN9ZOP04kZo",
			want:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateJWT(tc.input, secretKey)

			got := err == nil

			diff := cmp.Diff(tc.want, got)

			if diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := map[string]struct {
		input http.Header
		want  string
	}{
		"No Authorization Header": {
			input: http.Header{},
			want:  "",
		},
		"No token": {
			input: http.Header{
				"Authorization": []string{"Bearer "},
			},
			want: "",
		},
		"Invalid Grant": {
			input: http.Header{
				"Authorization": []string{"Bea something something"},
			},
			want: "",
		},
		"Success": {
			input: http.Header{
				"Authorization": []string{"Bearer somethingsomething"},
			},
			want: "somethingsomething",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ApiKey, _ := GetBearerToken(tc.input)
			diff := cmp.Diff(tc.want, ApiKey)
			if diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedLength := 64
	if len(token) != expectedLength {
		t.Errorf("Expected token length %d, got %d", expectedLength, len(token))
	}
}
