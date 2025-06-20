package encryption

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestAES256Encrypt(t *testing.T) {
	apiKey := []byte{111, 231, 189, 121, 208, 13, 1, 227, 244, 152, 242, 194, 22, 8, 1, 9,
		178, 69, 94, 170, 9, 113, 128, 116, 117, 24, 173, 100, 6, 51, 64, 196}

	hashKey := []byte{108, 2, 40, 130, 254, 142, 63, 49, 217, 189, 56, 33, 174, 30, 110, 240,
		125, 122, 11, 116, 239, 138, 105, 172, 246, 230, 129, 4, 213, 0, 101, 156}

	hashedApiKey, err := HMAC256Hash(hashKey, apiKey)
	if err != nil {
		t.Fatal(err)
	}

	apiKeyBase64 := base64.RawStdEncoding.EncodeToString(apiKey)
	hashKeyBase64 := base64.RawStdEncoding.EncodeToString(hashKey)
	hashedApiKeyBase64 := base64.RawStdEncoding.EncodeToString(hashedApiKey)

	fmt.Println("API key:", apiKeyBase64)
	fmt.Println("Hash key:", hashKeyBase64)
	fmt.Println("Hashed API key:", hashedApiKeyBase64)

	fmt.Println("womp womp")
}
