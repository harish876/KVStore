package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
)

func main() {
	store := store.New()
	portPtr := flag.Int("port", 6379, "Redis Server Port")
	flag.Parse()
	port := *portPtr
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleClient(conn, store)
	}
}

func handleClient(conn net.Conn, s *store.Store) {
	defer conn.Close()
	for {
		buffer := make([]byte, 1024)
		recievedBytes, err := conn.Read(buffer)
		if err == io.EOF || recievedBytes == 0 {
			break
		}
		fmt.Printf("Recieved Bytes in request: %d\n", recievedBytes)
		parsedMessage, _ := parser.Decode(buffer[:recievedBytes])
		var response string
		switch parsedMessage.Method {
		case "ping":
			response = parser.Encode([]string{"PONG"}, true)
			fmt.Printf("Response is %s ", response)
		case "echo":
			response = parser.Encode(parsedMessage.Messages, false)
			fmt.Printf("Response is %s ", response)

		case "set":
			if parsedMessage.MessagesLength == 2 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				s.Set(key, value)
				response = parser.Encode([]string{"OK"}, true)
			} else if parsedMessage.MessagesLength == 4 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				ttl, _ := strconv.Atoi(parsedMessage.Messages[3])
				s.SetWithTTL(key, value, ttl)
				response = parser.Encode([]string{"OK"}, true)
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response is %s ", response)

		case "get":
			if parsedMessage.MessagesLength >= 1 {
				key := parsedMessage.Messages[0]
				if value, ok := s.Get(key); !ok {
					response = parser.BULK_NULL_STRING
				} else {
					response = parser.Encode([]string{value}, false)
				}
			} else {
				response = ""
			}
			fmt.Printf("Response is %s ", response)

		default:
			fmt.Printf("Buffer: %s\n", buffer[:recievedBytes])
			fmt.Printf("Parsed Message: %s\n", parsedMessage.Messages)
		}

		sentBytes, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			break
		}
		fmt.Printf("Number of Bytes sent : %d\n", sentBytes)
	}
}
