package utils

import (
	"fmt"
	"net"
	"os"
)

func GetLocalIPAddressAndPort() (string, int) {
	// Get local IP address
	var localIP string
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting network interfaces: %v\n", err)
		os.Exit(1)
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				localIP = ipnet.IP.String()
				break
			}
		}
		if localIP != "" {
			break
		}
	}

	if localIP == "" {
		fmt.Println("Unable to determine local IP address; please check your network settings.")
		os.Exit(1)
	}

	// Get a random available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("Error getting random port: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	randomPort := listener.Addr().(*net.TCPAddr).Port

	return localIP, randomPort
}
