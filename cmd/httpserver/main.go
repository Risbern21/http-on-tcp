package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"http.go/internal/request"
	"http.go/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, r *request.Request) *server.HandlerError {
		if r.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: 400,
				Message: "Your problem is not my problem\n",
			}
		} else if r.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: 500,
				Message: "Woopsie, my bad\n",
			}
		} else {
			w.Write([]byte("all good frfr\n"))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
