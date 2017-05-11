package main

import (
	"cisgo/util"
	"flag"
	"log"
	"os"
	"time"
)

func watch(serverIP, serverPort, repoFolder string) {
	// Watch every interval seconds for any new commit by the developer
	for {
		// Get the most recent commit hash
		lastCommitHash := util.RunOrFail(util.GitExecutable, util.GitHashSwitch)

		// Fetch the latest commit for the repo
		util.RunOrFail(util.GitExecutable, util.GitPullSwitch)

		// Get the latest commit hash
		latestCommitHash := util.RunOrFail(util.GitExecutable, util.GitHashSwitch)

		if lastCommitHash != latestCommitHash {
			resp := util.SendAndReceiveData(serverIP, serverPort, util.StatMsg)
			if resp == util.OkMsg {
				resp := util.SendAndReceiveData(serverIP, serverPort, latestCommitHash)
				if resp == util.OkMsg {
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
		time.Sleep(util.WaitInterval)
	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information
	schedServerIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	schedServerPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	// Local repository path to observe
	repoPathPtr := flag.String("rpath", util.EmptyStr, "Path to the repository folder to observe")

	flag.Parse()

	// Validate the local repository path
	if *repoPathPtr == util.EmptyStr {
		log.Fatal("Path to local repository folder required")
	}

	// Navigate to the local repository path
	if err := os.Chdir(*repoPathPtr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Watching %s at %s:%s \n", *repoPathPtr, *schedServerIPPtr, *schedServerPortPtr)

	// Start watching the given repository path
	watch(*schedServerIPPtr, *schedServerPortPtr, *repoPathPtr)
}
