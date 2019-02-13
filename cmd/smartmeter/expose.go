package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"
	"sync"
	"time"
)

var exposeOptions = struct {
	Addr string
}{}

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "send all raw datagram packets over a TCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := OpenPort()
		if err != nil {
			return fmt.Errorf("failed to open port: %v", err)
		}
		defer port.Close()

		l, err := net.Listen("tcp", exposeOptions.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen on address %v: %v", exposeOptions.Addr, err)
		}
		defer l.Close()

		connections := make(map[net.Conn]struct{}, 100)
		connMutex := sync.Mutex{}

		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					log.Println(fmt.Errorf("failed to listen on address %v: %v", exposeOptions.Addr, err))
					continue
				}

				connMutex.Lock()
				connections[conn] = struct{}{}
				connMutex.Unlock()
			}
		}()

		buf := make([]byte, 32 * 1024)
		for {
			nr, err := port.Read(buf)
			if nr > 0 {
				connMutex.Lock()
				for conn := range connections {
					if err := conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
						conn.Close()
						delete(connections, conn)
					}
					if _, err := conn.Write(buf[:nr]); err != nil {
						conn.Close()
						delete(connections, conn)
					}
				}
				connMutex.Unlock()
			}
			if err != nil {
				break
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)
	exposeCmd.Flags().StringVar(&exposeOptions.Addr, "addr", ":8888", "TCP server address")
}
