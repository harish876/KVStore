package replication

import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
)

func SendRdbMessage(conn net.Conn, glb args.RedisArgs) {
	base64String := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	response, _ := base64.StdEncoding.DecodeString(base64String)
	stringResponse := string(response)
	sentBytes, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(stringResponse), stringResponse)))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Printf("Sent Byte count of RDB message %d\n", sentBytes)
}

func ReplicateWrite(glb args.RedisArgs) {
	for msg := range glb.ReplicationChannel {
		// fmt.Printf("Message Recieved from Channel %s", msg)
		for _, rConn := range glb.ReplicationConfig.Replicas {
			fmt.Println("The Replica url is ", rConn.Conn.LocalAddr().String())
			sentBytes, err := rConn.Conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
			}
			fmt.Printf("Sent Byte count of SET command %d", sentBytes)
		}
	}
}
