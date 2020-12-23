// +build windows

package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"

	"github.com/awnumar/memguard"
	"github.com/gookit/color"
	"github.com/rs/zerolog/log"
)

var (
	errPasswordMismatch = errors.New("The two password inserted are not the same.")
)

// passwordReader is an io.Reader that reads from a specific file descriptor.
type passwordReader int

func readPassword(fd int) (*os.File, error) {

	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return nil, err
	}
	old := st

	st &^= (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT)
	st |= (windows.ENABLE_PROCESSED_OUTPUT | windows.ENABLE_PROCESSED_INPUT)
	if err := windows.SetConsoleMode(windows.Handle(fd), st); err != nil {
		return nil, err
	}

	defer windows.SetConsoleMode(windows.Handle(fd), old)

	var h windows.Handle
	p, _ := windows.GetCurrentProcess()
	if err := windows.DuplicateHandle(p, windows.Handle(fd), p, &h, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(h), "stdin")
	return f, nil
}

func (t *GetInputWrapper) GetPassword(question string, only4Decription bool) (password *memguard.Enclave, err error) {
	fmt.Print(question)

	readPasswdFd, errCreateReader := readPassword(int(syscall.Stdin))
	if errCreateReader != nil {
		return nil, fmt.Errorf("get password %w", errCreateReader)
	}

	defer readPasswdFd.Close()

	passEnclave, errEclBuf := memguard.NewBufferFromReaderUntil(readPasswdFd, '\n')
	if errEclBuf != nil {
		return nil, fmt.Errorf("get password enclave %w", errEclBuf)
	}

	if only4Decription {
		password = passEnclave.Seal()

		return password, nil
	}

	fmt.Println()
	passMsg := fmt.Sprintf("%s Please, insert the password again: ", color.Yellow.Sprint("==>"))
	fmt.Print(passMsg)

	passEnclave2, err := memguard.NewBufferFromReaderUntil(readPasswdFd, '\n')
	if err != nil {
		return nil, fmt.Errorf("get password check %w", err)
	}

	if bytes.Equal(passEnclave.Bytes(), passEnclave2.Bytes()) {
		password = passEnclave.Seal()
		return password, nil
	}

	log.Err(errPasswordMismatch).Msg("GetPassword")
	return nil, errPasswordMismatch
}
