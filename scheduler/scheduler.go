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
var serverAlive = make(chan bool)

func listen(serverIP, serverPort string) {
	server := serverIP + util.Colon + serverPort
	listner, err := net.Listen(util.Protocol, server)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s:%s \n", serverIP, serverPort)
	// Watch for testers failing
	go watchTesters()

	// Try to assign commits on failed testers to
	// new testers
	go recoverFailedTests()

	for {
		conn, err := listner.Accept()
		if err != nil {
			listner.Close()
			serverAlive <- false
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
		} else {
			log.Println("Told tester that I am alive")
		}
	} else {
		serverMutex.Lock()
		log.Println(msg)
		msgCont := msg[1]
		if contMsg && header == util.RegMsg {
			testerInfo := msgCont
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
			commitToTest := msgCont
			tester := getIdleTester()
			if tester == util.EmptyStr {
				commitsToTest = append(commitsToTest, commitToTest)
				_, err := conn.Write([]byte(util.FailMsg + util.MsgDel))
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Told watcher that I cannot test for %s right now\n", commitToTest)
				}
			} else {
				commitTestersMap[commitToTest] = tester
				_, err := conn.Write([]byte(util.OkMsg + util.MsgDel))
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Told watcher that I have assigned test for %s to %s\n", commitToTest, tester)
				}
			}

		} else {
			log.Printf("Unknown Request %s\n", resp)
		}
		serverMutex.Unlock()
	}
	conn.Close()
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
		alive := <-serverAlive
		if alive == false {
			return
		}
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

func recoverFailedTests() {
	for {
		alive := <-serverAlive
		if alive == false {
			return
		}
		for _, recoverCommit := range commitsToTest {
			serverMutex.Lock()
			tester := getIdleTester()
			if tester != util.EmptyStr {
				commitTestersMap[recoverCommit] = tester
				log.Printf("Assigned failed %s to %s", recoverCommit, tester)
			}
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
	listen(*serverIPPtr, *serverPortPtr)
}
