package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/args"
	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func ConnectToMaster(glb args.RedisArgs) (net.Conn, error) {
	master := fmt.Sprintf("%s:%d", glb.MasterHost, glb.MasterPort)
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

func HandleHandShake(glb args.RedisArgs) error {
	conn, err := ConnectToMaster(glb)
	if err != nil {
		fmt.Printf("Failed to connect to master %v", err)
		return err
	}
	// defer conn.Close()
	PingMaster(conn, glb)
	data := make([]byte, 1024)
	d, err := conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	SendReplConfMessage(conn, glb)
	res := data[:d]
	fmt.Println(string(res))
	return nil
}
