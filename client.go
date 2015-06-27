package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
)

const (
	obsServer = "localhost:8090"
)

type command struct {
	Cmd string
}

type devNuller struct{}

func (d *devNuller) Write(b []byte) (int, error) {
	return len(b), nil
}

type feed struct {
	from  io.Reader
	to    io.Writer
	newTo chan io.Writer
}

func (f *feed) SendTo(w io.Writer) {
	f.newTo <- w
}

func (f *feed) Go() {
	buff := make([]byte, 1024*10)
	go func() {
		for {
			select {
			case newTo := <-f.newTo:
				// send ack to server
				// switch writer
			default:
				n, err := f.from.Read(buff)
				if err != nil {
					log.Fatal(err)
				}
				m := 0
				for m < n {
					written, err := f.to.Write(buff[m:n])
					if err != nil {
						log.Fatal(err)
					}
					m += written
				}
			}
		}
	}()
}

func main() {
	obsListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := obsListener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Need to devnull
		obsConn, err := net.Dial("tcp", obsServer)
		if err != nil {
			log.Fatal(err)
		}

		signalDecoder := gob.NewDecoder(obsConn)
		for {
			var cmd command
			if err := signalDecoder.Decode(&cmd); err != nil {
				log.Fatal(err)
			}

			switch cmd.Cmd {
			case "start":
				// feed.SendTo(conn)
			case "stop":
				// feed.SendTo(nuller)
			default:
				log.Println("Unknown Command: %v", cmd)
			}
		}

	}
}
