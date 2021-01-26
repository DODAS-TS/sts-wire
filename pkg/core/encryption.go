package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/denisbrodbeck/machineid"
	"github.com/rs/zerolog/log"
)

func tryContainerMachineID() (machineID string, err error) {
	// Ref: https://stackoverflow.com/questions/23513045/how-to-check-if-a-process-is-running-inside-docker-container
	cgroupFile, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return machineID, fmt.Errorf("cannot open cgroup: %w", err)
	}

	defer cgroupFile.Close()

	var buff bytes.Buffer

	_, err = buff.ReadFrom(cgroupFile)
	if err != nil {
		return machineID, fmt.Errorf("cannot read cgroup: %w", err)
	}

	for _, line := range strings.Split(buff.String(), "\n") {
		if strings.Contains(line, "/docker/") {
			parts := strings.Split(line, "/docker/")
			if len(parts) != 2 {
				return machineID, fmt.Errorf("not a valid docker container id: %w", err)
			}

			machineID = parts[1]

			break
		}
	}

	if machineID == "" {
		return machineID, fmt.Errorf("docker container id not found: %w", err)
	}

	return machineID, nil
}

func CreateHash(key string) string {
	log.Debug().Msg("create hash")

	id, errID := machineid.ProtectedID("sts-wire")
	if errID != nil {
		if strings.Contains(errID.Error(), "open /etc/machine-id: no such file or directory") {
			// TODO: get a unique uuid for containers:
			// Ref: https://github.com/denisbrodbeck/machineid/issues/5
			// Ref: https://github.com/panta/machineid/blob/master/id_linux.go
			id, errID = tryContainerMachineID()
			if errID != nil {
				id = "notAMachine"
				log.Debug().Str("machineID", id).Msg("Cannot find a docker container id")
			} else {
				log.Debug().Str("machineID", id).Msg("Found docker container id")
			}
		} else {
			panic(errID)
		}
	}

	hasher := hmac.New(md5.New, []byte(id))
	_, errWrite := hasher.Write([]byte(key))

	if errWrite != nil {
		panic(errWrite)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data []byte, password *memguard.Enclave) []byte {
	log.Debug().Msg("encryption - open enclave")

	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}

	defer passphrase.Destroy() // Destroy the copy when we return

	log.Debug().Msg("encryption - create cipher")

	block, errNewCiper := aes.NewCipher([]byte(CreateHash(string(passphrase.Bytes()))))
	if errNewCiper != nil {
		panic(errNewCiper)
	}

	log.Debug().Msg("encryption - create block")

	gcm, errNewGCM := cipher.NewGCM(block)
	if errNewGCM != nil {
		panic(errNewGCM.Error())
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	log.Debug().Msg("encryption - encode")

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext
}

func Decrypt(data []byte, password *memguard.Enclave) []byte {
	log.Debug().Msg("decryption - open enclave")

	passphrase, errOpenEnclave := password.Open()
	if errOpenEnclave != nil {
		memguard.SafePanic(errOpenEnclave)
	}

	defer passphrase.Destroy() // Destroy the copy when we return

	log.Debug().Msg("decryption - create key")

	key := []byte(CreateHash(string(passphrase.Bytes())))

	log.Debug().Msg("decryption - create cipher")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	log.Debug().Msg("decryption - create block")

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	log.Debug().Msg("decryption - decode")

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return plaintext
}
