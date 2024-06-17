package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func (s *Server) SendRdbMessage(conn net.Conn) {
	base64String := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	response, _ := base64.StdEncoding.DecodeString(base64String)
	stringResponse := string(response)
	sentBytes, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(stringResponse), stringResponse)))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Printf("Sent Byte count of RDB message %d\n", sentBytes)
}

// gets both the raw and parser message
func (s *Server) PropagateMessageToReplica(request string, parsedMessage parser.RESPMessage) {
	successfulWrites := 0
	s.ReplicaLock.Lock()
	defer s.ReplicaLock.Unlock()
	for {
		log.Println("Propagate Message Request", request)
		replicaConn, err := s.ReplicaPool.Get()
		if err != nil {
			fmt.Println("Error getting connection from pool:", err)
			break
		}
		log.Printf("Propagating Message... %s to server %s ", request, replicaConn.LocalAddr().String())
		_, err = replicaConn.Write([]byte(request))
		if err != nil {
			fmt.Println("Error writing to replica:", err)
			s.ReplicaPool.Put(replicaConn)
			break
		}
		successfulWrites++
		s.ReplicaPool.Put(replicaConn)
		if successfulWrites == len(s.ReplicaPool.Replicas) {
			break
		}
	}
}

// gets only the parsed message
func (s *Server) PropagateMessageToReplicaV1(messages []string) {
	successfulWrites := 0
	s.ReplicaLock.Lock()
	defer s.ReplicaLock.Unlock()
	request := parser.EncodeRespArray(messages)
	for {
		log.Println("Propagate Parsed message", request)
		replicaConn, err := s.ReplicaPool.Get()
		if err != nil {
			fmt.Println("Error getting connection from pool:", err)
			break
		}
		log.Printf("Propagating Message... %s to server %s ", request, replicaConn.LocalAddr().String())
		_, err = replicaConn.Write([]byte(request))
		if err != nil {
			fmt.Println("Error writing to replica:", err)
			s.ReplicaPool.Put(replicaConn)
			break
		}
		successfulWrites++
		s.ReplicaPool.Put(replicaConn)
		if successfulWrites == len(s.ReplicaPool.Replicas) {
			break
		}
	}
}
