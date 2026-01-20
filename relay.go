package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func startRelayReporter(ctx context.Context, relayTarget string, relaySecret string, machineName string, in <-chan output) {
	startRelayReporterWith(ctx, relayTarget, relaySecret, machineName, in, http.DefaultClient.Do)
}

func startRelayReporterWith(ctx context.Context, relayTarget string, relaySecret string, machineName string, in <-chan output, doRequest func(*http.Request) (*http.Response, error)) {
	if in == nil {
		log.Println("Relay reporting disabled: no input channel")
		return
	}

	wg := ctx.Value(wgKey).(*sync.WaitGroup)
	wg.Add(1)

	go func() {
		log.Println("Relay reporting started")
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Println("Relay reporting stopped")
				return
			case data, ok := <-in:
				if !ok {
					log.Println("Relay reporting input closed")
					return
				}

				payload, err := json.Marshal(data)
				if err != nil {
					log.Println("error marshalling relay data:", err)
					continue
				}

				endpoint, err := buildRelayURL(relayTarget, machineName, relaySecret)
				if err != nil {
					log.Println("error building relay url:", err)
					continue
				}

				req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
				if err != nil {
					log.Println("error creating relay request:", err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := doRequest(req)
				if err != nil {
					log.Println("error posting relay data:", err)
					continue
				}
				resp.Body.Close()
			}
		}
	}()
}

func buildRelayURL(relayTarget string, machineName string, relaySecret string) (string, error) {
	parsed, err := url.Parse(relayTarget)
	if err != nil {
		return "", err
	}

	path := parsed.Path
	if path == "" || path == "/" {
		path = "/relay"
	} else if !strings.HasSuffix(path, "/relay") {
		path = strings.TrimRight(path, "/") + "/relay"
	}
	parsed.Path = path

	query := parsed.Query()
	query.Set("machine_name", machineName)
	if relaySecret != "" {
		query.Set(secretKey, relaySecret)
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
