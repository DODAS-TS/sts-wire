package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"golang.org/x/sys/windows"

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

// passwordReader is an io.Reader that reads from a specific file descriptor.
type passwordReader int

func (r passwordReader) Read(buf []byte) (int, error) {
	return unix.Read(int(r), buf)
}

const ioctlReadTermios = unix.TIOCGETA
const ioctlWriteTermios = unix.TIOCSETA

func readPassword(fd int) (*io.Read, error) {
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
	readPasswdFd, termios, errCreateReader := readPassword(syscall.Stdin)
	if errCreateReader != nil {
		return nil, fmt.Errorf("get password %w", errCreateReader)
	}

	defer readPasswdFd.Close()

	passEnclave, errEclBuf := memguard.NewBufferFromReaderUntil(readPasswdFd, '\n')
	if errEclBuf != nil {
		return nil, fmt.Errorf("get password enclave %w", errEclBuf)
	}
	password = passEnclave.Seal()

	if only4Decription {
		return password, nil
	}

	fmt.Println()
	passMsg := fmt.Sprintf("%s Please, insert the password again: ", color.Yellow.Sprint("==>"))
	fmt.Print(passMsg)

	bytePassword2, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, fmt.Errorf("get password check %w", err)
	}

	if bytes.Compare(passEnclave.Bytes(), bytePassword2) == 0 {
		return password, nil
	}

	log.Err(errPasswordMismatch).Msg("GetPassword")
	return nil, errPasswordMismatch
}
