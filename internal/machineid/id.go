package machineid

import "fmt"

// ID returns the platform specific machine id of the current host OS.
func ID() (string, error) {
	id, err := readPlatformMachineID()
	if err != nil {
		return "", fmt.Errorf("machineid: %v", err)
	}
	return id, nil
}
