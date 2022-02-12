package easy

import "net"

// GetOutboundIP returns the preferred outbound ip of this machine.
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "10.20.30.40:56789")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
