package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/oatsaysai/test-go-server/webserver"
)

var port = 8888

var client *http.Client

var serverURL = "http://localhost:8888"
var path = "/test"

func main() {

	// config := webserver.Default()

	var s = webserver.NewServer(webserver.Default())

	s.Run()
}

func main2() {

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	m := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: m,
	}

	m.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(""))
	})

	m.HandleFunc("/test_call", func(w http.ResponseWriter, req *http.Request) {

		newHTTPRequest(
			serverURL,
			"POST",
			path,
			nil,
		)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(""))
	})

	fmt.Printf("Serving Internal (GRPC) API at: %d\n", port)
	srv.ListenAndServe()
	defer func() {
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.Fatalf("http server shutdown error: %v", err)
		}
	}()
}

func newHTTPRequest(apiAddr, method, path string, jsonBody []byte) (int, []byte, error) {
	var URL *url.URL
	URL, err := url.Parse(apiAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += path
	parameters := url.Values{}
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest(method, encodedURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println(err.Error())
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	// defer resp.Body.Close()
	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return resp.StatusCode, nil, nil
}
