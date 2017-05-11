package main

import (
	"bufio"
	"cisgo/src/communicator"
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var serverAlive = make(chan bool)
var testerBusy = make(chan bool)

func watchScheduler(serverIP, serverPort string) {
	for {
		resp := communicator.SendAndReceiveData(serverIP, serverPort, communicator.StatMsg)
		if resp != communicator.OkMsg {
			log.Fatalf("Scheduler at %s:%s is no longer active\n", serverIP, serverPort)
			serverAlive <- false
		}
		time.Sleep(communicator.WaitInterval)
	}
}

func listen(testerIP, testerPort, repo, serverIP, serverPort string) {
	server := testerIP + communicator.Colon + testerPort
	listner, err := net.Listen(communicator.Protocol, server)
	if err != nil {
		log.Fatal(err)
	}
	listner.Close()
	log.Printf("Listening on %s:%s \n", testerIP, testerPort)
	for {
		schedStatus := <-serverAlive
		if schedStatus == false {
			log.Fatalf("Shutting down the tester daemon at %s:%s", testerIP, testerPort)
		}
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn, serverIP, serverPort, repo)
	}
}

func handleRequest(conn net.Conn, serverIP, serverPort, repo string) {
	resp, err := bufio.NewReader(conn).ReadString(communicator.MsgDelByte)

	if err != nil {
		log.Println(err)
	}

	// Tokenize the protocol
	msg := strings.Split(resp, communicator.Dash)

	statMsg := len(msg) == 1
	contMsg := len(msg) > 1
	header := msg[0]
	msgCont := msg[1]

	if statMsg && header == communicator.StatMsg {
		communicator.SendData(serverIP, serverPort, communicator.OkMsg)
	} else {
		if contMsg && header == communicator.TestMsg {
			busy := <-testerBusy
			if busy == true {
				communicator.SendData(serverIP, serverPort, communicator.FailMsg)
			} else {
				commitToTest := msgCont
				// Start running test for the
				testerBusy <- true
				go runTest(commitToTest, serverIP, serverPort, repo)
				communicator.SendData(serverIP, serverPort, communicator.OkMsg)
			}
		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
	}
	conn.Close()
}

func runOrFail(cmd string, arg string) string {
	out, err := exec.Command(cmd, arg)

	if err != nil {
		log.Fatal(err)
	}

	return out
}

func runTest(commit, serverIP, serverPort, repo string) {
	// Clean the repo
	runOrFail(communicator.GitCleanCmd)

	// Fetch the latest commit for the repo
	runOrFail(communicator.GitPullCmd)

	// Set repository to given commit
	resetCmd := communicator.GitResetToCommitCmd + " " + commit
	runOrFail(resetCmd)

	// Now run the actual tests
	testOutput := runOrFail(communicator.GoTestCmd)
	completeResult := communicator.ResMsg + communicator.Dash + testOutput
	communicator.SendData(serverIP, serverPort, completeResult)
	testerBusy <- false
}

func main() {
	// Get required information in command line from user

	// Scheduler information
	testerIPPtr := flag.String("tip", "localhost", "IP address of the tester")
	testerPortPtr := flag.String("tport", "0", "Port of the tester")

	// Scheduler information required to setup the server
	serverIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	serverPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	// Local repository path to observe
	repoPathPtr := flag.String("rpath", communicator.EmptyStr, "Path to the repository folder to observe")

	flag.Parse()

	// Validate the local repository path
	if *repoPathPtr == communicator.EmptyStr {
		log.Fatal("Path to local repository folder required")
	}

	// Navigate to the local repository path
	if err := os.Chdir(*repoPathPtr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting tester daemon at %s:%s watching %s \n", *testerIPPtr, *testerPortPtr, *repoPathPtr)
	log.Printf("Registering tester daemon to %s:%s \n", *serverIPPtr, *serverPortPtr)

	// Register tester to the scheduler server
	regInfo := communicator.RegMsg + communicator.Dash + *testerIPPtr + communicator.Colon + *testerPortPtr

	resp := communicator.SendAndReceiveData(*serverIPPtr, *serverPortPtr, regInfo)

	if resp != communicator.OkMsg {
		log.Fatalf("Cannot register tester to %s:%s\n", *serverIPPtr, *serverPortPtr)
	}

	go watchScheduler(*serverIPPtr, *serverPortPtr)

	go listen(*testerIPPtr, *testerPortPtr, *repoPathPtr, *serverIPPtr, *serverPortPtr)
}
