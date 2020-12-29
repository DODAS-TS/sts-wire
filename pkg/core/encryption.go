package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog/log"
)

func CreateHash(key string) string {
	log.Info().Msg("create hash")

	hasher := hmac.New(md5.New, []byte("sts-wire"))
	_, errWrite := hasher.Write([]byte(key))

	if errWrite != nil {
		panic(errWrite)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data []byte, password *memguard.Enclave) []byte {
	log.Info().Msg("encryption - open enclave")

	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}

	defer passphrase.Destroy() // Destroy the copy when we return

	log.Info().Msg("encryption - create cipher")

	block, errNewCiper := aes.NewCipher([]byte(CreateHash(string(passphrase.Bytes()))))
	if errNewCiper != nil {
		panic(errNewCiper)
	}

	log.Info().Msg("encryption - create block")

	gcm, errNewGCM := cipher.NewGCM(block)
	if errNewGCM != nil {
		panic(errNewGCM.Error())
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	log.Info().Msg("encryption - encode")

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext
}

func Decrypt(data []byte, password *memguard.Enclave) []byte {
	log.Info().Msg("decryption - open enclave")

	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}

	defer passphrase.Destroy() // Destroy the copy when we return

	log.Info().Msg("decryption - create key")

	key := []byte(CreateHash(string(passphrase.Bytes())))

	log.Info().Msg("decryption - create cipher")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	log.Info().Msg("decryption - create block")

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	log.Info().Msg("decryption - decode")

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return plaintext
}
