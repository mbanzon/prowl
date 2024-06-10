package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func startServer(port int, in chan output) *http.Server {
	data := output{}

	go func() {
		for d := range in {
			data = d
		}

		log.Println("Server data receiver stopped")
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		for {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Println("Error running server:", err)
				log.Println("Waiting for 5 seconds before retrying to start server...")
				time.Sleep(5 * time.Second)
			} else {
				log.Println("Server stopped")
				return
			}
		}
	}()

	return server
}
