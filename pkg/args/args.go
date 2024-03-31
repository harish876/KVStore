package args

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ReplicationConfig struct {
	Replicas          ConnectionPool
	ReplicationId     string
	ReplicationOffset int
	// ReplicaLock       sync.Mutex
}

type RedisArgs struct {
	ServerPort         int
	MasterHost         string
	MasterPort         int
	Role               string
	ReplicationConfig  *ReplicationConfig
	ReplicationChannel chan string
}

var (
	DEFAULT_PORT = 6379
	DEFAULT_HOST = "localhost"
	MASTER_ROLE  = "master"
	SLAVE_ROLE   = "slave"
)

func ParseArgs() *RedisArgs {
	currentPortPtr := flag.Int("port", DEFAULT_PORT, "Current Redis Server Port")
	masterServerDetailsPtr := flag.String("replicaof", "localhost 6379", "Current Redis Server Port")
	flag.Parse()
	port := *currentPortPtr
	masterServerDetails := *masterServerDetailsPtr
	var masterDetails = strings.Split(masterServerDetails, " ")

	var masterHost string
	var masterPort int
	var role string

	if len(masterDetails) < 2 {
		masterHost = DEFAULT_HOST
		masterPort = DEFAULT_PORT
	} else {
		parsedMasterPort, err := strconv.Atoi(masterDetails[1])
		if err != nil {
			masterPort = DEFAULT_PORT
		} else {
			masterPort = parsedMasterPort
		}
		masterHost = masterDetails[0]
	}

	var replicationConfig *ReplicationConfig
	var replicationChan chan string
	if port == masterPort {
		role = MASTER_ROLE
		replicationChan = make(chan string)
		replicationConfig = &ReplicationConfig{
			Replicas:          NewConnectionPool(),
			ReplicationId:     GenerateReplicationId(),
			ReplicationOffset: 0,
		}
	} else {
		role = SLAVE_ROLE
	}

	return &RedisArgs{
		ServerPort:         port,
		MasterPort:         masterPort,
		MasterHost:         masterHost,
		Role:               role,
		ReplicationConfig:  replicationConfig,
		ReplicationChannel: replicationChan,
	}
}

func GenerateReplicationId() string {
	timestamp := time.Now().Unix()
	machineID, err := os.Hostname()
	if err != nil {
		machineID = DEFAULT_HOST
	}
	data := fmt.Sprintf("%d%s", timestamp, machineID)
	hash := sha512.Sum512([]byte(data))
	hashHex := hex.EncodeToString(hash[:])
	truncatedHash := hashHex[:40]

	return truncatedHash
}

type ConnectionPool struct {
	Replicas []net.Conn
	mutex    sync.Mutex
}

func NewConnectionPool() ConnectionPool {
	return ConnectionPool{
		Replicas: make([]net.Conn, 0),
	}
}

func (cp *ConnectionPool) Add(conn net.Conn) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.Replicas = append(cp.Replicas, conn)
}

// Function to get a connection from the pool
func (cp *ConnectionPool) Get() (net.Conn, error) {
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
func (cp *ConnectionPool) Put(conn net.Conn) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.Replicas = append(cp.Replicas, conn)
}
