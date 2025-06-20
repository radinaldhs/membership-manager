package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

func AES256Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

func AES256Decrypt(cipherData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherData := cipherData[:nonceSize], cipherData[nonceSize:]

	return gcm.Open(nil, nonce, cipherData, nil)
}

func HMAC256Hash(key, message []byte) ([]byte, error) {
	hash := hmac.New(sha256.New, key)
	_, err := hash.Write(message)
	if err != nil {
		return nil, err
	}

	sum := hash.Sum(nil)

	return sum, nil
}
