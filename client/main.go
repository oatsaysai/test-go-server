package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	_ = iota
	INCREASE_THREAD
	DEFAULT_THREAD
)

const (
	_ = iota
	ResetCount
	Add
	SendCount
)

var serverURL = "http://localhost:8888"
var path = "/test"

var client *http.Client

func main() {

	var threads = flag.Int("threads", 10, "Threads: 0 for auto")
	var duration = flag.Int("duration", 10, "Duration of the test")

	flag.Parse()

	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns: 100,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	benchmarkTransferNo(*threads, *duration)
}

func benchmarkTransferNo(threads, duration int) {
	countChan := make(chan int)
	statChan := make(chan int)

	for i := 0; i < threads; i++ {
		go forTransfer(countChan)
	}

	go func(countChan, statChan chan int) {
		count := 0
		for {
			action := <-countChan
			if action == Add {
				count++
			} else if action == ResetCount {
				count = 0
			} else if action == SendCount {
				statChan <- count
			}
		}
	}(countChan, statChan)

	go func(duration int) {
		wait := time.Tick(time.Second * time.Duration(duration))
		_ = <-wait
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}(duration)

	tick := time.Tick(time.Second)
	for range tick {
		countChan <- SendCount
		count := <-statChan
		countChan <- ResetCount

		fmt.Fprintf(os.Stderr, "%d\n", count)
	}
}

func forTransfer(outChan chan int) {
	for {

		params := []byte(fmt.Sprintf(
			`
			{
				"data": "0"
			}
			`,
		))

		httpStatusCode, _ := transfer(serverURL, params)
		if httpStatusCode == http.StatusCreated {
			outChan <- Add
		}
	}
}

func transfer(addr string, paramJSON []byte) (int, error) {
	httpStatusCode, _, err := newHTTPRequest(
		addr,
		"POST",
		path,
		paramJSON,
	)
	if err != nil {
		return 0, err
	}
	return httpStatusCode, nil
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
