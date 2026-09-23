package utils

import (
	"crypto/rand"
	"math/big"
)

// TokenCharset contains unambiguous uppercase letters and digits (omits 0, O, 1, I, L)
const TokenCharset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// GenerateRandomExamToken generates a cryptographically secure, unambiguous alphanumeric token.
func GenerateRandomExamToken(length int) string {
	if length <= 0 {
		length = 6
	}
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(TokenCharset)))
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = TokenCharset[i%len(TokenCharset)]
			continue
		}
		result[i] = TokenCharset[num.Int64()]
	}
	return string(result)
}
