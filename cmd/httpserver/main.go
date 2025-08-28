package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http.go/internal/headers"
	"http.go/internal/request"
	"http.go/internal/response"
	"http.go/internal/server"
)

const port = 42069

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func respond400() []byte {
	return []byte(`<html>
  	<head>
    	<title>400 Bad Request</title>
  	</head>
  	<body>
    	<h1>Bad Request</h1>
    	<p>Your request honestly kinda sucked.</p>
  	</body>
	</html>`)
}

func respond500() []byte {
	return []byte(`<html>
	<head>
	    <title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
	    <p>Okay, you know what? This one is on me.</p>
	</body>
	</html>
`)
}

func respond200() []byte {
	return []byte(`<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
	    <h1>Success!</h1>
	    <p>Your request was an absolute banger.</p>
	</body>
	</html>`)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, r *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		if r.RequestLine.RequestTarget == "/yourproblem" {
			body = respond400()
			status = response.StatusBadRequest
		} else if r.RequestLine.RequestTarget == "/myproblem" {
			body = respond500()
			status = response.StatusInternalServerError
		} else if strings.HasPrefix(r.RequestLine.RequestTarget, "/video") {
			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				w.WriteHeaders(h)
				body = respond500()
				w.WriteBody(body)
				return
			}

			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(f)
			return 

		} else if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			target := r.RequestLine.RequestTarget
			r, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)

				h.Delete("content-length")
				h.Set("content-encoding", "chunked")
				h.Replace("Content-Type", "text/plain")

				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := r.Body.Read(data)
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)
					//write the hex
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					//write the data
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				trailers := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailers.Set("X-Content-SHA256", toStr(out[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(trailers)
				w.WriteBody([]byte("\r\n"))

				return
			}
		}

		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(body)
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
