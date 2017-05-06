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

// Version control system specific commands

// GitHashCmd - command to get last git commit hash
const GitHashCmd = "git rev-parse HEAD"

// GitPullCmd - command to fetch and apply latest git commit
const GitPullCmd = "git pull"

// EmptyStr - string to denote empty string
const EmptyStr = ""

// Function to get the latest commit hash of a given
// local repository folder
func getLastCommitHash(repoFolder string) string {
	out, err := exec.Command(GitHashCmd, arg)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func watch(serverIP, serverPort, repoFolder string) {
	// Watch every 5 seconds for any new commit by the developer
	for {
		// Get the most recent commit hash
		lastCommitHash := getLastCommitHash(repoFolder)

		// Fetch the latest commit for the repo
		out, err := exec.Command(GitPullCmd, arg)

		// Get the latest commit hash
		latestCommitHash := getLastCommitHash(repoFolder)

		fmt.Println(lastCommitHash, latestCommitHash)

		if lastCommitHash != latestCommitHash {
			resp := communicator.SendAndReceiveData(serverIP, serverPort, communicator.StatMsg)
			if resp == communicator.OkMsg {
				resp := communicator.SendAndReceiveData(serverIP, serverPort, latestCommitHash+communicator.MsgDel)
				if resp == communicator.OkMsg {
					log.Println("Scheduled tests for %s", latestCommitHash)
				} else {
					log.Println("Unable to schedule test for %s", latestCommitHash)
				}
			} else {
				log.Println("Cannot communicate with server on %s:%s", serverIP, serverPort)
			}
		} else {
			log.Println("No new commit found")
		}
		time.Sleep(5)
	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information
	schedServerIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	schedServerPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

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

	fmt.Println(*schedServerIPPtr, *schedServerPortPtr, *repoPathPtr)

	// Start watching the given repository path
	watch(*schedServerIPPtr, *schedServerPortPtr, *repoPathPtr)
}
