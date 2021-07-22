package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"./alpha9wordcounter"
)

const (
	connHost = "localhost"
	connPort = "9999"
	connType = "tcp"
)

func main() {

	alpha9wordcounter.CountWords()
	// Listen for incoming connections.
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + connHost + ":" + connPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	n := bytes.Index(buf, []byte{0})
	message := string(buf[:n-1])

	var result string

	command := strings.Split(message, " ")
	if command[0] == "search" {
		result = alpha9wordcounter.Search(command[1])
	} else if command[0] == "common" {
		n, e := strconv.Atoi(command[1])
		if e == nil {
			result = alpha9wordcounter.Common(n)
		} else {
			result = "A number is needed after common command\n"
		}
	} else {
		result = "Invalid command\n"
	}

	// Send a response back to person contacting us.
	// conn.Write([]byte("Message received."))
	conn.Write([]byte(result))

	// Close the connection when you're done with it.
	conn.Close()
}
