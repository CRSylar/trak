package main

import (
	"fmt"
	"os"

	"github.com/CRSylar/trak/internal/daemon"
)

func main() {
	srv, err := daemon.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "trakd: failed to start: %v\n", err)
		os.Exit(1)
	}

	// Serve blocks until the listener is closed (i.e. after 'trak end')
	srv.Serve()
}
