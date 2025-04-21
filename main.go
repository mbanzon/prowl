package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type wgKeyType string

const (
	wgKey        wgKeyType = "waitGroup"
	secretKeyEnv string    = "PROWL_SECRET"
)

func main() {
	port := flag.Int("port", 5001, "port to run the server on")
	protected := flag.Bool("protect", false, "set to true if the server access should be protected")
	protectionKey := flag.String("secret", "", "should be set (or set through environment) if you use server access protection")
	flag.Parse()

	validatePort(*port)
	secret := validateProtection(*protected, *protectionKey)

	sig, ctx, cancel, wg := setupSync()
	dataChannel := handleData(ctx)
	server := startServer(*port, secret, dataChannel)

	<-sig
	log.Println("Shutting down...")

	cancel()
	wg.Wait() // wait for all goroutines to finish

	// shutdown the server
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		log.Println("error shutting down server:", err)
	}
}

func setupSync() (chan os.Signal, context.Context, context.CancelFunc, *sync.WaitGroup) {
	// setup handler for shutndown using CTRL+C etc.
	sig := make(chan os.Signal, 100)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	// add wait group to wait for all goroutines to finish
	wg := &sync.WaitGroup{}

	// put wait group in context
	ctx = context.WithValue(ctx, wgKeyType("waitGroup"), wg)
	return sig, ctx, cancel, wg
}

func validatePort(port int) {
	if port < 1024 {
		log.Fatal("port number must be greater than 1024:", port)
	}
}

func validateProtection(protected bool, key string) string {
	if !protected {
		return ""
	}

	if key != "" {
		return key
	}

	envSecret := os.Getenv(secretKeyEnv)
	if envSecret != "" {
		return envSecret
	}

	log.Fatal("no secret given with protection enabled")
	return ""
}
