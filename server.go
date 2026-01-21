package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const secretKey string = "secret"

type relayStore struct {
	local    *atomic.Value
	machines sync.Map
}

func newRelayStore() *relayStore {
	store := &relayStore{
		local: &atomic.Value{},
	}
	store.local.Store([]byte("{}"))
	return store
}

func (store *relayStore) set(machine string, data []byte) {
	if machine == "" {
		store.local.Store(data)
		return
	}

	value, _ := store.machines.LoadOrStore(machine, &atomic.Value{})
	value.(*atomic.Value).Store(data)
}

func (store *relayStore) get(machine string) []byte {
	if machine == "" {
		return store.local.Load().([]byte)
	}

	value, ok := store.machines.Load(machine)
	if !ok {
		return []byte("{}")
	}

	return value.(*atomic.Value).Load().([]byte)
}

func startServer(port int, secret string, relayEnabled bool, relayLocal bool, in chan output) *http.Server {
	return startServerWith(port, secret, relayEnabled, relayLocal, in, func(server *http.Server) error {
		return server.ListenAndServe()
	}, time.Sleep)
}

func startServerWith(port int, secret string, relayEnabled bool, relayLocal bool, in chan output, listenAndServe func(*http.Server) error, sleep func(time.Duration)) *http.Server {
	store := newRelayStore()

	go func() {
		if in == nil {
			return
		}

		data := output{}

		for d := range in {
			data = d
			jsonData, err := json.MarshalIndent(data, "", "\t")
			if err != nil {
				log.Println("Error marshalling data:", err)
				continue
			}
			store.set("", jsonData)
		}

		log.Println("Server data receiver stopped")
	}()

	mux := buildMux(secret, relayEnabled, relayLocal, store)
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

func buildMux(secret string, relayEnabled bool, relayLocal bool, store *relayStore) *http.ServeMux {
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
		machineName := r.URL.Query().Get("machine_name")
		if relayEnabled && machineName == "" && !relayLocal {
			http.Error(w, "machine_name is required", http.StatusBadRequest)
			return
		}
		w.Write(store.get(machineName))
	}))

	mux.HandleFunc("/r", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Refresh", "5")
		machineName := r.URL.Query().Get("machine_name")
		if relayEnabled && machineName == "" && !relayLocal {
			http.Error(w, "machine_name is required", http.StatusBadRequest)
			return
		}
		w.Write(store.get(machineName))
	}))

	if relayEnabled {
		mux.HandleFunc("/relay", secureWrapper(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			machineName := r.URL.Query().Get("machine_name")
			if machineName == "" {
				http.Error(w, "machine_name is required", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			if !json.Valid(body) {
				http.Error(w, "invalid json payload", http.StatusBadRequest)
				return
			}

			store.set(machineName, body)
			w.WriteHeader(http.StatusAccepted)
		}))
	}

	return mux
}
