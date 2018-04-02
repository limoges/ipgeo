package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
)

// main launches the program and returns an appropriate error.
func mainWithError() error {
	var (
		locsFilepath string
		netsFilepath string
		port         string
		verbose      bool
	)
	flag.StringVar(&locsFilepath, "locations", "locations.csv",
		"path to csv file mapping IDs to locations")
	flag.StringVar(&netsFilepath, "networks", "ipv4.csv",
		"path to csv file mapping CIDR-formatted networks to IDs")
	flag.StringVar(&port, "port", "8080", "port to bind the http server to")
	flag.BoolVar(&verbose, "verbose", false, "use verbose output")
	flag.Parse()

	if verbose {
		log.Printf(`options: --locations=%s --networks=%s --port=%s --verbose=%t`,
			locsFilepath, netsFilepath, port, verbose)
	}

	// HTTP handler to respond to geolocation requests.
	locate := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ip/", locate)

	srv := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
	}

	var rg run.Group
	{
		// Start & stop for http.Server
		rg.Add(func() error {
			log.Printf("listening on %s", srv.Addr)
			return srv.ListenAndServe()
		}, func(error) {
			// Wait 10 seconds before shutting down the server so that
			// open connexions can be naturally closed withtout errors.
			grace := 10 * time.Second
			ctx, _ := context.WithDeadline(context.Background(),
				time.Now().Add(grace))
			if err := srv.Shutdown(ctx); err != nil {
				log.Println(err)
			}
		})
	}
	{
		// Start & stop for signal listener
		// cancel is a synchronization channel. close(cancel) signals
		// that the listener should abort.
		cancel := make(chan struct{})
		rg.Add(func() error {
			return interruptListener(cancel)
		}, func(error) {
			close(cancel)
		})
	}
	return rg.Run()
}

// interruptListener waits for an interrupt signal (^C) and returns an error when
// caught. When the cancel channel is closed, interruptListener returns.
func interruptListener(cancel <-chan struct{}) error {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, syscall.SIGINT)
	select {
	case <-interrupt:
		return errors.New("caught interrupt")
	case <-cancel:
		// Abort gracefully.
		return nil
	}
}

// main handles program failure/success, printing error and exiting properly.
func main() {
	if err := mainWithError(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
