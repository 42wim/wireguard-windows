/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package updater

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/ipc/winpipe"
)

func runMsi(msiPath string, userToken uintptr) error {
	system32, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()
	attr := &os.ProcAttr{
		Sys: &syscall.SysProcAttr{
			Token: syscall.Token(userToken),
		},
		Files: []*os.File{devNull, devNull, devNull},
		Dir:   path.Dir(msiPath),
	}
	msiexec := path.Join(system32, "msiexec.exe")
	proc, err := os.StartProcess(msiexec, []string{msiexec, "/qb!-", "/i", path.Base(msiPath)}, attr)
	if err != nil {
		return err
	}
	state, err := proc.Wait()
	if err != nil {
		return err
	}
	if !state.Success() {
		return &exec.ExitError{ProcessState: state}
	}
	return nil
}

func msiTempFile() (*os.File, error) {
	var randBytes [32]byte
	n, err := rand.Read(randBytes[:])
	if err != nil {
		return nil, err
	}
	if n != int(len(randBytes)) {
		return nil, errors.New("Unable to generate random bytes")
	}
	sd, err := winpipe.SddlToSecurityDescriptor("O:SYD:PAI(A;;FA;;;SY)(A;;FR;;;BA)")
	if err != nil {
		return nil, err
	}
	sa := &windows.SecurityAttributes{
		Length:             uint32(len(sd)),
		SecurityDescriptor: uintptr(unsafe.Pointer(&sd[0])),
	}
	//TODO: os.TempDir() returns C:\windows\temp when calling from this context. Supposedly this is mostly secure
	// against TOCTOU, but who knows! Look into this!
	name := path.Join(os.TempDir(), hex.EncodeToString(randBytes[:]))
	name16 := windows.StringToUTF16Ptr(name)
	//TODO: it would be nice to specify delete_on_close, but msiexec.exe doesn't open its files with read sharing.
	fileHandle, err := windows.CreateFile(name16, windows.GENERIC_WRITE, windows.FILE_SHARE_READ, sa, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	runtime.KeepAlive(sd)
	if err != nil {
		return nil, err
	}
	windows.MoveFileEx(name16, nil, windows.MOVEFILE_DELAY_UNTIL_REBOOT)
	return os.NewFile(uintptr(fileHandle), name), nil
}
