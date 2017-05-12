package util

import (
	"bufio"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// FailMsg - string returned on error
const FailMsg = "FAIL"

// HelloMsg - string used to say hello to the server
const HelloMsg = "HELLO"

// RegMsg - string used to register as a tester
const RegMsg = "REGISTER"

// TestMsg - string used to request tests for a commit
const TestMsg = "TEST"

// StatMsg - string used to get status of the server
const StatMsg = "STATUS"

// OkMsg - string used to denote healthy server
const OkMsg = "OKAY"

// Protocol - string denoting protocol to use for communication
const Protocol = "tcp"

// MsgDelByte - Byte indicating end of data
const MsgDelByte = '|'

// MsgDel - string to mark end of data
const MsgDel = string(MsgDelByte)

// Colon - string to denote Colon
const Colon = ":"

// ResMsg - string denoting request for result of test
const ResMsg = "RESULT"

// Dash - string to seperate message type from content
const Dash = "-"

// EmptyStr - string to denote empty string
const EmptyStr = ""

// WaitInterval - integer representing interval to wait before
// doing the same operation again in seconds
const WaitInterval = 5 * time.Second

// Version control system specific commands

// GitExecutable - command to run git
const GitExecutable = "git"

// GitHashSwitch - command to get last git commit hash
var GitHashSwitch = []string{"rev-parse", "HEAD"}

// GitPullSwitch - command to fetch and apply latest git commit
var GitPullSwitch = []string{"pull"}

// GitCleanSwitch - command to clean the repository
var GitCleanSwitch = []string{"clean", "-d", "-f", "-x"}

// GitResetToCommitSwitch - command to reset the repository to given commit
var GitResetToCommitSwitch = []string{"reset", "--hard"}

// Testing command

// GoExecutable - command to run go
const GoExecutable = "go"

// GoTestSwitch - command to run all go tests for a project
var GoTestSwitch = []string{"test", "./..."}

// SendAndReceiveData : Function to send given data
// on the given ip and port. Returns the response
func SendAndReceiveData(ip, port, data string) string {
	conn := SendData(ip, port, data)
	if conn == nil {
		log.Println("Cannot receive data without sending")
		return FailMsg
	}
	resp, respErr := bufio.NewReader(conn).ReadBytes(MsgDelByte)
	if respErr != nil {
		log.Println(respErr)
		return FailMsg
	}
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	return FormatResp(resp)
}

// SendData : Function to send given data
// on the given ip and port. Returns connection
func SendData(ip, port, data string) net.Conn {
	server := ip + Colon + port
	log.Printf("Send %s to %s \n", data, server)
	conn, connErr := net.Dial(Protocol, server)
	if connErr != nil {
		log.Println(connErr)
		return nil
	}
	_, sendErr := conn.Write([]byte(data + MsgDel))
	if sendErr != nil {
		log.Println(sendErr)
		return nil
	}
	return conn
}

// RunOrFail : Function to run a command and return
// the output or fail trying
func RunOrFail(cmd string, arg []string) string {
	out, err := exec.Command(cmd, arg...).Output()

	if err != nil {
		log.Fatal(err)
	}
	ans := string(out)
	return strings.TrimSpace(ans)
}

// FormatResp : Returns formatted string representation
// of the data after removing MsgDelByte
func FormatResp(resp []byte) string {
	resp = resp[:len(resp)-1]
	return string(resp)
}

// GetRandomPortStr : Returns string denoting random port
// number in range of allowed port numbers
func GetRandomPortStr() string {
	addr, err := net.ResolveTCPAddr(Protocol, "localhost:0")
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenTCP(Protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port
	return strconv.FormatUint(uint64(port), 10)
}
