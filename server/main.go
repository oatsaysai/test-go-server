package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

var port = 8888

func main() {
	m := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: m,
	}

	m.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
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
