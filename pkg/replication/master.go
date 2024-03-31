package replication

import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
)

func SendRdbMessage(conn net.Conn, glb *args.RedisArgs) {
	base64String := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	response, _ := base64.StdEncoding.DecodeString(base64String)
	stringResponse := string(response)
	sentBytes, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(stringResponse), stringResponse)))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Printf("Sent Byte count of RDB message %d\n", sentBytes)
}

func PropagateMessageToReplica(request string, r *args.ConnectionPool) {
	successfulWrites := 0

	for {
		replicaConn, err := r.Get()
		if err != nil {
			fmt.Println("Error getting connection from pool:", err)
			break // Break loop if there are no available connections
		}

		_, err = replicaConn.Write([]byte(request))
		if err != nil {
			fmt.Println("Error writing to replica:", err)
			r.Put(replicaConn) // Return the connection to the pool
			break
		}

		// Increment successful writes
		successfulWrites++

		// Return the connection to the pools
		r.Put(replicaConn)

		// Check if all replicas received the command
		if successfulWrites == len(r.Replicas) {
			break
		}
	}
}
