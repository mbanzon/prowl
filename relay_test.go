package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestBuildRelayURL(t *testing.T) {
	endpoint, err := buildRelayURL("http://example.com", "host-a", "s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "http://example.com/relay?machine_name=host-a&secret=s3cr3t" {
		t.Fatalf("unexpected relay url: %s", endpoint)
	}

	endpoint, err = buildRelayURL("http://example.com/api", "host-b", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "http://example.com/api/relay?machine_name=host-b" {
		t.Fatalf("unexpected relay url: %s", endpoint)
	}
}

func TestStartRelayReporterPostsData(t *testing.T) {
	var receivedBody []byte
	var receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		receivedBody = body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, wgKey, wg)

	ch := make(chan output, 1)
	startRelayReporter(ctx, server.URL, "relaysecret", "machine-1", ch)

	ch <- output{Time: 42}
	close(ch)

	waitForWaitGroup(t, wg)
	cancel()

	if receivedQuery != "machine_name=machine-1&secret=relaysecret" {
		t.Fatalf("unexpected query: %s", receivedQuery)
	}

	var payload output
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.Time != 42 {
		t.Fatalf("expected time 42, got %d", payload.Time)
	}
}
