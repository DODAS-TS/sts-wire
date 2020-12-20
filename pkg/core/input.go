package core

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/awnumar/memguard"
	"github.com/gookit/color"
	"github.com/rs/zerolog/log"
)

type GetInputWrapper struct {
	Scanner bufio.Reader
}

var (
	errPasswordMismatch = errors.New("The two password inserted are not the same.")
)

func (t *GetInputWrapper) GetPassword(question string, only4Decription bool) (password *memguard.Enclave, err error) {
	fmt.Print(question)
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, fmt.Errorf("get password %w", err)
	}

	if only4Decription {
		passEnclave := memguard.NewEnclave(bytePassword)
		return passEnclave, nil
	}

	fmt.Println()
	passMsg := fmt.Sprintf("%s Please, insert the password again: ", color.Yellow.Sprint("==>"))
	fmt.Print(passMsg)

	bytePassword2, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("get password check %w", err)
	}
	if string(bytePassword) == string(bytePassword2) {
		passEnclave := memguard.NewEnclave(bytePassword)
		return passEnclave, nil
	}

	log.Err(errPasswordMismatch).Msg("GetPassword")
	return nil, errPasswordMismatch
}

func (t *GetInputWrapper) GetInputString(question string, def string) (text string, err error) {
	if def != "" {
		fmt.Print(question + "\n" + "press enter for default [" + def + "]\n")
		text, err = t.Scanner.ReadString('\n')
		if err != nil {
			return "", err
		}
		text = strings.Replace(text, "\r\n", "", -1)
		text = strings.Replace(text, "\n", "", -1)

		if text == "" {
			text = def
		}

	} else {
		fmt.Print(question + "\n")

		text, err = t.Scanner.ReadString('\n')
		if err != nil {
			return "", err
		}
		text = strings.Replace(text, "\n", "", -1)
	}

	return text, nil
}
