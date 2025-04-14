package gracefulshutdown

import "context"

type Server interface {
	Name() string

	// Start starts the server and listens for incoming requests.
	Start(ctx context.Context) error

	// Shutdown gracefully shuts down the server, allowing any ongoing requests to complete.
	Shutdown() error
}
