package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func ConnectToMaster(glb args.RedisArgs) (net.Conn, error) {
	if glb.Role == args.MASTER_ROLE && glb.ServerPort != glb.MasterPort {
		fmt.Printf("This is the master itself. No need to connect")
		return nil, nil
	}
	fmt.Println("Degugging...")
	master := fmt.Sprintf("0.0.0.0:%d", glb.MasterPort)
	conn, err := net.Dial("tcp", master)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func PingMaster(conn net.Conn, glb args.RedisArgs) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"PING"})))
	if err != nil {
		return fmt.Errorf("error sending PING message to master: %v", err)
	}
	// buffer := make([]byte, 1024)
	// recievedBytes, err := conn.Read(buffer)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("Recieved Bytes in request: %d\n", recievedBytes)
	// parsedMessage, _ := parser.Decode(buffer[:recievedBytes])
	// fmt.Println("Response to Ping Command", parsedMessage)
	return nil
}

func SendReplConfMessage(conn net.Conn, glb args.RedisArgs) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", glb.ServerPort)})))
	if err != nil {
		return fmt.Errorf("error sending REPLCONF listening port message to master: %v", err)
	}
	_, err = conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "capa", "psync2"})))
	if err != nil {
		return fmt.Errorf("error sending REPLCONF capa psync2 message to master: %v", err)
	}
	return nil
}
