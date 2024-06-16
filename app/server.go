package main

import (
	"fmt"
	"net"
	"os"

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
		if err := s.HandleHandShakeWithMaster(store); err != nil {
			fmt.Printf("Failed to Connect to master: %v", err)
		}
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}
		go s.HandleClient(conn, store)
	}
	os.Exit(1)
}
