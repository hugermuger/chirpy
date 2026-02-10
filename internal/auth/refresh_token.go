package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	ran := make([]byte, 32)
	rand.Read(ran)
	return hex.EncodeToString(ran), nil
}
