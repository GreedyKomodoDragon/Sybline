package auth

import (
	"crypto/sha512"
	"encoding/hex"
)

func GenerateHash(input string, salt string) string {
	sha := sha512.New()
	sha.Write([]byte(salt + input)) // Concatenate salt and input

	return hex.EncodeToString(sha.Sum(nil))
}
