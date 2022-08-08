/*	client-chat.go

	TCP chat client that attempts to connect to the 
	configured IP:PORT. Client's stdin stream to sent over 
	the server connection. Copies server's output stream to
	client's stdout.
*/

package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	// attempt to connect to the chat server
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}

	// channel for signaling when disconnected
	done := make(chan struct{})

	// goroutine for printing messages from the server
	go func() {
		io.Copy(os.Stdout, conn)
		log.Println("done")
		done <- struct{}{} // signal main goroutine
	}()

	// Copies client stdin stream to server connection
	if _,err := io.Copy(conn, os.Stdin); err != nil {
		log.Fatal(err)
	}
	
	conn.Close()
	<-done // wait for background goroutine
}