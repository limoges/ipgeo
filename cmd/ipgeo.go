package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	return nil
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
