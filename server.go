package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const secretKey string = "secret"

func startServer(port int, secret string, in chan output) *http.Server {
	cachedData := []byte("{}")

	go func() {
		data := output{}

		for d := range in {
			data = d
			jsonData, err := json.MarshalIndent(data, "", "\t")
			if err != nil {
				log.Println("Error marshalling data:", err)
				continue
			}
			cachedData = jsonData
		}

		log.Println("Server data receiver stopped")
	}()

	secureWrapper := func(f http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret != "" {
				passedSecret := r.URL.Query().Get(secretKey)
				if secret != passedSecret {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}

			f(w, r)
		})
	}

	http.HandleFunc("/", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedData)
	}))

	http.HandleFunc("/r", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Refresh", "5")
		w.Write(cachedData)
	}))

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		log.Println("Server started on port", port)

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
