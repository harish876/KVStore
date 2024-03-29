package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

var RedisMap map[string]string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	RedisMap = make(map[string]string)
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
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
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
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
			response = "+PONG\r\n"
			fmt.Printf("Response is %s ", response)
		case "echo":
			if parsedMessage.MessagesLength > 0 {
				response = fmt.Sprintf("$%d\r\n%s\r\n", len(parsedMessage.Messages[0]), parsedMessage.Messages[0])
			} else {
				response = "$-1\r\n"
			}
			fmt.Printf("Response is %s ", response)

		case "set":
			if parsedMessage.MessagesLength >= 2 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				RedisMap[key] = value
				response = "+OK\r\n"
			} else {
				response = "$-1\r\n"
			}
			fmt.Printf("Response is %s ", response)
		case "get":
			if parsedMessage.MessagesLength >= 1 {
				key := parsedMessage.Messages[0]
				if value, ok := RedisMap[key]; !ok {
					response = "$-1\r\n"
				} else {
					response = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				}
			} else {
				response = ""
			}
			fmt.Printf("Response is %s ", response)
		}

		sentBytes, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			break
		}
		fmt.Printf("Number of Bytes sent : %d\n", sentBytes)
	}
}
