package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func Random32Generator() (string, error) {
	number := make([]byte, 32)

	_, err := rand.Read(number)

	if err != nil {

		return "", err
	}

	return hex.EncodeToString(number), nil
}
