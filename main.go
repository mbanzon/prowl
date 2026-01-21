package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type wgKeyType string

const (
	wgKey           wgKeyType = "waitGroup"
	secretKeyEnv    string    = "PROWL_SECRET"
	shutdownTimeout           = 10 * time.Second
)

func main() {
	port := flag.Int("port", 5001, "port to run the server on")
	protected := flag.Bool("protect", false, "set to true if the server access should be protected")
	protectionKey := flag.String("secret", "", "should be set (or set through environment) if you use server access protection")
	relayMode := flag.Bool("relay", false, "enable relay mode to accept reports from other machines")
	relayLocal := flag.Bool("relay-local", true, "set to false to disable local reporting while in relay mode")
	relayTarget := flag.String("relay-target", "", "relay server base URL to post stats to")
	machineName := flag.String("machine-name", "", "machine name to use when reporting to a relay server")
	relaySecret := flag.String("relay-secret", "", "relay server secret key (or set PROWL_RELAY_SECRET)")
	flag.Parse()

	shouldServe := *relayMode || *relayTarget == ""
	shouldCollectLocal := *relayTarget != "" || !*relayMode || *relayLocal

	if shouldServe {
		validatePort(*port)
	}

	secret := ""
	if shouldServe {
		secret = validateProtection(*protected, *protectionKey)
	}

	if *relayTarget != "" && *machineName == "" {
		log.Fatal("machine-name must be set when using relay-target")
	}

	relaySecretValue := *relaySecret
	if relaySecretValue == "" {
		relaySecretValue = os.Getenv("PROWL_RELAY_SECRET")
	}

	sig, ctx, cancel, wg := setupSync()
	var server *http.Server
	var dataChannel chan output

	if shouldCollectLocal {
		dataChannel = handleData(ctx)
	}

	if shouldServe {
		serverChannel := dataChannel
		if *relayMode && !*relayLocal && *relayTarget != "" {
			serverChannel = nil
		}
		if *relayMode && *relayTarget != "" && *relayLocal {
			serverChannel, dataChannel = teeOutput(ctx, dataChannel)
		}
		server = startServer(*port, secret, *relayMode, *relayLocal, serverChannel)
	}

	if *relayTarget != "" {
		startRelayReporter(ctx, *relayTarget, relaySecretValue, *machineName, dataChannel)
	}

	<-sig
	log.Println("Shutting down...")

	cancel()
	wg.Wait() // wait for all goroutines to finish

	// shutdown the server
	if server != nil {
		shutdownCtx, shutdownCancel := shutdownContext()
		defer shutdownCancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Println("error shutting down server:", err)
		}
	}
}

func shutdownContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), shutdownTimeout)
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
	validatePortWith(port, log.Fatal)
}

func validatePortWith(port int, fatalf func(...any)) {
	if port < 1024 {
		fatalf("port number must be greater than 1024:", port)
	}
}

func validateProtection(protected bool, key string) string {
	return validateProtectionWith(protected, key, os.Getenv, log.Fatal)
}

func validateProtectionWith(protected bool, key string, getenv func(string) string, fatalf func(...any)) string {
	if !protected {
		return ""
	}

	if key != "" {
		return key
	}

	envSecret := getenv(secretKeyEnv)
	if envSecret != "" {
		return envSecret
	}

	fatalf("no secret given with protection enabled")
	return ""
}
