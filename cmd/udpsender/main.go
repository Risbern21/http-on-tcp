package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	conn, err := net.DialUDP("udp", nil,addr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for {
		fmt.Print(">")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		conn.Write([]byte(input))
	}
}
