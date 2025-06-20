package random

import (
	"crypto/rand"
	"math/big"
)

const numericCharset = "0123456789"

func GetRandomNumerics(length int) string {
	buffer := make([]byte, length)
	for i := range buffer {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numericCharset)-1)))
		buffer[i] = numericCharset[num.Int64()]
	}

	return string(buffer)
}
