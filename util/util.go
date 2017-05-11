package util

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
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

// MsgDel - string to mark end of data
const MsgDel = "'\n'"

// MsgDelByte - Byte indicating end of data
const MsgDelByte = '\n'

// Colon - string to denote Colon
const Colon = ":"

// ResMsg - string denoting request for result of test
const ResMsg = "RESULT"

// Dash - string to seperate message type from content
const Dash = "-"

// EmptyStr - string to denote empty string
const EmptyStr = ""

// WaitInterval - integer representing interval to wait before
// doing the same operation again
const WaitInterval = 5

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
	server := ip + Colon + port
	conn, err := net.Dial(Protocol, server)
	if err != nil {
		log.Println(err)
		return FailMsg
	}
	fmt.Fprintf(conn, data+MsgDel)
	resp, err := bufio.NewReader(conn).ReadString(MsgDelByte)
	if err != nil {
		log.Println(err)
		return FailMsg
	}
	return resp
}

// SendData : Function to send given data
// on the given ip and port.
func SendData(ip, port, data string) {
	server := ip + Colon + port
	conn, err := net.Dial(Protocol, server)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(conn, data+MsgDel)
}

// RunOrFail : Function to run a command and return
// the output or fail trying
func RunOrFail(cmd string, arg []string) string {
	out, err := exec.Command(cmd, arg...).Output()

	if err != nil {
		log.Fatal(err)
	}

	return string(out)
}
