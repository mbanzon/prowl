package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestBuildMuxConcurrentCachedDataAccess(t *testing.T) {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte(`{"time":0}`))

	server := httptest.NewServer(buildMux("", cachedData))
	defer server.Close()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				resp, err := http.Get(server.URL + "/")
				if err != nil {
					t.Errorf("unexpected request error: %v", err)
					return
				}
				var payload map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
					t.Errorf("invalid JSON response: %v", err)
					resp.Body.Close()
					return
				}
				resp.Body.Close()
			}
		}()
	}

	for i := 0; i < 500; i++ {
		payload, err := json.Marshal(output{Time: int64(i)})
		if err != nil {
			t.Fatalf("failed to marshal output: %v", err)
		}
		cachedData.Store(payload)
	}

	wg.Wait()
}

func TestBuildMuxRespectsSecret(t *testing.T) {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte(`{"time":123}`))

	server := httptest.NewServer(buildMux("topsecret", cachedData))
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, err = http.Get(server.URL + "/?secret=topsecret")
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK status, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(body) != `{"time":123}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestBuildMuxRefreshEndpoint(t *testing.T) {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte(`{"time":5}`))

	server := httptest.NewServer(buildMux("", cachedData))
	defer server.Close()

	resp, err := http.Get(server.URL + "/r")
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK status, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Refresh"); got != "5" {
		t.Fatalf("expected Refresh header 5, got %q", got)
	}
	resp.Body.Close()
}
