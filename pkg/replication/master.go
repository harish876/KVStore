package replication

import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
)

func SendRdbMessage(conn net.Conn, glb *args.RedisArgs) {
	fmt.Println("Remote", conn.RemoteAddr().String())
	fmt.Println("Remote", conn.LocalAddr().String())
	base64String := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	response, _ := base64.StdEncoding.DecodeString(base64String)
	stringResponse := string(response)
	sentBytes, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(stringResponse), stringResponse)))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Printf("Sent Byte count of RDB message %d\n", sentBytes)
}

func ReplicateWrite(glb *args.RedisArgs) {
	for msg := range glb.ReplicationChannel {
		fmt.Printf("Message Recieved from Channel %s", msg)
		fmt.Println(glb.ReplicationConfig.Replicas)
		for _, replicaPort := range glb.ReplicationConfig.Replicas {
			url := fmt.Sprintf("0.0.0.0:%d", replicaPort.Port)
			fmt.Println("The url is", url)
			conn, err := net.Dial("tcp", url)
			if err != nil {
				fmt.Printf("Unable to replicate message: to server with port %d. Error: %v", replicaPort.Port, err)
				continue
			}
			sentBytes, err := conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
			}
			fmt.Printf("Sent Byte count of SET command %d", sentBytes)
		}
		// conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", glb.MasterHost, glb.MasterPort))
		// if err != nil {
		// 	fmt.Printf("Unable to replicate message: to server with port %d. Error: %v", glb.MasterPort, err)
		// 	continue
		// }
		// sentBytes, err := conn.Write([]byte(msg))
		// if err != nil {
		// 	fmt.Println("Error writing response: ", err.Error())
		// }
		// fmt.Printf("Sent Byte count of SET command %d", sentBytes)
	}
}
