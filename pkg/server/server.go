package server

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

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

type ServerConfig struct {
	Port          int
	Role          string
	ReplicationId string
	ReplicaofHost string
	ReplicaofPort int
	Dir           string
	Dbfilename    string
}

type Server struct {
	Store             *store.Store
	Config            ServerConfig
	ReplicaPool       ConnectionPool
	ReplicationOffset int
	AcksRecieved      chan bool
	ReplicaLock       sync.Mutex
}

func NewServer() *Server {
	var config ServerConfig
	flag.IntVar(&config.Port, "port", DEFAULT_PORT, "listen on specified port")
	flag.StringVar(&config.ReplicaofHost, "replicaof", "", "start server in replica mode of given host and port")
	//yet to be completed
	flag.StringVar(&config.Dir, "dir", "", "directory where RDB files are stored")
	flag.StringVar(&config.Dbfilename, "dbfilename", "", "name of the RDB file")
	flag.Parse()

	log.Println(config.ReplicaofHost)

	if len(config.ReplicaofHost) == 0 {
		config.Role = MASTER_ROLE
		config.ReplicationId = utils.GenerateReplicationId(DEFAULT_HOST)
	} else {
		config.ReplicaofHost = strings.Split(config.ReplicaofHost, " ")[0]
		config.Role = SLAVE_ROLE
		switch flag.NArg() {
		case 0:
			config.ReplicaofPort = 6379
		case 1:
			config.ReplicaofPort, _ = strconv.Atoi(flag.Arg(0))
		default:
			flag.Usage()
		}
	}

	return &Server{
		Store:        store.New(),
		Config:       config,
		AcksRecieved: make(chan bool),
		ReplicaLock:  sync.Mutex{},
	}
}

func (s *Server) Start() {
	if s.Config.Role == SLAVE_ROLE {
		s.HandleHandShakeWithMaster()
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Config.Port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", s.Config.Port)
		os.Exit(1)
	}
	for clientId := 1; ; clientId++ {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}
		go s.ServeClient(clientId, conn)
	}
	os.Exit(1)
}

func (s *Server) HandleCommand(message []string) (string, bool) {
	var response string
	var resync bool

	switch strings.ToLower(message[0]) {
	case "ping":
		response = parser.EncodeSimpleString("PONG")

	case "echo":
		response = parser.EncodeRespString(message[1])

	case "wait":
		count, _ := strconv.Atoi(message[1])
		timeout, _ := strconv.Atoi(message[2])
		response = s.HandleWait(count, timeout)

	case "set":
		if len(message) == 3 {
			key, value := message[1], message[2]
			s.Store.Set(key, value)
			log.Printf("Set Key %s - Value %s: ", key, value)
		} else if len(message) == 5 && strings.ToLower(message[3]) == "px" {
			key, value, ttlString := message[1], message[2], message[4]
			ttl, _ := strconv.Atoi(ttlString)
			s.Store.SetWithTTL(key, value, ttl)
			log.Printf("Set Key %s - Value %s: ", key, value)
		}
		s.PropagateMessageToReplicaV1(message)
		response = parser.EncodeSimpleString("OK")

	case "get":
		if len(message) >= 1 {
			key := message[1]
			if value, ok := s.Store.Get(key); !ok {
				log.Println("Value not found for key: ", key)
				response = parser.BULK_NULL_STRING
			} else {
				response = parser.EncodeRespString(value)
			}
		} else {
			response = parser.BULK_NULL_STRING
		}

	case "info":
		if len(message) >= 1 && strings.ToLower(message[1]) == "replication" {
			var infoParams []string
			infoParams = append(infoParams, fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d", s.Config.Role, s.Config.ReplicationId, s.ReplicationOffset))
			response = parser.EncodeRespString(infoParams[0])
		} else {
			response = parser.BULK_NULL_STRING
		}

	case "replconf":
		subCommand := strings.ToLower(message[1])
		switch subCommand {
		case "getack":
			//master sends this to replica
			response = parser.EncodeRespArray([]string{"REPLCONF", "ACK", strconv.Itoa(s.ReplicationOffset)})
			log.Printf("Master sent me a getack request, responding with Response: %v", response)
		case "ack":
			log.Printf("Acks recieved: %v", response)
			s.AcksRecieved <- true
			response = ""
		default:
			//revisit
			response = parser.EncodeSimpleString("OK")

		}

	case "psync":
		response = parser.EncodeSimpleString(fmt.Sprintf("FULLRESYNC %s %d", s.Config.ReplicationId, s.ReplicationOffset))
		resync = true

	default:
		response = "-ERR unknown command\r\n"
	}

	return response, resync
}

func (s *Server) ServeClient(clientId int, conn net.Conn) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, recievedBytes, err := parser.DecodeV1(reader)
		if err != nil {
			if err == io.EOF || recievedBytes == 0 {
				break
			} else {
				log.Printf("Error parsing request: %v ", err.Error())
				return err
			}
		}

		if len(message) == 0 {
			break
		}

		log.Println("Parsed Message: ", message)
		response, resync := s.HandleCommand(message)
		n, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			return err
		}
		log.Printf("Sent Bytes to clientId %d - response %s\n", n, response)

		if resync {
			s.SendRdbMessage(conn)
			s.ReplicaLock.Lock()
			s.ReplicaPool.Add(FromConn(conn))
			s.ReplicaLock.Unlock()
		}
	}
	return nil
}

func (s *Server) HandleWait(count, timeout int) string {
	acks := 0
	ackCmd := parser.EncodeRespArray([]string{"REPLCONF", "GETACK", "*"})

	for idx := 0; idx < len(s.ReplicaPool.Replicas); idx++ {
		currReplica := s.ReplicaPool.Replicas[idx]
		if currReplica.Offset > 0 {
			currReplica.Conn.Write([]byte(ackCmd))
			go func(conn net.Conn) {
				log.Println("waiting response from replica", conn.RemoteAddr().String())
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					log.Println("unable to read ack response from the replica", err)
				}
				log.Println("Response from the replica", string(buffer[:n]))
				s.AcksRecieved <- true

			}(currReplica.Conn)
		} else {
			acks++
		}
	}
	timer := time.After(time.Duration(timeout) * time.Millisecond)

outer:
	for acks < count {
		select {
		case <-s.AcksRecieved:
			acks++
		case <-timer:
			log.Println("timeout exceeded during wait")
			break outer
		}
	}

	return parser.EncodeInt(acks)

}
