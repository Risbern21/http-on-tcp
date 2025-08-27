package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"http.go/internal/headers"
)

type parserState string

const (
	stateInit    parserState = "init"
	stateDone    parserState = "done"
	stateError   parserState = "error"
	stateHeaders parserState = "headers"
	stateBody    parserState = "body"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
	Body        string
}

func getInt(headers *headers.Headers, key string, defaultValue int) int {
	v, exists := headers.Get(key)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}

	return value
}

func (r *Request)hasBody()bool{
	length := getInt(r.Headers, "content-length", 0)
	return length > 0 
}

var ErrorMalformedRequest = fmt.Errorf("malformed request-line")
var ErrorIncompleteStartLine = fmt.Errorf("incomplete start line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData)==0{
			break outer
		}

		switch r.state {
		case stateError:
			return read, ErrorRequestInErrorState

		case stateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = stateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n
			r.RequestLine = *rl

			r.state = stateHeaders

		case stateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = stateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody(){
					r.state = stateBody
				}else{
					r.state=stateDone
				}
			}

		case stateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				panic("chunked not implemented")
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = stateDone
			}

		case stateDone:
			break outer

		default:
			panic("somehow we fucked up")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == stateDone || r.state == stateError
}

func NewRequest() *Request {
	return &Request{
		state:   stateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequest
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
