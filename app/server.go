package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/codecrafters-io/redis-starter-go/pkg/replication"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
)

type Server struct {
	Store *store.Store
	Args  *args.RedisArgs
}

func main() {
	glbArgs := args.ParseArgs()
	s := Server{
		Store: store.New(),
		Args:  &glbArgs,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", glbArgs.MasterPort))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", glbArgs.MasterPort)
		os.Exit(1)
	}
	if glbArgs.Role == args.SLAVE_ROLE {
		var wg sync.WaitGroup
		wg.Add(1)
		mConn, err := replication.HandleHandShakeWithMaster(&wg, glbArgs)
		if err != nil {
			log.Fatalf("Error at Handle Hand Shake with master %v", err)
		}
		wg.Wait()
		go s.handleClient(mConn)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}
		go s.handleClient(conn)
	}
	close(glbArgs.ReplicationChannel)
	os.Exit(1)
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	go replication.ReplicateWrite(s.Args)
	for {
		buffer := make([]byte, 1024)
		recievedBytes, err := conn.Read(buffer)
		if err == io.EOF || recievedBytes == 0 {
			break
		}
		fmt.Printf("Recieved Bytes in request: %d\n", recievedBytes)
		request := string(buffer[:recievedBytes])
		parsedMessage, _ := parser.Decode(buffer[:recievedBytes])

		var response string
		switch parsedMessage.Method {
		case "ping":
			response = parser.EncodeSimpleString("PONG")
			fmt.Printf("Response is %s ", response)

		case "echo":
			response = parser.EncodeRespString(parsedMessage.Messages)
			fmt.Printf("Response is %s ", response)

		case "set":
			fmt.Printf("SET Message for %s %q\n", s.Args.Role, parsedMessage.Messages)
			if parsedMessage.MessagesLength == 2 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				s.Store.Set(key, value)
				response = parser.EncodeSimpleString("OK")

			} else if parsedMessage.MessagesLength == 4 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				ttl, _ := strconv.Atoi(parsedMessage.Messages[3])
				s.Store.SetWithTTL(key, value, ttl)
				response = parser.EncodeSimpleString("OK")

			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("\nResponse is %s \n", response)

		case "get":
			if parsedMessage.MessagesLength >= 1 {
				key := parsedMessage.Messages[0]
				if value, ok := s.Store.Get(key); !ok {
					response = parser.BULK_NULL_STRING
				} else {
					response = parser.EncodeRespString([]string{value})
				}
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response is %s ", response)

		case "info":
			if parsedMessage.MessagesLength >= 1 && parsedMessage.Messages[0] == "replication" {
				var infoParams []string
				if s.Args.Role == args.MASTER_ROLE {
					/* Test Case issue when building a single string */
					infoParams = append(infoParams, fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", s.Args.Role, s.Args.ReplicationConfig.ReplicationId, s.Args.ReplicationConfig.ReplicationOffset))
				} else {
					infoParams = append(infoParams, parser.GetLablelledMessage("role", args.SLAVE_ROLE))
				}
				response = parser.EncodeRespString(infoParams)
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response is %s ", []byte(response))

		case "replconf":
			if s.Args.Role == args.MASTER_ROLE {
				if parsedMessage.MessagesLength == 2 && parsedMessage.Messages[0] == "listening-port" {
					lport, err := strconv.Atoi(parsedMessage.Messages[1])
					if err == nil {
						fmt.Println("Incoming Replica Connection is", fmt.Sprintf("0.0.0.0:%d", lport))
						s.Args.ReplicationConfig.Replicas = append(s.Args.ReplicationConfig.Replicas, args.Replicas{Conn: conn})
					}
				}
				response = parser.EncodeSimpleString("OK")
			} else {
				response = ""
			}
			fmt.Printf("Response for replconf is %s ", []byte(response))

		case "psync":
			if s.Args.Role == args.MASTER_ROLE {
				response = parser.EncodeSimpleString(fmt.Sprintf("FULLRESYNC %s %d", s.Args.ReplicationConfig.ReplicationId, s.Args.ReplicationConfig.ReplicationOffset))
			} else {
				response = ""
			}
			fmt.Printf("Response for psync is %s ", []byte(response))

		default:
			fmt.Printf("Buffer: %s\n", buffer[:recievedBytes])
			fmt.Printf("Parsed Message: %s\n", parsedMessage.Messages)
		}

		sentBytes, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			break
		}
		if s.Args.Role == args.MASTER_ROLE && parsedMessage.Method == "psync" {
			replication.SendRdbMessage(conn, s.Args)
		}
		if s.Args.Role == args.MASTER_ROLE && parsedMessage.Method == "set" {
			s.Args.ReplicationChannel <- request
		}
		fmt.Printf("Number of Bytes sent : %d\n", sentBytes)
	}
}
