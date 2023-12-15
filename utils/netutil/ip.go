package netutil

import "net"

// GetOutboundIP returns the preferred outbound ip of this machine.
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "1.2.3.4:56789")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
