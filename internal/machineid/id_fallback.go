//go:build !darwin && !linux && !freebsd && !windows

package machineid

import "errors"

func readPlatformMachineID() (string, error) {
	return "", errors.New("not implemented for this platform")
}
