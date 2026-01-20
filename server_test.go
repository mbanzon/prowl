package main

import (
	"encoding/json"
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

func TestBuildMuxRejectsMissingSecret(t *testing.T) {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte(`{"time":42}`))

	mux := buildMux("topsecret", cachedData)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestBuildMuxAllowsSecretAndSetsRefresh(t *testing.T) {
	cachedData := &atomic.Value{}
	cachedData.Store([]byte(`{"time":99}`))

	mux := buildMux("topsecret", cachedData)
	req := httptest.NewRequest(http.MethodGet, "/r?secret=topsecret", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Refresh"); got != "5" {
		t.Fatalf("expected Refresh header to be 5, got %q", got)
	}
	if got := rec.Body.String(); got != `{"time":99}` {
		t.Fatalf("unexpected response body: %s", got)
	}
}
