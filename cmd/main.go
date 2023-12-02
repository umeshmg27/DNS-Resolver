package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/umeshmg27/dns-resolver/resolver" // This imports your DNS resolver package
)

func main() {
	port := flag.Int("port", 53, "The port number to listen on for DNS queries")
	flag.Parse()

	// Set up a UDP network listener for the DNS server
	addr := net.UDPAddr{
		Port: *port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to set up UDP listener: %v", err)
	}
	defer conn.Close()

	fmt.Printf("DNS Server is listening on port %d\n", *port)

	// Infinite loop to handle incoming DNS queries
	for {
		err := resolver.HandleDNSRequest(conn)
		if err != nil {
			log.Printf("Error handling DNS request: %v", err)
		}
	}
}
