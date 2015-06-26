package main

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	prefix = "/tmp/"
)

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		if conn, err := l.Accept(); err == nil {
			go handleConn(conn)
		}
	}
	log.Fatal(err)
}

type nuller struct {
	pause  chan struct{}
	kill   chan struct{}
	resume chan struct{}
	c      net.Conn
}

func (n *nuller) Pause() {
	<-n.pause
}

func (n *nuller) Kill() {
	<-n.kill
}

func (n *nuller) Resume() {
	<-n.resume
}

func (n *nuller) Go() {
	nullBuff := make([]byte, 1024*1024)
	go func() {
		defer close(n.kill)
		defer close(n.pause)
		defer close(n.resume)

	readLoop:
		for {
			select {
			case n.pause <- struct{}{}:
				select {
				case n.resume <- struct{}{}:
					continue readLoop
				case n.kill <- struct{}{}:
					return
				}
			case n.kill <- struct{}{}:
				return
			default:
				// read off c with timeout
				err := n.c.SetReadDeadline(time.Now().Add(time.Second * 1))
				if err != nil {
					log.Printf("Error setting deadline to 1 second: %v", err)
				}
				num, err := n.c.Read(nullBuff)
				log.Printf("Num bytes read: %d", num)
				if err != nil {
					if err == io.EOF {
						log.Println("EOF encountered when nulling socket")
						close(n.kill)
						close(n.pause)
						close(n.resume)
						return
					}
					log.Printf("Error nulling connection: %v", err)
				}
				err = n.c.SetReadDeadline(time.Time{})
				if err != nil {
					log.Printf("Error setting deadline back to 0: %v", err)
				}
			}
		}
	}()
}

func devNull(c net.Conn) *nuller {
	n := &nuller{
		pause:  make(chan struct{}),
		resume: make(chan struct{}),
		kill:   make(chan struct{}),
		c:      c,
	}

	n.Go()
	return n
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
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	myNuller := devNull(c)
	defer myNuller.Kill()

	if socketPipe, err := socket.Accept(); err == nil {
		myNuller.Pause()
		_, err = io.Copy(socketPipe, c)
		myNuller.Resume()
		socketPipe.Close()
		log.Println(err)
	} else {
		log.Fatal("Uncaught error on socker listen")
	}
}
