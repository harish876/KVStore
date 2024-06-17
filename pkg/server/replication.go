package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func (s *Server) ConnectToMaster() (net.Conn, error) {
	master := fmt.Sprintf("%s:%d", s.Config.ReplicaofHost, s.Config.ReplicaofPort)
	conn, err := net.Dial("tcp", master)
	if err != nil {
		log.Println("Error at ConnectToMaster", err)
		return nil, err
	}
	return conn, nil
}
func (s *Server) PingMaster(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"PING"})))
	if err != nil {
		return fmt.Errorf("error sending PING message to master: %v", err)
	}
	return nil
}
func (s *Server) SendReplConfListeningMessage(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", s.Config.Port)})))
	if err != nil {
		return fmt.Errorf("error sending REPLCONF listening port message to master: %v", err)
	}
	return nil
}
func (s *Server) SendReplConfCapaMessage(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "capa", "psync2"})))
	if err != nil {
		return fmt.Errorf("error sending REPLCONF capa psync2 message to master: %v", err)
	}
	return nil
}
func (s *Server) SendPsyncMessage(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"PSYNC", "?", fmt.Sprintf("%d", -1)})))
	if err != nil {
		return fmt.Errorf("error sending PSYNC listening port message to master: %v", err)
	}
	return nil
}
func (s *Server) HandleHandShakeWithMaster() {
	masterConn, err := s.ConnectToMaster()
	if err != nil {
		fmt.Printf("Failed to connect to master %v", err)
		//return err
	}
	reader := bufio.NewReader(masterConn)
	s.PingMaster(masterConn)
	//Ping command Read
	reader.ReadString('\n')

	s.SendReplConfListeningMessage(masterConn)
	//Replf Conf Listening Port Message Read
	reader.ReadString('\n')

	s.SendReplConfCapaMessage(masterConn)
	//Replf Conf Psync Message Read
	reader.ReadString('\n')

	s.SendPsyncMessage(masterConn)
	//Replf Conf Psync Message Read
	reader.ReadString('\n')

	//RDB Read
	response, _ := reader.ReadString('\n')
	if response[0] != '$' {
		fmt.Printf("Invalid response\n")
		os.Exit(1)
	}
	rdbSize, _ := strconv.Atoi(response[1 : len(response)-2])
	buffer := make([]byte, rdbSize)
	receivedSize, err := reader.Read(buffer)
	if err != nil {
		log.Printf("Invalid RDB received %v\n", err)
		os.Exit(1)
	}
	if rdbSize != receivedSize {
		log.Printf("Size mismatch - got: %d, want: %d\n", receivedSize, rdbSize)
	}

	// this should be a persisten connection I guess here
	go s.ServeReplicas(reader, masterConn)
}

// This is where the replicas recieve messages from the master and the "replicas" respond to the master
func (s *Server) ServeReplicas(reader *bufio.Reader, masterConn net.Conn) error {
	defer masterConn.Close()
	for {
		message, recievedBytes, err := parser.DecodeV1(reader)
		log.Println("Receieved Bytes", recievedBytes)
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
		response, _ := s.HandleCommand(message)

		switch strings.ToLower(message[0]) {
		case "replconf":
			_, err := masterConn.Write([]byte(response))
			if err != nil {
				log.Println("error responding to master connection: ", err)
			}
		}
		s.ReplicationOffset += recievedBytes
	}
	return nil
}
