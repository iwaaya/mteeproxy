package mteeproxy

import (
	"fmt"
	"net"
	"net/http"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("net.Listen failed: %s", err)
	}

	tee := Handler{
		Target:      "localhost:4000",
		Alternative: []string{"5000", "6000"},
	}

	err = http.Serve(listen, tee)
	if err != nil {
		fmt.Printf("http.Serve failed: %s", err)
	}
}
