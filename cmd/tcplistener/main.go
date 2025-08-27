package main

import (
	"fmt"
	"log"
	"net"

	"http.go/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("err: %v", err)
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		r.Headers.ForEach(func(k, v string) {
			fmt.Printf("- %s: %s\n", k, v)
		})
		fmt.Printf("Body:\n%s", r.Body)
	}
}
