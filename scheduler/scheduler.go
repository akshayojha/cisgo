package main

import (
	"bufio"
	"cisgo/util"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var serverMutex = &sync.Mutex{}
var commitTestersMap = make(map[string]string)
var testersList = []string{}
var commitsToTest = []string{}

func listen(serverIP, serverPort string) {
	server := serverIP + util.Colon + serverPort
	listner, err := net.Listen(util.Protocol, server)
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
		go handleRequest(conn, serverIP, serverPort)
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

func handleRequest(conn net.Conn, serverIP, serverPort string) {
	resp, err := bufio.NewReader(conn).ReadString(util.MsgDelByte)

	if err != nil {
		log.Println(err)
	}

	// Tokenize the protocol
	msg := strings.Split(resp, util.Dash)

	statMsg := len(msg) == 1
	contMsg := len(msg) > 1
	header := msg[0]
	msgCont := msg[1]

	if statMsg && header == util.StatMsg {
		util.SendData(serverIP, serverPort, util.OkMsg)
	} else {
		serverMutex.Lock()
		if contMsg && header == util.RegMsg {
			testerInfo := msgCont
			testersList = append(testersList, testerInfo)
			tokens := strings.Split(testerInfo, util.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			util.SendData(testerIP, testerPort, util.OkMsg)
		} else if contMsg && header == util.ResMsg {
			resForCommit := msg[1]
			result := msg[2]
			writeResultToFile(resForCommit, result)
			delete(commitTestersMap, resForCommit)
		} else if contMsg && header == util.TestMsg {
			commitToTest := msgCont
			assignTester(commitToTest, serverIP, serverPort)
		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
		serverMutex.Unlock()
	}
	conn.Close()
}

func assignTester(commitToTest, serverIP, serverPort string) {
	commitAssigned := false
	for _, tester := range testersList {
		tokens := strings.Split(tester, util.Colon)
		testerIP, testerPort := tokens[0], tokens[1]
		resp := util.SendAndReceiveData(testerIP, testerPort, util.HelloMsg)
		if resp == util.HelloMsg {
			commitTestersMap[commitToTest] = tester
			util.SendData(serverIP, serverPort, util.OkMsg)
			commitAssigned = true
			break
		}
	}
	if commitAssigned == false {
		commitsToTest = append(commitsToTest, commitToTest)
		util.SendData(serverIP, serverPort, util.FailMsg)
	}
}

func watchTesters() {
	for {
		for index := 0; index < len(testersList); index++ {
			tokens := strings.Split(testersList[index], util.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			resp := util.SendAndReceiveData(testerIP, testerPort, util.StatMsg)
			if resp != util.OkMsg {
				serverMutex.Lock()
				log.Printf("Removing tester running at %s:%s\n", testerIP, testerPort)
				// Remove the tester
				failedCommit := getMapKeyFromValue(testersList[index])
				commitsToTest = append(commitsToTest, failedCommit)
				testersList = append(testersList[:index], testersList[index+1:]...)
				serverMutex.Unlock()
			}
		}
		time.Sleep(util.WaitInterval)
	}
}

func getMapKeyFromValue(value string) string {
	for k, v := range commitTestersMap {
		if v == value {
			return k
		}
	}
	log.Fatalf("Can't find value %s in %s\n", value, commitTestersMap)
	return ""
}

func recoverFailedTests(serverIP, serverPort string) {
	for {
		for _, recoverCommit := range commitsToTest {
			serverMutex.Lock()
			assignTester(recoverCommit, serverIP, serverPort)
			serverMutex.Unlock()
		}
		time.Sleep(util.WaitInterval)
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
	go recoverFailedTests(*serverIPPtr, *serverPortPtr)
}
