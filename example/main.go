package main

import (
	"context"
	"errors"
	gracefulshutdown "github/benlocal/graceful-shutdown"
	"net/http"
)

func main() {
	ctx := context.Background()

	// Create a new GracefulShutdown instance
	g := gracefulshutdown.NewGracefulShutdown()
	g.Add("Server1", &Server1{})
	g.Add("Server2", &Server2{})
	g.CatchSignals()

	// Start the servers
	err := g.Start(ctx)
	if err != nil {
		println("Error starting servers:", err)
	}
}

type Server1 struct{}

func (s *Server1) Start(ctx context.Context) error {
	println("Server1 started")
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from Server1"))
	})

	err := http.ListenAndServe(":6080", mux)
	if err != nil {
		println("Error starting Server1:", err)
		return err
	}

	return nil
}

func (s *Server1) Shutdown() error {
	println("Server1 shutting down")
	return nil
}

type Server2 struct{}

func (s *Server2) Start(ctx context.Context) error {
	println("Server2 started")
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	<-ctx.Done()
	return errors.New("Server2 error")
}

func (s *Server2) Shutdown() error {
	println("Server2 shutting down")
	return nil
}
