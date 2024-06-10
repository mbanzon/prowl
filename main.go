package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := -1

	flag.IntVar(&port, "port", 5001, "port to run the server on")
	flag.Parse()

	// check that port number is > 1024
	if port < 1024 {
		log.Fatal("port number must be greater than 1024:", port)
	}

	// setup handler for shutndown using CTRL+C etc.
	sig := make(chan os.Signal, 100)
	stopChan := make(chan bool)

	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	// start the CPU reporting
	cpuChannel := startCpuUsageReporting(stopChan)

	// start the load average reporting
	loadChannel := startLoadAverageReporting(stopChan)

	// handle the data
	dataChannel := handleData(cpuChannel, loadChannel, stopChan)

	// start the server
	server := startServer(port, dataChannel)

	go func() {
		<-sig
		close(stopChan)
		server.Shutdown(context.Background())
	}()

	<-stopChan
	log.Println("Shutting down...")
}
