package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func ConnectToMaster(glb args.RedisArgs) error {
	if glb.Role == args.MASTER_ROLE && glb.ServerPort != glb.MasterPort {
		fmt.Printf("This is the master itself. No need to connect")
		return nil
	}
	fmt.Println("Degugging...")
	master := fmt.Sprintf("0.0.0.0:%d", glb.MasterPort)
	conn, err := net.Dial("tcp", master)
	if err != nil {
		return err
	}
	err = PingMaster(conn)
	if err != nil {
		return err
	}
	return nil
}

// Fire and Forget now, Listen to the message!
func PingMaster(conn net.Conn) error {
	defer conn.Close()

	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"PING"})))
	if err != nil {
		return fmt.Errorf("error sending PING response to master: %v", err)
	}
	return nil
}
