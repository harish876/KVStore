package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
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
func (s *Server) PropagateMessageToReplica(request string) {
	successfulWrites := 0
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
