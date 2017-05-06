package communicator

import (
	"bufio"
	"fmt"
	"log"
	"net"
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
const MsgDel = '\n'

// Colon - string to denote Colon
const Colon = ":"

// ResMsg - string denoting request for result of test
const ResMsg = "RESULT"

// Dash - string to seperate message type from content
const Dash = "-"

// SendAndReceiveData : Function to send given data
// on the given ip and port. Returns the response
func SendAndReceiveData(ip, port, data string) string {
	server = ip + Colon + port
	conn, err := net.Dial(Protocol, server)
	if err != nil {
		log.Println(err)
		return FailMsg
	}
	fmt.Fprintf(conn, data+MsgDel)
	resp, err := bufio.NewReader(conn).ReadString(MsgDel)
	if err != nil {
		log.Println(err)
		return FailMsg
	}
	return resp
}

// SendData : Function to send given data
// on the given ip and port.
func SendData(ip, port, data string) {
	server = ip + Colon + port
	conn, err := net.Dial(Protocol, server)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintf(conn, data+MsgDel)
}
