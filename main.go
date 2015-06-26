package main

import (
	"io"
	"log"
	"net"
)

const (
	prefix = "/tmp/"
)

func main() {
	l, err := net.Listen("tcp", ":8080")
	defer l.Close()

	for {
		if conn, err := l.Accept(); err == nil {
			go handleConn(conn)
		}
	}
	log.Fatal(err)
}

// handleConn takes a network connection,
// and translates it's output to a socket file
func handleConn(c net.Conn) {
	defer c.Close()
	address, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		log.Fatal(err)
	}

	socket, err := net.Listen("unix", prefix+address)
	defer socket.Close()

	if err != nil {
		log.Fatal(err)
	}

	if socketPipe, err := socket.Accept(); err == nil {
		_, err = io.Copy(socketPipe, c)
		socketPipe.Close()
		log.Println(err)
	} else {
		log.Fatal("Uncaught error on socker listen")
	}
}
