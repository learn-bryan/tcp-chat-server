/*	server-chat.go

	TCP chat server that accepts any connections on the 
	configured listening port.

	- main goroutine listens for and accepts tcp 
	connections on the configured port.

	- broadcaster goroutine broadcasts messages to each 
	connected client.

	- clientHandler goroutine handles client connections, 
	consuming client input and constructing messages.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

// CLI flags
var flagIP = flag.String("ip", "127.0.0.1", "Listening IP address")
var flagPort = flag.String("port", "9000", "Listening port number")

// A client is the representation of all the data needed
// for managing the client and handling messages.
type client struct {
	Conn	net.Conn
	Name	string
	Message	chan string
}

// map of client address to client object
var clients = make(map[string]*client)

// channel for signaling client disconnect
var disconnect = make(chan string)

func main(){
	flag.Parse()

	// attempt to listen to provided IP and Port
	listen, err := net.Listen("tcp", *flagIP+":"+*flagPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening on: ", listen.Addr())

	// server listen/accept loop
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go clientHandler(conn)
	}
}

// Goroutine that handles client connections.
// Consumes client input and constructs messages.
func clientHandler(conn net.Conn){
	log.Printf("clientHandler: connection from: %v", conn.RemoteAddr())

	name, err := readInput(conn, "Enter name: ")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: send welcome message

	// init client object
	client := &client{
			Message:	make(chan string),
			Conn:		conn,
			Name:		name,
	}

	// start the send and receive goroutines
	go client.send()
	go client.receive()

	clients[conn.RemoteAddr().String()] = client
	//messages <- newMessage(" joined", conn)
	
	//input := bufio.NewScanner(conn)
	//for input.Scan() {
	//	messages <- newMessage(": "+input.Text(), conn)
	//}

	//delete(clients, conn.RemoteAddr().String())

	//disconnect <- newMessage(" disconnected.", conn)

	//conn.Close()
}

func (c *client) receive() {
	for {
		select {
		case msg := <-c.Message:
			writeMessage(c.Conn, msg)
		case msg := <-disconnect:
			writeMessage(c.Conn, msg+" disconnected.")
		}
	}
}

func (c *client) send() {
Loop:
	for {
		// read client input
		msg, err := readInput(c.Conn, "")
		if err != nil {
			log.Fatal(err)
		}

		// parse commands
		if msg == "\\quit" {
			//c.close()
			log.Printf("%v disconnected.", c.Name)
			break Loop
			//break
		}

		// broadcast message
		for _, remoteClient := range clients {
			if c.Conn.RemoteAddr().String() == remoteClient.Conn.RemoteAddr().String() {
				continue
			}
			remoteClient.Message <- c.Name+": "+msg
		}

	}

	c.close()
}

// Terminate a client connection
func (c *client) close() {
	delete(clients, c.Conn.RemoteAddr().String())
	disconnect <- c.Name
	c.Conn.Close()
}

func readInput(conn net.Conn, qst string) (string, error) {
	conn.Write([]byte(qst))
	s, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("readinput: could not read input from stdin: %v from client %v", err, conn.RemoteAddr().String())
		return "", err
	}
	s = strings.Trim(s, "\r\n")
	return s, nil
}

func writeMessage(conn net.Conn, msg string) {
	fmt.Fprintln(conn, msg+"\r")
}