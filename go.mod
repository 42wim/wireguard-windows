module golang.zx2c4.com/wireguard/windows

require (
	github.com/Microsoft/go-winio v0.4.11
	golang.org/x/crypto v0.0.0-20190211182817-74369b46fc67
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	golang.org/x/sys v0.0.0-20190220154126-629670e5acc5
	golang.zx2c4.com/winipcfg latest
	golang.zx2c4.com/wireguard latest
)

replace (
	github.com/lxn/walk => golang.zx2c4.com/wireguard/windows pkg/walk
	github.com/lxn/win => golang.zx2c4.com/wireguard/windows pkg/walk-win
)
