package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
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

func DeleteMedia(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		log.Println("File already deleted form os")
		return nil
	}

	err := os.Remove(filepath)

	if err != nil {
		return err
	}

	return nil
}

func EnsureAssetsDir(root string) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return os.Mkdir(root, 0750)
	}
	return nil
}
