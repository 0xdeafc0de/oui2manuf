package main

import (
	"fmt"
	"os"

	"github.com/0xdeafc0de/oui2manuf/ouidb"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <MAC address>")
		os.Exit(1)
	}

	mac := os.Args[1]
	fmt.Println("Looking up MAC Address:", mac)

	manuf, err := ouidb.Lookup(mac)
	if err != nil {
		fmt.Println("Manufacturer not found for MAC:", mac)
	} else {
		fmt.Println("Manufacturer:", manuf)
	}
}
