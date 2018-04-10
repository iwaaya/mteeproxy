package mteeproxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Handler struct {
	Target      string
	Alternative []string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	dupReqs := duplicateRequest(req, len(h.Alternative)+1)

	// alternative
	for i, dr := range dupReqs[1:] {
		go func(req *http.Request, i int) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered in ServeHTTP(alternative request) from: %s", r)
				}
			}()

			_, err := handleRequest(req, h.Alternative[i])
			if err != nil {
				fmt.Printf("alternative port: %s", err)
			}
		}(dr, i)
	}

	// target
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in ServeHTTP(target request) from: %s", r)
		}
	}()

	resp, err := handleRequest(dupReqs[0], h.Target)
	if err != nil {
		fmt.Printf("target port: %s", err)
	}
	if resp != nil {
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		if _, err = io.Copy(w, resp.Body); err != nil {
			fmt.Printf("target port: io.Copy failed: %s", err)
		}
	}
}

func handleRequest(req *http.Request, host string) (resp *http.Response, err error) {
	URL, err := url.Parse("http://" + host + req.URL.String())
	if err != nil {
		return nil, nil
	}
	req.URL = URL
	transport := &http.Transport{}
	return transport.RoundTrip(req)
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func duplicateRequest(req *http.Request, number int) (dupReqs []*http.Request) {
	var buffers = make([]io.Writer, number)
	for i := 0; i < number; i++ {
		buffers[i] = new(bytes.Buffer)
	}

	w := io.MultiWriter(buffers...)
	io.Copy(w, req.Body)

	for i := 0; i < number; i++ {
		b := buffers[i].(*bytes.Buffer)
		req := &http.Request{
			Method:        req.Method,
			URL:           req.URL,
			Proto:         "HTTP.1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        req.Header,
			Body:          nopCloser{b},
			Host:          req.Host,
			ContentLength: req.ContentLength,
		}
		dupReqs = append(dupReqs, req)
	}
	return
}
