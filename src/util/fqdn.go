package util

import (
	"net"
	"os"
	"strings"
)

// GetFQDN returns the fully qualified hostname of the current host. If the fqdn
// can't be determined the hostname is returned.
func GetFQDN() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, nil
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 == nil {
			continue
		}

		ip, err := ipv4.MarshalText()
		if err != nil {
			continue
		}

		hosts, err := net.LookupAddr(string(ip))
		if err != nil || len(hosts) == 0 {
			continue
		}

		fqdn := hosts[0]

		// return fqdn without trailing dot
		return strings.TrimSuffix(fqdn, "."), nil
	}

	return hostname, nil
}
