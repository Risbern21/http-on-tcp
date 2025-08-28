package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '1' && ch <= '9' {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':

			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

type Headers struct {
	headers map[string]string
}

func (h *Headers) Get(key string) (string, bool) {
	value, exists := h.headers[strings.ToLower(key)]
	return value, exists
}

func (h *Headers) Set(key string, value string) {
	key = strings.ToLower(key)

	if v, ok := h.headers[key]; ok {
		h.headers[key] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[key] = value
	}
}

func (h *Headers) Replace(key, value string) {
	key= strings.ToLower(key)
	h.headers[key] = value
}

func (h *Headers) ForEach(cb func(k, v string)) {
	for k, v := range h.headers {
		cb(k, v)
	}
}

var rn = []byte("\r\n")

func parseHeader(fieldLine []byte) (string, string, error) {

	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	key := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field key")
	}

	return string(key), string(value), nil
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		key, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(key)) {
			return 0, false, fmt.Errorf("invalid characters in field name")
		}

		read += idx + len(rn)
		h.Set(key, value)
	}

	return read, done, nil
}
