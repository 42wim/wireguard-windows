/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"fmt"
	"os"
	"runtime"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

const wireguardUIClass = "WireGuard UI - MainWindow"

func raise(hwnd win.HWND) {
	if win.IsIconic(hwnd) {
		win.ShowWindow(hwnd, win.SW_RESTORE)
	}

	win.SetActiveWindow(hwnd)

	win.SetWindowPos(hwnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_SHOWWINDOW)
	win.SetForegroundWindow(hwnd)
	win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_SHOWWINDOW)
}

func RaiseUI() bool {
	hwnd := win.FindWindow(windows.StringToUTF16Ptr(wireguardUIClass), nil)
	if hwnd == 0 {
		return false
	}
	raise(hwnd)
	return true
}

func WaitForRaiseUIThenQuit() {
	var handle win.HWINEVENTHOOK
	runtime.LockOSThread()
	handle, err := win.SetWinEventHook(win.EVENT_OBJECT_CREATE, win.EVENT_OBJECT_CREATE, 0, func(hWinEventHook win.HWINEVENTHOOK, event uint32, hwnd win.HWND, idObject int32, idChild int32, idEventThread uint32, dwmsEventTime uint32) uintptr {
		class := make([]uint16, len(wireguardUIClass)+2) /* Plus 2, one for the null terminator, and one to see if this is only a prefix */
		n, err := win.GetClassName(hwnd, &class[0], len(class))
		if err != nil || n != len(wireguardUIClass) || windows.UTF16ToString(class) != wireguardUIClass {
			return 0
		}
		win.UnhookWinEvent(handle)
		raise(hwnd)
		os.Exit(0)
		return 0
	}, 0, 0, win.WINEVENT_SKIPOWNPROCESS|win.WINEVENT_OUTOFCONTEXT)
	if err != nil {
		walk.MsgBox(nil, "WireGuard Detection Error", fmt.Sprintf("Unable to wait for WireGuard window to appear: %v", err), walk.MsgBoxIconError)
	}
	for {
		var msg win.MSG
		if m := win.GetMessage(&msg, 0, 0, 0); m != 0 {
			win.TranslateMessage(&msg)
			win.DispatchMessage(&msg)
		}
	}
}
