package integration

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/onsi/gomega/gbytes"
)

type TcpSyslogServer struct {
	Addr   string
	Buffer *gbytes.Buffer
}

func (s *TcpSyslogServer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + s.Addr)

	close(ready)

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				return
			}

			_, err = io.Copy(s.Buffer, conn)
			if err != nil {
				panic(err)
			}

			conn.Close()
		}
	}()

	<-signals

	return nil
}
