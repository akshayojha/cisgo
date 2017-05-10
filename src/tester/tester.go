package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	// Get required information in command line from user

	// Scheduler information
	testerIPPtr := flag.String("tip", "localhost", "IP address of the tester")
	testerPortPtr := flag.String("tport", "8080", "Port of the tester")

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
