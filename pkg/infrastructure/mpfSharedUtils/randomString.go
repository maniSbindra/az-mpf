package mpfSharedUtils

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := big.NewInt(int64(len(charset)))

	randomString := make([]byte, length)
	for i := range randomString {
		randomIndex, _ := rand.Int(rand.Reader, charsetLength)
		randomString[i] = charset[randomIndex.Int64()]
	}

	return string(randomString)
}
