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
	servers map[string]Server

	signal  chan os.Signal
	errChan chan error
}

// New creates a new GracefulShutdown instance.
func New() *GracefulShutdown {
	return &GracefulShutdown{
		servers: make(map[string]Server),
		signal:  make(chan os.Signal, 1),
		errChan: make(chan error, 1),
	}
}

func (g *GracefulShutdown) Add(name string, server Server) *GracefulShutdown {
	g.servers[name] = server
	return g
}

func (g *GracefulShutdown) CatchSignals() {
	signal.Notify(g.signal, syscall.SIGINT, syscall.SIGTERM)
}

func (g *GracefulShutdown) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for name, server := range g.servers {
		go func(name string, srv Server) {
			log.Printf("Starting server: %s", name)
			if err := srv.Start(ctx); err != nil {
				log.Printf("Error starting server: %v", err)
				g.errChan <- err
			}
		}(name, server)
	}

	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal, shutting down servers...")
	case <-g.errChan:
		log.Println("Server start error, shutting down servers...")
	case sig := <-g.signal:
		log.Printf("Received signal: %s, shutting down servers...", sig)
	}

	g.shutdown()
	return nil
}

func (g *GracefulShutdown) shutdown() {
	var wg sync.WaitGroup
	for name, server := range g.servers {
		wg.Add(1)
		go func(name string, srv Server) {
			log.Printf("Shutting down server: %s", name)
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down server: %v", err)
			}
			wg.Done()
		}(name, server)
	}
	wg.Wait()
	log.Println("All servers shut down gracefully.")
}
