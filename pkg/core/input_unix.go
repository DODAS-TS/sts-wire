// +build !windows,linux darwin

package core

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/awnumar/memguard"
	"github.com/gookit/color"
	"github.com/rs/zerolog/log"
)

var (
	errPasswordMismatch = errors.New("The two password inserted are not the same.")
)

// passwordReader is an io.Reader that reads from a specific file descriptor.
type passwordReader int

func (r passwordReader) Read(buf []byte) (int, error) {
	return unix.Read(int(r), buf)
}

const ioctlReadTermios = unix.TIOCGETA
const ioctlWriteTermios = unix.TIOCSETA

func readPassword(fd int) (passwordReader, *unix.Termios, error) {
	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return -1, nil, fmt.Errorf("readPassword %w", err)
	}

	newState := *termios
	newState.Lflag &^= unix.ECHO
	newState.Lflag |= unix.ICANON | unix.ISIG
	newState.Iflag |= unix.ICRNL

	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &newState); err != nil {
		return -1, nil, fmt.Errorf("readPassword %w", err)
	}

	return passwordReader(fd), termios, nil
}

func (t *GetInputWrapper) GetPassword(question string, only4Decription bool) (password *memguard.Enclave, err error) {
	fmt.Print(question)

	readPasswdFd, termios, errCreateReader := readPassword(syscall.Stdin)
	if errCreateReader != nil {
		return nil, fmt.Errorf("get password %w", errCreateReader)
	}

	defer unix.IoctlSetTermios(int(readPasswdFd), ioctlWriteTermios, termios) // nolint: errcheck

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
