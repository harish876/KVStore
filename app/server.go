package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/pkg/server"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
)

func main() {
	store := store.New()
	s := server.NewServer()

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.ServerPort))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", s.ServerPort)
		os.Exit(1)
	}
	if s.Role == server.SLAVE_ROLE {
		var wg sync.WaitGroup
		wg.Add(1)
		mConn, err := s.HandleHandShakeWithMaster(&wg)
		if err != nil {
			fmt.Printf("Failed to Connect to master: %v", err)

		}
		wg.Wait()
		_ = mConn
		if mConn != nil {
			fmt.Println("Replica connected to master!...")
			go s.HandleClient(mConn, store)
		}
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("Connected to: ", conn.LocalAddr().String())
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}
		go s.HandleClient(conn, store)
	}
	os.Exit(1)
}
