package gracefulshutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type GracefulShutdown struct {
	servers []Server

	signal  chan os.Signal
	errChan chan error
}

// New creates a new GracefulShutdown instance.
func NewGracefulShutdown() *GracefulShutdown {
	return &GracefulShutdown{
		servers: []Server{},
		signal:  make(chan os.Signal, 1),
		errChan: make(chan error, 1),
	}
}

func (g *GracefulShutdown) Add(name string, server Server) {
	g.servers = append(g.servers, server)
}

func (g *GracefulShutdown) CatchSignals() {
	signal.Notify(g.signal, syscall.SIGINT, syscall.SIGTERM)
}

func (g *GracefulShutdown) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, server := range g.servers {
		go func(srv Server) {
			log.Printf("Starting server: %s", srv.Name())
			if err := srv.Start(ctx); err != nil {
				log.Printf("Error starting server: %v", err)
				g.errChan <- err
			}
		}(server)
	}

	defer g.shutdown()
	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal, shutting down servers...")
	case e := <-g.errChan:
		log.Println("Server start error, shutting down servers...")
		return e
	case sig := <-g.signal:
		log.Printf("Received signal: %s, shutting down servers...", sig)
	}

	return nil
}

func (g *GracefulShutdown) shutdown() {
	var wg sync.WaitGroup
	for _, server := range g.servers {
		wg.Add(1)
		go func(srv Server) {
			log.Printf("Shutting down server: %s", srv.Name())
			if err := srv.Shutdown(); err != nil {
				log.Printf("Error shutting down server: %v", err)
			}
			wg.Done()
		}(server)
	}
	wg.Wait()
	log.Println("All servers shut down gracefully.")
}
