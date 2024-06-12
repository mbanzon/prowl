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
	wgKey wgKeyType = "waitGroup"
)

func main() {
	port := getServerPortNumber()
	sig, ctx, cancel, wg := setupSync()
	dataChannel := setupDataHandling(ctx)
	server := startServer(port, dataChannel)

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

func getServerPortNumber() int {
	port := -1

	flag.IntVar(&port, "port", 5001, "port to run the server on")
	flag.Parse()
	flag.Parse()

	if port < 1024 {
		log.Fatal("port number must be greater than 1024:", port)
	}
	return port
}

func setupDataHandling(ctx context.Context) chan output {
	// start the CPU reporting
	cpuChannel := startCpuUsageReporting(ctx)
	// start the load average reporting
	loadChannel := startLoadAverageReporting(ctx)

	// handle the data
	dataChannel := handleData(cpuChannel, loadChannel, ctx)

	return dataChannel
}
