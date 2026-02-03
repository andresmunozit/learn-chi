package main

import (
	"crypto/rand"
	"encoding/hex"
)

func generateId(len int) (string, error) {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), err
}
