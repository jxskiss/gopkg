//go:build linux

package machineid

import (
	"os"
	"strings"
)

const (
	// dbusPath is the default path for dbus machine id.
	dbusPath = "/var/lib/dbus/machine-id"

	// dbusPathEtc is the default path for dbus machine id located in /etc.
	// Some systems (like Fedora 20) only know this path.
	// Sometimes it's the other way round.
	dbusPathEtc = "/etc/machine-id"

	// Use the boot_id generated on each boot as fallback.
	bootIDPath = "/proc/sys/kernel/random/boot_id"
)

func readPlatformMachineID() (string, error) {
	b, err := os.ReadFile(dbusPath)
	if err != nil || len(b) == 0 {
		b, err = os.ReadFile(dbusPathEtc)
		if err != nil || len(b) == 0 {
			b, err = ioutil.ReadFile(bootIDPath)
		}
	}
	return strings.TrimSpace(string(b)), err
}
