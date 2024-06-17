package server

import (
	"errors"
	"net"
	"sync"
)

type Replica struct {
	Conn   net.Conn
	Offset int
}

func FromConn(conn net.Conn) *Replica {
	return &Replica{
		Conn:   conn,
		Offset: 0,
	}
}

type ConnectionPool struct {
	Replicas []*Replica
	mutex    sync.Mutex
}

func NewConnectionPool() ConnectionPool {
	return ConnectionPool{
		Replicas: make([]*Replica, 0),
	}
}

func (cp *ConnectionPool) Add(replica *Replica) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.Replicas = append(cp.Replicas, replica)
}

// Function to get a connection from the pool
func (cp *ConnectionPool) Get() (*Replica, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	if len(cp.Replicas) == 0 {
		return nil, errors.New("connection pool is empty")
	}
	conn := cp.Replicas[0]
	cp.Replicas = cp.Replicas[1:]
	return conn, nil
}

// Function to return a connection to the pool
func (cp *ConnectionPool) Put(replica *Replica) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.Replicas = append(cp.Replicas, replica)
}
