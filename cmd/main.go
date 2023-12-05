package main

import (
	"flag"
	"fmt"

	"github.com/umeshmg27/dns-resolver/resolver"
	// This imports your DNS resolver package
)

func main() {
	// port := flag.Int("port", 53, "The port number to listen on for DNS queries")
	// flag.Parse()
	// fmt.Printf("\n\n port %+v", port)
	domain := flag.String("domain", "google.com", "Domain name to find the IP address")
	flag.Parse()
	nameServer := flag.String("nameServer", "192.168.6.181", "Domain name to find the IP address")
	flag.Parse()
	result := []string{}
	result, _, err := resolver.HandleDNSRequest(*domain, result, *nameServer)
	if err != nil {
		fmt.Printf("\n err %+v", err)
	}
	fmt.Printf("\n DNS output %+v", result)
}
