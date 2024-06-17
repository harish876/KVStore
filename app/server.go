package main

import (
	"github.com/codecrafters-io/redis-starter-go/pkg/server"
)

func main() {
	s := server.NewServer()
	s.Start()
}
