package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog/log"
)

func CreateHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data []byte, password *memguard.Enclave) []byte {
	log.Info().Msg("encryption - open enlcave")
	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}
	defer passphrase.Destroy() // Destroy the copy when we return

	block, _ := aes.NewCipher([]byte(CreateHash(string(passphrase.Bytes()))))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func Decrypt(data []byte, password *memguard.Enclave) []byte {
	log.Info().Msg("decryption - open enlcave")
	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}
	defer passphrase.Destroy() // Destroy the copy when we return

	fmt.Println(string(passphrase.Bytes()))

	key := []byte(CreateHash(string(passphrase.Bytes())))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
