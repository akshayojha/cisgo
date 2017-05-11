package main

import (
	"bufio"
	"cisgo/util"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var serverAlive = make(chan bool)
var testerBusy = make(chan bool)

func watchScheduler(serverIP, serverPort string) {
	for {
		resp := util.SendAndReceiveData(serverIP, serverPort, util.StatMsg)
		if resp != util.OkMsg {
			log.Fatalf("Scheduler at %s:%s is no longer active\n", serverIP, serverPort)
			serverAlive <- false
		}
		time.Sleep(util.WaitInterval)
	}
}

func registerTester(testerIP, testerPort, serverIP, serverPort string) {

	log.Printf("Registering tester daemon to %s:%s \n", serverIP, serverPort)

	regInfo := util.RegMsg + util.Dash + testerIP + util.Colon + testerPort

	resp := util.SendAndReceiveData(serverIP, serverPort, regInfo)

	if resp != util.OkMsg {
		log.Fatalf("Cannot register tester to %s:%s\n", serverIP, serverPort)
	} else {
		log.Printf("Registered tester to %s:%s\n", serverIP, serverPort)
	}
}

func listen(testerIP, testerPort, repo, serverIP, serverPort string) {

	server := testerIP + util.Colon + testerPort
	listner, err := net.Listen(util.Protocol, server)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s:%s \n", testerIP, testerPort)

	for {
		schedStatus := <-serverAlive
		if schedStatus == false {
			log.Fatalf("Shutting down the tester daemon at %s:%s", testerIP, testerPort)
		}
		conn, err := listner.Accept()
		if err != nil {
			listner.Close()
			log.Fatal(err)
		}
		go handleRequest(conn, serverIP, serverPort, repo)
	}
}

func handleRequest(conn net.Conn, serverIP, serverPort, repo string) {
	resp, err := bufio.NewReader(conn).ReadBytes(util.MsgDelByte)

	if err != nil {
		log.Println(err)
	}

	formattedResp := util.FormatResp(resp)

	// Tokenize the protocol
	msg := strings.Split(formattedResp, util.Dash)
	statMsg := len(msg) == 1
	contMsg := len(msg) > 1
	header := msg[0]

	if statMsg && header == util.StatMsg {
		_, err := conn.Write([]byte(util.OkMsg + util.MsgDel))
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Told the server at %s:%s that I am alive/n", serverIP, serverPort)
		}
	} else {
		msgCont := msg[1]
		if contMsg && header == util.TestMsg {
			busy := <-testerBusy
			if busy == true {
				_, err := conn.Write([]byte(util.FailMsg + util.MsgDel))
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Told the server at %s:%s that I am busy/n", serverIP, serverPort)
				}
			} else {
				commitToTest := msgCont
				// Start running test for the
				testerBusy <- true
				go runTest(commitToTest, serverIP, serverPort, repo)
				_, err := conn.Write([]byte(util.FailMsg + util.MsgDel))
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Told the server at %s:%s that I have started running test for %s/n", serverIP, serverPort, commitToTest)
				}
			}
		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
	}
	conn.Close()
}

func runTest(commit, serverIP, serverPort, repo string) {
	// Clean the repo
	util.RunOrFail(util.GitExecutable, util.GitCleanSwitch)

	// Fetch the latest commit for the repo
	util.RunOrFail(util.GitExecutable, util.GitPullSwitch)

	// Set repository to given commit
	resetToCommitSwitch := append(util.GitResetToCommitSwitch, commit)
	util.RunOrFail(util.GitExecutable, resetToCommitSwitch)

	// Now run the actual tests
	testOutput := util.RunOrFail(util.GoExecutable, util.GoTestSwitch)
	completeResult := util.ResMsg + util.Dash + testOutput
	util.SendData(serverIP, serverPort, completeResult)
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

	log.Printf("Starting tester daemon at %s:%s watching %s \n", *testerIPPtr, *testerPortPtr, *repoPathPtr)

	registerTester(*testerIPPtr, *testerPortPtr, *serverIPPtr, *serverPortPtr)

	go watchScheduler(*serverIPPtr, *serverPortPtr)

	listen(*testerIPPtr, *testerPortPtr, *repoPathPtr, *serverIPPtr, *serverPortPtr)
}
