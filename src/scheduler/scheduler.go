package main

import (
	"bufio"
	"communicator"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var serverMutex = sync.Mutex
var commitTestersMap = make(map[string]string)
var testersList = []string{}
var commitsToTest = []string{}

func listen(serverIP, serverPort string) {
	server := serverIP + communicator.Colon + serverPort
	listner, err := net.Listen(communicator.Protocol, server)
	if err != nil {
		log.Fatal(err)
	}
	listner.Close()
	log.Printf("Listening on %s:%s \n", serverIP, serverPort)
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn)
	}
}

func writeResultToFile(hash, result string) {
	f, err := os.Create(hash)
	if err != nil {
		log.Println(err)
	} else {
		_, err := f.WriteString(result)
		if err != nil {
			log.Println(err)
		} else {
			f.Close()
		}
	}
}

func handleRequest(conn net.Conn) {
	resp, err := bufio.NewReader(conn).ReadString(communicator.MsgDel)

	if err != nil {
		log.Println(err)
	}

	// Tokenize the protocol
	msg := strings.Split(resp, communicator.Dash)

	statMsg := len(msg) == 1
	contMsg := len(msg) > 1
	header := msg[:1]
	msgCont := msg[1:]

	if statMsg && header == communicator.StatMsg {
		communicator.SendData(serverIP, serverPort, communicator.OkMsg)
	} else {
		serverMutex.Lock()
		if contMsg && header == communicator.RegMsg {
			testerInfo := msgCont
			testersList = append(testersList, testerInfo)
			tokens := strings.Split(testerInfo, communicator.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			communicator.SendData(testerIP, testerPort, communicator.OkMsg)
		} else if contMsg && header == communicator.ResMsg {
			resForCommit := msgCont[0]
			result := msgCont[1:]
			writeResultToFile(resForCommit, result)
			delete(commitTestersMap[resForCommit])
		} else if contMsg && header == communicator.TestMsg {
			commitToTest := msgCont
			assignTester(commitToTest)
		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
		serverMutex.Unlock()
	}
	conn.Close()
}

func assignTester(commitToTest string) {
	commitAssigned := false
	for _, tester := range testersList {
		tokens := strings.Split(tester, communicator.Colon)
		testerIP, testerPort := tokens[0], tokens[1]
		resp := communicator.SendAndReceiveData(testerIP, testerPort, communicator.HelloMsg)
		if resp == communicator.HelloMsg {
			commitTestersMap[commitToTest] = tester
			communicator.SendData(serverIP, serverPort, communicator.OkMsg)
			commitAssigned = true
			break
		}
	}
	if commitAssigned == false {
		commitsToTest = append(commitsToTest, commitToTest)
		communicator.SendData(serverIP, serverPort, communicator.FailMsg)
	}
}

func watchTesters() {
	for {
		for index := 0; index < len(testersList); index++ {
			tokens := strings.Split(testersList[i], communicator.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			resp := communicator.SendAndReceiveData(testerIP, testerPort, communicator.HelloMsg)
			if resp != communicator.HelloMsg {
				serverMutex.Lock()
				log.Printf("Removing tester running at %s:%s\n", testerIP, testerPort)
				// Remove the tester
				failedCommit := getMapKeyFromValue(testersList[i])
				commitsToTest = append(commitsToTest, failedCommit)
				testersList = append(testersList[:index], testersList[index+1:])
				serverMutex.Unlock()
			}
		}
		time.Sleep(5)
	}
}

func getMapKeyFromValue(value string) {
	for k, v := range commitTestersMap {
		if v == value {
			return k
		}
	}
	log.Fatalf("Can't find value %s in %s\n", value, commitTestersMap)
}

func recoverFailedTests() {
	for {
		for _, recoverCommit := range commitsToTest {
			serverMutex.Lock()
			assignTester(recoverCommit)
			serverMutex.Unlock()
		}
		time.Sleep(5)
	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information required to setup the server
	serverIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	serverPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	flag.Parse()

	log.Printf("Starting scheduler server at %s:%s\n", *serverIPPtr, *serverPortPtr)

	// Start watching the given repository path
	go listen(*serverIPPtr, *serverPortPtr)

	// Watch for testers failing
	go watchTesters()

	// Try to assign commits on failed testers to
	// new testers
	go recoverFailedTests()
}
