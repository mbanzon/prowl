package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

const secretKey string = "secret"

func startServer(port int, secret string, in chan output) *http.Server {
	return startServerWith(port, secret, in, func(server *http.Server) error {
		return server.ListenAndServe()
	}, time.Sleep)
}

func startServerWith(port int, secret string, in chan output, listenAndServe func(*http.Server) error, sleep func(time.Duration)) *http.Server {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte("{}"))

	go func() {
		data := output{}

		for d := range in {
			data = d
			jsonData, err := json.MarshalIndent(data, "", "\t")
			if err != nil {
				log.Println("Error marshalling data:", err)
				continue
			}
			cachedData.Store(jsonData)
		}

		log.Println("Server data receiver stopped")
	}()

	mux := buildMux(secret, cachedData)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Println("Server started on port", port)

		for {
			err := listenAndServe(server)
			if err != nil && err != http.ErrServerClosed {
				log.Println("Error running server:", err)
				log.Println("Waiting for 5 seconds before retrying to start server...")
				sleep(5 * time.Second)
			} else {
				log.Println("Server stopped")
				return
			}
		}
	}()

	return server
}

func buildMux(secret string, cachedData *atomic.Value) *http.ServeMux {
	mux := http.NewServeMux()

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

	mux.HandleFunc("/", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedData.Load().([]byte))
	}))

	mux.HandleFunc("/r", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Refresh", "5")
		w.Write(cachedData.Load().([]byte))
	}))

	return mux
}
