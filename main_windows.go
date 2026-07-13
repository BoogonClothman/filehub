//go:build windows

package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sys/windows/svc"
)

type serviceHandler struct {
	srv *http.Server
}

func (h *serviceHandler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const accepts = svc.AcceptStop | svc.AcceptShutdown

	s <- svc.Status{State: svc.StartPending}

	stop := make(chan struct{})
	ready := make(chan struct{})
	done := make(chan struct{})

	go func() {
		Serve(h.srv, ready, stop)
		close(done)
	}()
	<-ready

	s <- svc.Status{State: svc.Running, Accepts: accepts}

	for c := range r {
		switch c.Cmd {
		case svc.Stop, svc.Shutdown:
			log.Println("Service stop requested by SCM")
			s <- svc.Status{State: svc.StopPending}
			close(stop)
			<-done
			return false, 0

		case svc.Interrogate:
			s <- c.CurrentStatus
		}
	}

	close(stop)
	<-done
	return false, 0
}

func main() {
	srv, _ := Run()

	isSvc, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to check service status: %v", err)
	}

	if isSvc {
		svc.Run("FileHub", &serviceHandler{srv: srv})
		return
	}

	// Interactive mode
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		Serve(srv, nil, stop)
		close(done)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	log.Println("Ctrl+C received, shutting down...")
	close(stop)
	<-done
}
