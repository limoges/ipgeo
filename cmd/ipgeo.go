package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/limoges/ipgeo"
	"github.com/oklog/run"
)

// A LocateRequest represents a request to geolocate a given address.
type LocateRequest struct {
	Addr net.IP // An IPv4 address
}

// DecodeLocateRequest inspects an http.Request and extracts the necessary elements
// into a valid LocateRequest.
func DecodeLocateRequest(req *http.Request) (*LocateRequest, error) {
	ip := filepath.Base(req.URL.Path)
	addr := net.ParseIP(ip)
	if addr == nil {
		return nil, fmt.Errorf("invalid request: %s", ip)
	}
	addr = addr.To4()
	if addr == nil {
		return nil, fmt.Errorf("unsupported ipv6: %s", ip)
	}
	return &LocateRequest{Addr: addr}, nil
}

// A LocateResponse represents a response to a LocateRequest.
type LocateResponse struct {
	Country      string `json:"country"`
	Subdivision1 string `json:"subdivision1"`
	Subdivision2 string `json:"subdivision2"`
	City         string `json:"city"`
}

// An IPGeolocator is a
type IPGeolocator interface {
	LocateIP(addr net.IP) (ipgeo.Location, error)
}

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

	var locator IPGeolocator

	// HTTP handler to respond to geolocation requests.
	locate := func(w http.ResponseWriter, r *http.Request) {
		request, err := DecodeLocateRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err)
			return
		}

		location, err := locator.LocateIP(request.Addr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}
		response := LocateResponse{
			Country:      location.Country,
			Subdivision1: location.Subdivision1,
			Subdivision2: location.Subdivision2,
			City:         location.City,
		}
		buf, err := json.MarshalIndent(response, "", "\t")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
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
