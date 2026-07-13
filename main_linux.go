//go:build !windows

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	srv, _ := Run()

	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		Serve(srv, nil, stop)
		close(done)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh

	log.Printf("Received %v, shutting down...", sig)
	close(stop)
	<-done
}
