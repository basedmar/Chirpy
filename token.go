package main

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	dat := make([]byte, 32)
	rand.Read(dat)
	value := hex.EncodeToString(dat)
	return value
}
