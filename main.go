package main

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"io"
	"net"
	"net/http"
	"os"
)

// Define EchoHandler Class
type EchoHandler struct{}

// Constructor
func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

// Handler
func (*EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to handle request:", err)
	}
}

// NewServeMux will build a ServeMux that helps us to process the target api
func NewServeMux(echo *EchoHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/echo", echo)
	return mux
}

// NewHTTPServer
// HTTP Server Creator
// The function will generate a http server in the memory so we could directly use it with pointers.
func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux) *http.Server {
	// Create and config the server
	srv := &http.Server{Addr: ":8080", Handler: mux}
	// define the hooks and lifecycles
	lc.Append(fx.Hook{
		// Runs when the server is on
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			// Success notification
			fmt.Println("Starting HTTP server at", srv.Addr)
			go srv.Serve(ln)
			return nil
		},
		// Runs when the server receive system interrupted
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func main() {
	// Create new fx application
	fx.New(
		// Let Fx knows what packages we have to launch the application
		fx.Provide(
			NewHTTPServer,
			NewEchoHandler,
			NewServeMux,
		),
		// Invoke to let the system to run the server
		fx.Invoke(func(server *http.Server) {}),
	).Run()
}
