package server

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
	"github.com/codecrafters-io/redis-starter-go/pkg/utils"
)

var (
	DEFAULT_PORT = 6379
	DEFAULT_HOST = "localhost"
	MASTER_ROLE  = "master"
	SLAVE_ROLE   = "slave"
)

type Server struct {
	ServerPort        int
	MasterHost        string
	MasterPort        int
	Role              string
	ReplicaPool       ConnectionPool
	ReplicationId     string
	ReplicationOffset int
	ReplicaLock       sync.Mutex
}

func NewServer() *Server {
	currentPortPtr := flag.Int("port", DEFAULT_PORT, "Current Redis Server Port")
	masterServerDetailsPtr := flag.String("replicaof", "localhost 6379", "Current Redis Server Port")
	flag.Parse()
	port := *currentPortPtr
	masterServerDetails := *masterServerDetailsPtr
	var masterDetails = strings.Split(masterServerDetails, " ")

	var masterHost string
	var masterPort int

	if len(masterDetails) < 2 {
		masterHost = DEFAULT_HOST
		masterPort = DEFAULT_PORT
	} else {
		parsedMasterPort, err := strconv.Atoi(masterDetails[1])
		if err != nil {
			masterPort = DEFAULT_PORT
		} else {
			masterPort = parsedMasterPort
		}
		masterHost = masterDetails[0]
	}

	var server *Server = &Server{
		ServerPort:  port,
		MasterPort:  masterPort,
		MasterHost:  masterHost,
		ReplicaLock: sync.Mutex{},
	}
	if port == masterPort {
		server.Role = MASTER_ROLE
		server.ReplicaPool = NewConnectionPool()
		server.ReplicationId = utils.GenerateReplicationId(DEFAULT_HOST)
		server.ReplicationOffset = 0
	} else {
		server.Role = SLAVE_ROLE
	}

	return server
}

func (s *Server) HandleClient(conn net.Conn, st *store.Store) {
	defer conn.Close()
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
			fmt.Printf("SET Message for %s %q %s\n", s.Role, parsedMessage.Messages, request)
			if parsedMessage.MessagesLength == 2 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				st.Set(key, value)
				st.PrintMap()
				response = parser.EncodeSimpleString("OK")
			} else if parsedMessage.MessagesLength == 4 {
				key := parsedMessage.Messages[0]
				value := parsedMessage.Messages[1]
				ttl, _ := strconv.Atoi(parsedMessage.Messages[3])
				st.SetWithTTL(key, value, ttl)
				response = parser.EncodeSimpleString("OK")

			} else {
				//debug parsing issue from the master for replication
				//SET Message for slave ["bar" "456" "SET" "baz" "789"] *3
				// response = parser.BULK_NULL_STRING //temp check
				log.Println("Debug Here...")
				if len(parsedMessage.Messages) == 5 {
					m := parsedMessage.Messages
					st.Set(m[0], m[1])
					st.Set(m[3], m[4])
				}
				// response = parser.BULK_NULL_STRING

				response = parser.EncodeSimpleString("OK")
			}
			fmt.Printf("\nResponse is %s \n", response)

		case "get":
			fmt.Printf("GET Message for %s %q\n", s.Role, parsedMessage.Messages)
			if parsedMessage.MessagesLength >= 1 {
				key := parsedMessage.Messages[0]
				if value, ok := st.Get(key); !ok {
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
				if s.Role == MASTER_ROLE {
					infoParams = append(infoParams, fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", s.Role, s.ReplicationId, s.ReplicationOffset))
				} else {
					infoParams = append(infoParams, parser.GetLablelledMessage("role", SLAVE_ROLE))
				}
				response = parser.EncodeRespString(infoParams)
			} else {
				response = parser.BULK_NULL_STRING
			}
			fmt.Printf("Response is %s ", []byte(response))

		case "replconf":
			if s.Role == MASTER_ROLE {
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
			if s.Role == MASTER_ROLE {
				response = parser.EncodeSimpleString(fmt.Sprintf("FULLRESYNC %s %d", s.ReplicationId, s.ReplicationOffset))
				s.ReplicaLock.Lock()
				s.ReplicaPool.Add(conn)
				s.ReplicaLock.Unlock()
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
		if s.Role == MASTER_ROLE && parsedMessage.Method == "psync" {
			s.SendRdbMessage(conn)
		}
		if s.Role == MASTER_ROLE && parsedMessage.Method == "set" {
			//something is wrong here
			s.PropagateMessageToReplica(request, parsedMessage)
		}
		fmt.Printf("Number of Bytes sent : %d\n", sentBytes)
	}
}
