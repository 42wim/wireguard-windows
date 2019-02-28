/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package service

import (
	"encoding/gob"
	"golang.zx2c4.com/wireguard/windows/conf"
	"net/rpc"
	"os"
)

type Tunnel struct {
	Name string
}

type TunnelState int

const (
	TunnelUnknown TunnelState = iota
	TunnelStarted
	TunnelStopped
	TunnelStarting
	TunnelStopping
	TunnelDeleting
)

type NotificationType int

const (
	TunnelChangeNotificationType NotificationType = iota
	TunnelsChangeNotificationType
)

var rpcClient *rpc.Client

type tunnelChangeCallback struct {
	cb func(tunnel string)
}

var tunnelChangeCallbacks = make(map[*tunnelChangeCallback]bool)

type tunnelsChangeCallback struct {
	cb func()
}

var tunnelsChangeCallbacks = make(map[*tunnelsChangeCallback]bool)

func InitializeIPCClient(reader *os.File, writer *os.File, events *os.File) {
	rpcClient = rpc.NewClient(&pipeRWC{reader, writer})
	go func() {
		decoder := gob.NewDecoder(events)
		for {
			var notificationType NotificationType
			err := decoder.Decode(&notificationType)
			if err != nil {
				return
			}
			switch notificationType {
			case TunnelChangeNotificationType:
				var tunnel string
				err := decoder.Decode(&tunnel)
				if err != nil || len(tunnel) == 0 {
					continue
				}
				for cb := range tunnelChangeCallbacks {
					cb.cb(tunnel)
				}
			case TunnelsChangeNotificationType:
				for cb := range tunnelsChangeCallbacks {
					cb.cb()
				}
			}
		}
	}()
}

func (t *Tunnel) StoredConfig() (c conf.Config, err error) {
	err = rpcClient.Call("ManagerService.StoredConfig", t.Name, &c)
	return
}

func (t *Tunnel) RuntimeConfig() (c conf.Config, err error) {
	err = rpcClient.Call("ManagerService.RuntimeConfig", t.Name, &c)
	return
}

func (t *Tunnel) Start() (TunnelState, error) {
	var state TunnelState
	return state, rpcClient.Call("ManagerService.Start", t.Name, &state)
}

func (t *Tunnel) Stop() (TunnelState, error) {
	var state TunnelState
	return state, rpcClient.Call("ManagerService.Stop", t.Name, &state)
}

func (t *Tunnel) Delete() (TunnelState, error) {
	var state TunnelState
	return state, rpcClient.Call("ManagerService.Delete", t.Name, &state)
}

func (t *Tunnel) State() (TunnelState, error) {
	var state TunnelState
	return state, rpcClient.Call("ManagerService.State", t.Name, &state)
}

func IPCClientNewTunnel(conf *conf.Config) (Tunnel, error) {
	var tunnel Tunnel
	return tunnel, rpcClient.Call("ManagerService.Create", *conf, &tunnel)
}

func IPCClientTunnels() ([]Tunnel, error) {
	var tunnels []Tunnel
	return tunnels, rpcClient.Call("ManagerService.Tunnels", 0, &tunnels)
}

func IPCClientQuit(stopTunnelsOnQuit bool) (bool, error) {
	var alreadyQuit bool
	return alreadyQuit, rpcClient.Call("ManagerService.Quit", stopTunnelsOnQuit, &alreadyQuit)
}

func IPCClientRegisterTunnelChange(cb func(tunnel string)) *tunnelChangeCallback {
	s := &tunnelChangeCallback{cb}
	tunnelChangeCallbacks[s] = true
	return s
}
func IPCClientUnregisterTunnelChange(cb *tunnelChangeCallback) {
	delete(tunnelChangeCallbacks, cb)
}
func IPCClientRegisterTunnelsChange(cb func()) *tunnelsChangeCallback {
	s := &tunnelsChangeCallback{cb}
	tunnelsChangeCallbacks[s] = true
	return s
}
func IPCClientUnregisterTunnelsChange(cb *tunnelsChangeCallback) {
	delete(tunnelsChangeCallbacks, cb)
}
