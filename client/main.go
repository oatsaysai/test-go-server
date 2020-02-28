package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
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

	benchmarkTransferNo2(*threads, *duration)
}

func benchmarkTransferNo2(threads, duration int) {

	var sem = semaphore.NewWeighted(int64(threads))
	var wg sync.WaitGroup
	var successCount = 0
	var mu sync.Mutex
	var ctx = context.TODO()

	startTime := time.Now()

	for i := 0; i < 10009; i++ {
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go (func(currentI int) {
				sem.Acquire(ctx, 1)
				defer wg.Done()

				params := []byte(fmt.Sprintf(
					`
				{
					"data": "0"
				}
				`,
				))
				httpStatusCode, _ := transfer(serverURL, params)
				sem.Release(1)

				if httpStatusCode == http.StatusCreated {
					mu.Lock()
					successCount++
					mu.Unlock()
				}

			})(j)
		}
	}

	wg.Wait()
	fmt.Printf("Done benchmarking. %d successful transfers. Took %v = %f TPS\n", successCount, 0, float64(successCount)/time.Since(startTime).Seconds())
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
	httpStatusCode, _, err := makePost(
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

func makePost(apiAddr, method, path string, jsonBody []byte) (int, []byte, error) {
	var URL *url.URL
	URL, err := url.Parse(apiAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += path
	parameters := url.Values{}
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	resp, err := http.Post(encodedURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, nil, nil
}
