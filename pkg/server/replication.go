package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/pkg/parser"
	"github.com/codecrafters-io/redis-starter-go/pkg/store"
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
func (s *Server) SendReplConfListeningMessage(conn net.Conn) error {
	_, err := conn.Write([]byte(parser.EncodeRespArray([]string{"REPLCONF", "listening-port", fmt.Sprintf("%d", s.ServerPort)})))
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
func (s *Server) HandleHandShakeWithMaster(store *store.Store) error {
	masterConn, err := s.ConnectToMaster()
	if err != nil {
		fmt.Printf("Failed to connect to master %v", err)
		return err
	}
	reader := bufio.NewReader(masterConn)
	s.PingMaster(masterConn)
	//Ping command Read
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	_ = data

	s.SendReplConfListeningMessage(masterConn)
	//Replf Conf Listening Port Message Read
	data, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	_ = data

	s.SendReplConfCapaMessage(masterConn)
	//Replf Conf Psync Message Read
	data, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	_ = data

	s.SendPsyncMessage(masterConn)
	//Replf Conf Psync Message Read
	data, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	_ = data

	//RDB Read
	data, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	_ = data
	/*
		Todo: Read RDB response here later
	*/

	// this should be a persisten connection I guess here
	go s.HandleClient(masterConn, store)
	return nil
}

// func handlerPropagation(reader *bufio.Reader, masterConn net.Conn) {
// 	defer masterConn.Close()

// 	for {

// 	}
// }
