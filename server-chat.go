/*	server-chat.go

	TCP chat server that accepts any connections on the 
	configured listening port.

	- main goroutine listens for and accepts tcp 
	connections on the configured port.

	- broadcaster goroutine broadcasts messages to each 
	connected client.

	- handler goroutine handles client connections, 
	consuming client input and constructing messages.
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

// map of client address to connection
var clients = make(map[string]net.Conn)

// channel for signaling client disconnet
var disconnect = make(chan message)

// channel for passing messages
var messages = make(chan message)

// A message is a representation of the content being
// shared between clients.
type message struct {
	text	string
	address	string
}

func main(){
	// attempt to listen to a pre-defined local port
	listen, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handler(conn)
	}
}

// Goroutine that handles client connections.
// Consumes client input and constructs messages.
func handler(conn net.Conn){
	clients[conn.RemoteAddr().String()] = conn
	messages <- newMessage(" joined", conn)
	
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- newMessage(": "+input.Text(), conn)
	}

	delete(clients, conn.RemoteAddr().String())

	disconnect <- newMessage(" disconnected.", conn)

	conn.Close()
}

// Goroutine that broadcasts all messages to all clients.
// Sends client input messages and disconnection messages.
func broadcaster(){
	for {
		select {
		case msg := <-messages:
			for _, conn := range clients {
				if msg.address == conn.RemoteAddr().String() {
					continue
				}
				fmt.Fprintln(conn, msg.text)
			}
		case msg := <-disconnect:
			for _, conn := range clients {
				fmt.Fprintln(conn, msg.text)
			}
		}
	}
}

// Constructor for a message struct
func newMessage(msg string, conn net.Conn) message {
	addr := conn.RemoteAddr().String()
	return message {
		text:		addr + msg,
		address:	addr,
	}
}