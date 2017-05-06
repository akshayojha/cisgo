package main

import (
	"bufio"
	"communicator"
	"flag"
	"fmt"
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
	log.Println("Listening on %s:%s", serverIP, serverPort)
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
		if contMsg && header == communicator.RegMsg {
			serverMutex.Lock()
			testerInfo := msgCont
			testersList = append(testersList, testerInfo)
			tokens := strings.Split(testerInfo, communicator.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			communicator.SendData(testerIP, testerPort, communicator.OkMsg)
			serverMutex.Unlock()
		} else if contMsg && header == communicator.ResMsg {
			serverMutex.Lock()
			resForCommit := msgCont[0]
			result := msgCont[1:]
			writeResultToFile(resForCommit, result)
			delete(commitTestersMap[hash])
			serverMutex.Unlock()
		} else if contMsg && header == communicator.TestMsg {
			serverMutex.Lock()
			commitAssigned := false
			for i := 0; i < len(testersList); i++ {
				commitToTest := msgCont
				tokens := strings.Split(testersList[i], communicator.Colon)
				testerIP, testerPort := tokens[0], tokens[1]
				resp := communicator.SendAndReceiveData(testerIP, testerPort, communicator.HelloMsg)
				if resp == communicator.HelloMsg {
					commitTestersMap[commitToTest] = testersList[i]
					communicator.SendData(serverIP, serverPort, communicator.OkMsg)
					commitAssigned = true
					break
				}
			}
			if commitAssigned == false {
				commitsToTest = append(commitsToTest, commitToTest)
				communicator.SendData(serverIP, serverPort, communicator.FailMsg)
			}
			serverMutex.Unlock()
		} else {
			log.Println("Unknown Request %s", resp)
		}
	}
	conn.Close()
}

func watchTesters() {
	for {
		for index := 0; index < len(testersList); index++ {
			tokens := strings.Split(testersList[i], communicator.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			resp := communicator.SendAndReceiveData(testerIP, testerPort, communicator.HelloMsg)
			if resp != communicator.HelloMsg {
				log.Println("Removing tester running at %s:%s", testerIP, testerPort)
				// Remove the tester
			}
		}
		time.Sleep(5)
	}
}

func recoverFailedTests() {
	for {

	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information required to setup the server
	serverIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	serverPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	flag.Parse()

	fmt.Println(*serverIPPtr, *serverPortPtr)

	// Start watching the given repository path
	listen(*serverIPPtr, *serverPortPtr)

	// Watch for testers failing
	go watchTesters()

	// Try to assign commits on failed testers to
	// new testers
	go recoverFailedTests()
}
