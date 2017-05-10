package main

import (
	"communicator"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// Function to get the latest commit hash of a given
// local repository folder
func getLastCommitHash(repoFolder string) string {
	out, err := exec.Command(communicator.GitHashCmd, arg)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func watch(serverIP, serverPort, repoFolder string) {
	// Watch every interval seconds for any new commit by the developer
	for {
		// Get the most recent commit hash
		lastCommitHash := getLastCommitHash(repoFolder)

		// Fetch the latest commit for the repo
		out, err := exec.Command(communicator.GitPullCmd, arg)

		if err != nil {
			log.Fatalf("Cannot fetch new changes - %s", err)
		}

		// Get the latest commit hash
		latestCommitHash := getLastCommitHash(repoFolder)

		fmt.Println(lastCommitHash, latestCommitHash)

		if lastCommitHash != latestCommitHash {
			resp := communicator.SendAndReceiveData(serverIP, serverPort, communicator.StatMsg)
			if resp == communicator.OkMsg {
				resp := communicator.SendAndReceiveData(serverIP, serverPort, latestCommitHash+communicator.MsgDel)
				if resp == communicator.OkMsg {
					log.Printf("Scheduled tests for %s \n", latestCommitHash)
				} else {
					log.Printf("Unable to schedule test for %s \n", latestCommitHash)
				}
			} else {
				log.Printf("Cannot communicate with server on %s:%s \n", serverIP, serverPort)
			}
		} else {
			log.Println("No new commit found")
		}
		time.Sleep(communicator.WaitInterval)
	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information
	schedServerIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	schedServerPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	// Local repository path to observe
	repoPathPtr := flag.String("rpath", communicator.EmptyStr, "Path to the repository folder to observe")

	flag.Parse()

	// Validate the local repository path
	if *repoPathPtr == communicator.EmptyStr {
		log.Fatal("Path to local repository folder required")
	}

	// Navigate to the local repository path
	if _, err := os.Chdir(*repoPathPtr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Watching %s at %s:%s \n", *repoPathPtr, *schedServerIPPtr, *schedServerPortPtr)

	// Start watching the given repository path
	watch(*schedServerIPPtr, *schedServerPortPtr, *repoPathPtr)
}
