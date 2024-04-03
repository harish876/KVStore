package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/codecrafters-io/redis-starter-go/pkg/replication"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
)

func main() {
	store := store.New()
	glbArgs := args.ParseArgs()
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", glbArgs.ServerPort))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", glbArgs.ServerPort)
		os.Exit(1)
	}
	if glbArgs.Role == args.SLAVE_ROLE {
		var wg sync.WaitGroup
		wg.Add(1)
		mConn, err := replication.HandleHandShakeWithMaster(&wg, glbArgs)
		if err != nil {
			fmt.Printf("FAILED TO CONNECT TO MASTER: %v", err)

		}
		wg.Wait()
		if mConn != nil {
			fmt.Println("Replica connected to master!...")
			go handleClient(mConn, store, glbArgs)
		}
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("Connected to: ", conn.LocalAddr().String())
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}
		go handleClient(conn, store, glbArgs)
	}
	close(glbArgs.ReplicationChannel)
	os.Exit(1)
}

func handleClient(conn net.Conn, s *store.Store, glb *args.RedisArgs) {
	defer conn.Close()
	for {
		buffer := make([]byte, 1024)
		recievedBytes, err := conn.Read(buffer)
		if err == io.EOF || recievedBytes == 0 {
			fmt.Printf("Received IOF Error for %s\n", glb.Role)
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
			fmt.Printf("SET Message for %s %q %s\n", glb.Role, parsedMessage.Messages, request)
			if parsedMessage.MessagesLength == 2 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				s.Set(key, value)
				fmt.Println("Here")
				s.PrintMap()
				response = parser.EncodeSimpleString("OK")
			} else if parsedMessage.MessagesLength == 4 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				ttl, _ := strconv.Atoi(parsedMessage.Messages[3])
				s.SetWithTTL(key, value, ttl)
				response = parser.EncodeSimpleString("OK")

			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("\nResponse is %s \n", response)

		case "get":
			fmt.Printf("GET Message for %s %q\n", glb.Role, parsedMessage.Messages)
			if parsedMessage.MessagesLength >= 1 {
				key := parsedMessage.Messages[0]
				if value, ok := s.Get(key); !ok {
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
				if glb.Role == args.MASTER_ROLE {
					infoParams = append(infoParams, fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", glb.Role, glb.ReplicationConfig.ReplicationId, glb.ReplicationConfig.ReplicationOffset))
				} else {
					infoParams = append(infoParams, parser.GetLablelledMessage("role", args.SLAVE_ROLE))
				}
				response = parser.EncodeRespString(infoParams)
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response is %s ", []byte(response))

		case "replconf":
			if glb.Role == args.MASTER_ROLE {
				if parsedMessage.MessagesLength == 2 && parsedMessage.Messages[0] == "listening-port" {
					lport, err := strconv.Atoi(parsedMessage.Messages[1])
					if err == nil {
						fmt.Println("Incoming Replica Connection is", fmt.Sprintf("0.0.0.0:%d", lport))
					}
				}
				response = parser.EncodeSimpleString("OK")
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response for replconf is %s ", []byte(response))

		case "psync":
			if glb.Role == args.MASTER_ROLE {
				response = parser.EncodeSimpleString(fmt.Sprintf("FULLRESYNC %s %d", glb.ReplicationConfig.ReplicationId, glb.ReplicationConfig.ReplicationOffset))
				// glb.ReplicationConfig.Replicas = append(glb.ReplicationConfig.Replicas, args.Replicas{Conn: conn})
				glb.ReplicationConfig.Replicas.Add(conn)
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response for psync is %s ", []byte(response))

		default:
			fmt.Printf("Buffer: %s\n", buffer[:recievedBytes])
			fmt.Printf("Parsed Message: %s\n", parsedMessage.Messages)
			response = "-ERR unknown command\r\n"
		}

		sentBytes, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			break
		}
		if glb.Role == args.MASTER_ROLE && parsedMessage.Method == "psync" {
			replication.SendRdbMessage(conn, glb)
		}
		if glb.Role == args.MASTER_ROLE && parsedMessage.Method == "set" {
			replication.PropagateMessageToReplica(request, &glb.ReplicationConfig.Replicas)
		}
		fmt.Printf("Number of Bytes sent : %d\n", sentBytes)
	}
}
