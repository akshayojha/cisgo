package main

import (
	"flag"
	"log"
	"os"
)

// EmptyStr - string to denote empty string
var EmptyStr = ""

func main() {
	// Get required information in command line from user

	// Scheduler information
	testerIPPtr := flag.String("tip", "localhost", "IP address of the tester")
	testerPortPtr := flag.String("tport", "0", "Port of the tester")

	// Scheduler information required to setup the server
	serverIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	serverPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	// Local repository path to observe
	repoPathPtr := flag.String("rpath", EmptyStr, "Path to the repository folder to observe")

	flag.Parse()

	// Validate the local repository path
	if *repoPathPtr == EmptyStr {
		log.Fatal("Path to local repository folder required")
	}

	// Navigate to the local repository path
	if _, err := os.Chdir(*repoPathPtr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting tester daemon at %s:%s watching %s \n", *testerIPPtr, *testerPortPtr, *repoPathPtr)
}
