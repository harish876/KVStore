package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
)

func (s *Server) ConnectToMaster() (net.Conn, error) {
	master := fmt.Sprintf("%s:%d", s.MasterHost, s.MasterPort)
	conn, err := net.Dial("tcp", master)
	if err != nil {
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
func (s *Server) SendReplConfMessage(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", s.ServerPort)})))
	if err != nil {
		return fmt.Errorf("error sending REPLCONF listening port message to master: %v", err)
	}
	_, err = conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "capa", "psync2"})))
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
func (s *Server) HandleHandShakeWithMaster(wg *sync.WaitGroup) (net.Conn, error) {
	conn, err := s.ConnectToMaster()
	if err != nil {
		fmt.Printf("Failed to connect to master %v", err)
		return nil, err
	}
	s.PingMaster(conn)
	//Ping command Read
	data := make([]byte, 1024)
	d, err := conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	res := data[:d]
	_ = res

	s.SendReplConfMessage(conn)
	//Replf Conf Listening Port Message Read
	data = make([]byte, 1024)
	d, err = conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	res = data[:d]
	_ = res

	//Replf Conf Psync Message Read
	data = make([]byte, 1024)
	d, err = conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	res = data[:d]
	_ = res

	s.SendPsyncMessage(conn)
	//Replf Conf Psync Message Read
	data = make([]byte, 1024)
	d, err = conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	res = data[:d]
	_ = res

	//RDB Read
	data = make([]byte, 1024)
	d, err = conn.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	res = data[:d]
	_ = res

	wg.Done()

	return conn, nil
}
