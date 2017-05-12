package main

import (
	"bufio"
	"cisgo/util"
	"flag"
	"fmt"
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
	log.Printf("Listening on %s:%s \n", serverIP, serverPort)

	for {
		conn, err := listner.Accept()
		if err != nil {
			listner.Close()
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
		}
	} else {
		serverMutex.Lock()
		if contMsg && header == util.RegMsg {
			testerInfo := msg[1]
			_, err := conn.Write([]byte(util.OkMsg + util.MsgDel))
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Added %s tester to the scheduler", testerInfo)
				testersList = append(testersList, testerInfo)
			}
		} else if contMsg && header == util.ResMsg {
			resForCommit := msg[1]
			result := msg[2]
			writeResultToFile(resForCommit, result)
			delete(commitTestersMap, resForCommit)
		} else if contMsg && header == util.TestMsg {
			commitToTest := msg[1]
			tester := getIdleTester()
			if tester == util.EmptyStr {
				sendFailMsgToWatcher(conn, commitToTest)
			} else {
				done := tryAssigningCommit(tester, commitToTest, formattedResp)
				if done == false {
					sendFailMsgToWatcher(conn, commitToTest)
				} else {
					commitTestersMap[commitToTest] = tester
					removeCommitFromPending(commitToTest)
					_, err := conn.Write([]byte(util.OkMsg + util.MsgDel))
					if err != nil {
						log.Println(err)
					} else {
						log.Printf("Assigned test for %s to %s\n", commitToTest, tester)
					}
				}
			}
		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
		serverMutex.Unlock()
	}
	conn.Close()
}

func tryAssigningCommit(tester, commitToTest, testInfo string) bool {
	tokens := strings.Split(tester, util.Colon)
	testerIP, testerPort := tokens[0], tokens[1]
	resp := util.SendAndReceiveData(testerIP, testerPort, testInfo)
	if resp == util.OkMsg {
		return true
	}
	return false
}

func sendFailMsgToWatcher(conn net.Conn, commitToTest string) {
	commitsToTest = append(commitsToTest, commitToTest)
	_, err := conn.Write([]byte(util.FailMsg + util.MsgDel))
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Cannot test for %s right now\n", commitToTest)
	}
}

func getIdleTester() string {
	for _, tester := range testersList {
		tokens := strings.Split(tester, util.Colon)
		testerIP, testerPort := tokens[0], tokens[1]
		resp := util.SendAndReceiveData(testerIP, testerPort, util.HelloMsg)
		if resp == util.HelloMsg {
			return tester
		}
	}
	return util.EmptyStr
}

func watchTesters() {
	for {
		time.Sleep(util.WaitInterval)
		log.Println("Active testers", testersList)
		for index := 0; index < len(testersList); index++ {
			tokens := strings.Split(testersList[index], util.Colon)
			testerIP, testerPort := tokens[0], tokens[1]
			resp := util.SendAndReceiveData(testerIP, testerPort, util.StatMsg)
			if resp != util.OkMsg {
				serverMutex.Lock()
				log.Printf("Removing tester running at %s:%s\n", testerIP, testerPort)
				// Remove the tester
				failedCommit := getMapKeyFromValue(testersList[index])
				if failedCommit != util.EmptyStr {
					commitsToTest = append(commitsToTest, failedCommit)
				}
				testersList = append(testersList[:index], testersList[index+1:]...)
				serverMutex.Unlock()
			}
		}
	}
}

func getMapKeyFromValue(value string) string {
	for k, v := range commitTestersMap {
		if v == value {
			return k
		}
	}
	return util.EmptyStr
}

func removeCommitFromPending(commit string) {
	for index := 0; index < len(commitsToTest); index++ {
		if commitsToTest[index] == commit {
			commitsToTest = append(commitsToTest[:index], commitsToTest[index+1:]...)
			fmt.Println(commitsToTest)
			return
		}
	}
}

func recoverFailedTests() {
	for {
		time.Sleep(util.WaitInterval)
		log.Println("Commits to test", commitsToTest)
		for _, recoverCommit := range commitsToTest {
			serverMutex.Lock()
			tester := getIdleTester()
			if tester != util.EmptyStr {
				testInfo := util.TestMsg + util.Dash + recoverCommit
				done := tryAssigningCommit(tester, recoverCommit, testInfo)
				if done {
					commitTestersMap[recoverCommit] = tester
					removeCommitFromPending(recoverCommit)
					log.Printf("Assigned failed %s to %s", recoverCommit, tester)
				}
			}
			serverMutex.Unlock()
		}
	}
}

func main() {

	// Get required information in command line from user

	// Scheduler information required to setup the server
	serverIPPtr := flag.String("sip", "localhost", "IP address of the scheduler server")
	serverPortPtr := flag.String("sport", "8080", "Port of the scheduler server")

	flag.Parse()

	log.Printf("Starting scheduler server at %s:%s\n", *serverIPPtr, *serverPortPtr)

	// Watch for testers failing
	go watchTesters()

	// Try to assign failed commit tests to new testers
	go recoverFailedTests()

	// Start watching the given repository path
	listen(*serverIPPtr, *serverPortPtr)
}
