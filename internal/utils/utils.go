package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func Random32Generator() (string, error) {
	number := make([]byte, 32)

	_, err := rand.Read(number)

	if err != nil {

		return "", err
	}

	return hex.EncodeToString(number), nil
}

type Location struct {
	Name string  `json:"name" validate:"required"`
	Lat  float64 `json:"lat" validate:"required"`
	Lng  float64 `json:"lng" validate:"required"`
}

func FormatPoint(loc Location) string {
	return fmt.Sprintf("SRID=4326;POINT(%f %f)", loc.Lng, loc.Lat)
}
