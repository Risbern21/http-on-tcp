package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFooFoo:    barbar   \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	host, ok := headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)

	FooFoo,ok:=headers.Get("FooFoo")
	assert.True(t,ok)
	assert.Equal(t, "barbar", FooFoo)

	MissingKey,ok:=headers.Get("MissingKey")
	assert.False(t,ok)
	assert.Equal(t, "", MissingKey)
	assert.Equal(t, 47, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	//Test: Invalid character
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nFooFoo:    barbar   \r\nFooFoo: barbar\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	host, ok = headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)

	FooFoo, ok = headers.Get("FooFoo")
	assert.True(t, ok)
	assert.Equal(t, "barbar,barbar", FooFoo)

	MissingKey, ok = headers.Get("MissingKey")
	assert.False(t, ok)
	assert.Equal(t, "", MissingKey)
	assert.Equal(t, 63, n)
	assert.True(t, done)

}
